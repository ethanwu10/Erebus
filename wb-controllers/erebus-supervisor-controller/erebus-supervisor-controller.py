from controller import Supervisor
from threading import Thread, Condition
from os import path
import sys
import subprocess
import platform
import grpc
import sim_pb2
import types_pb2
import control_pb2_grpc


# Set to address of an already-running broker to use it instead of starting a
# new broker (e.g. '127.0.0.1:51512')
EXTERN_BROKER_ADDRESS = None

# Port for broker started by this controller (ignored if EXTERN_BROKER_ADDRESS
# is set)
BROKER_PORT = 51512

# Webots run mode to use when simulation is in "running" state
RUN_MODE = Supervisor.SIMULATION_MODE_REAL_TIME


def findBrokerExecutable():
    arch = platform.machine().lower()
    if arch == 'x86_64':
        arch = 'amd64'
    os = platform.system().lower()
    exe = 'broker_{os}_{arch}'.format(os=os, arch=arch)
    if os == 'windows':
        exe += '.exe'
    return path.join(path.dirname(path.abspath(__file__)), exe)


simStateCV = Condition()
simulationIsPaused = False
simulationShouldReset = False


def maybeUpdateSimState(supervisor):
    """
    Check if the sim state has changed from what was last set (e.g. by the user
    clicking the buttons in the Webots UI) and update the broker state
    accordingly
    """
    # FIXME: implement
    pass


class WbtTicker(Thread):
    def __init__(self, supervisor, doneFunc=None, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.supervisor = supervisor
        self.step = int(self.supervisor.getBasicTimeStep())
        self.doneFunc = doneFunc
        self.isRunning = False

    def run(self):
        global simStateCV, simulationIsPaused, simulationShouldReset
        self.isRunning = True
        while self.isRunning:
            with simStateCV:
                if simulationShouldReset:
                    simulationIsPaused = True
                    simulationShouldReset = True
                shouldTick = not simulationIsPaused
            if shouldTick:
                stepVal = self.supervisor.step(self.step)
                maybeUpdateSimState(self.supervisor)
                if stepVal == -1:
                    self.isRunning = False
            else:
                self.pauseAndWait()
        if self.doneFunc is not None:
            self.doneFunc()

    def pauseAndWait(self):
        """
        Pause simulation and wait until either the simulation is resumed or is
        terminated
        """
        global simStateCV, simulationIsPaused, simulationShouldReset
        self.supervisor.simulationSetMode(
            Supervisor.SIMULATION_MODE_PAUSE)
        with simStateCV:
            waiting = True
            while waiting:
                if simStateCV.wait_for(
                    lambda: not simulationIsPaused or simulationShouldReset,
                    timeout=0.5
                ):
                    if simulationShouldReset:
                        print('resetting')
                        self.supervisor.simulationReset()
                        self.supervisor.simulationSetMode(
                            Supervisor.SIMULATION_MODE_RUN)
                        self.supervisor.step(self.step)
                        self.supervisor.simulationSetMode(
                            Supervisor.SIMULATION_MODE_PAUSE)
                        simulationIsPaused = True
                        simulationShouldReset = False
                    if not simulationIsPaused:
                        waiting = False
                else:  # Timeout
                    stepVal = self.supervisor.step(0)  # get updates
                    maybeUpdateSimState(self.supervisor)
                    if stepVal == -1:
                        self.isRunning = False
                        waiting = False


def setSimulationState(supervisor: Supervisor, state: sim_pb2.SimState):
    global simStateCV, simulationIsPaused, simulationShouldReset
    if state.state == sim_pb2.SimState.START:
        supervisor.simulationSetMode(RUN_MODE)
        if simulationIsPaused:
            with simStateCV:
                simulationIsPaused = False
                simStateCV.notifyAll()
    if state.state == sim_pb2.SimState.STOP:
        with simStateCV:
            simulationIsPaused = True
            simStateCV.notifyAll()
    if state.state == sim_pb2.SimState.RESET:
        with simStateCV:
            simulationShouldReset = True
            simStateCV.notifyAll()


class SimStateHandler(Thread):
    def __init__(self, stub, supervisor, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.supervisor = supervisor
        self.stub = stub
        self.call = self.stub.SubscribeSimulationState(
                types_pb2.Null(),
                wait_for_ready=True
        )

    def run(self):
        setSimulationState(
            self.supervisor,
            self.stub.GetSimulationState(types_pb2.Null(), wait_for_ready=True)
        )
        for simState in self.call:
            setSimulationState(self.supervisor, simState)

    def cancel(self):
        self.call.cancel()


def main():
    supervisor = Supervisor()
    # assert supervisor.getSynchronization()
    brokerAddress = EXTERN_BROKER_ADDRESS if EXTERN_BROKER_ADDRESS is not None\
        else '127.0.0.1:{}'.format(BROKER_PORT)
    if EXTERN_BROKER_ADDRESS is None:
        brokerProcess = subprocess.Popen(
            [findBrokerExecutable(), '-port', str(BROKER_PORT)],
            stdout=sys.stdout
        )
    else:
        brokerProcess = None
    try:
        channel = grpc.insecure_channel(brokerAddress)
        stub = control_pb2_grpc.ControlStub(channel)
        simStateHandler = SimStateHandler(stub=stub, supervisor=supervisor)
        simStateHandler.start()
        ticker = WbtTicker(supervisor=supervisor)
        ticker.start()
        ticker.join()  # run forever
    finally:
        if brokerProcess is not None:
            brokerProcess.terminate()
        if ticker.isRunning:
            ticker.isRunning = False
            ticker.join()
        simStateHandler.cancel()
        simStateHandler.join()


main()
