import threading
import queue
import grpc
from . import client_controller_pb2_grpc
from . import client_controller_pb2
from . import session_pb2
from . import sim_pb2
from .sensors import Sensors
from .commands import Commands
from .behavior import Behavior


class WorkerThread(threading.Thread):
    def __init__(self, behaviorClass: Behavior,
                 inQueue: queue.Queue, outQueue: queue.Queue,
                 *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.behaviorClass = behaviorClass
        self.inQueue = inQueue
        self.outQueue = outQueue

    def run(self):
        for serverMsg in iter(self.inQueue.get, None):
            if serverMsg.HasField('client_controller_handshake_response'):
                print('Handshake successful, connected.')
            if serverMsg.HasField('client_controller_bound'):
                print('robot bound')
                # TODO: maybe init when sim first transitions to running after
                # a reset?
                self.behaviorObj = self.behaviorClass()
            if serverMsg.HasField('client_controller_unbound'):
                self.behaviorObj = None
                print('robot unbound')
            if serverMsg.HasField('sim_state_change'):
                simState = serverMsg.sim_state_change
                if simState.state == sim_pb2.SimState.RESET:
                    self.behaviorObj = self.behaviorClass()
            # TODO: handle sim_state_change
            if serverMsg.HasField('ping'):
                pong = client_controller_pb2.ClientControllerMessage.\
                    ControllerMessage()
                pong.pong.nonce = serverMsg.ping.nonce
                self.outQueue.put(pong)
            if serverMsg.HasField('sensor_data'):
                cmd = Commands()
                self.behaviorObj.tick(Sensors(serverMsg.sensor_data), cmd)
                controllerMsg = client_controller_pb2.ClientControllerMessage.\
                    ControllerMessage()
                controllerMsg.commands.CopyFrom(cmd.getPBMessage())
                self.outQueue.put(controllerMsg)


class Client:
    def __init__(self, behaviorClass: Behavior, name: str):
        """
        Create an Erebus client using the provided name and behavior class
        """
        self.behaviorClass = behaviorClass
        self.name = name

    def run(self, address='127.0.0.1:51512'):
        """
        Run the client and connect to the broker server at the provided IP
        address and port
        """
        channel = grpc.insecure_channel(address)
        stub = client_controller_pb2_grpc.ClientControllerStub(channel)
        inQueue = queue.Queue()
        outQueue = queue.Queue()
        handshakeMsg = client_controller_pb2.ClientControllerMessage \
            .ControllerMessage()
        handshake = handshakeMsg.client_controller_handshake
        handshake.client_name = self.name
        outQueue.put(handshakeMsg)
        wt = WorkerThread(self.behaviorClass, inQueue, outQueue)
        wt.start()
        try:
            for serverMsg in stub.Session(iter(outQueue.get, None),
                                          wait_for_ready=True):
                inQueue.put(serverMsg)
        finally:
            inQueue.put(None)
            wt.join()
