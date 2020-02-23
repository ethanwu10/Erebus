from controller import Robot
from queue import Queue
from threading import Thread
import grpc
import sim_pb2
import wb_controller_pb2
import wb_controller_pb2_grpc


# Note: localhost doesn't always work correctly - use 127.0.0.1
BROKER_ADDRESS = '127.0.0.1:51512'

MOTORS = ['left wheel', 'right wheel']
DISTANCE_SENSORS = ['so{}'.format(i) for i in range(8)]
POSITION_SENSORS = ['left wheel sensor', 'right wheel sensor']
CAMERA_SENSORS = ['camera']
INERTIAL_SENSORS = []


def _cvtRecognitionObject(ro):
    pbRo = sim_pb2.SensorData.CameraRecognitionData.WbCameraRecognitionObject()
    pbRo.id = ro.get_id()
    pbRo.position_on_image.x = ro.get_position_on_image()[0]
    pbRo.position_on_image.y = ro.get_position_on_image()[1]
    pbRo.size_on_image.x = ro.get_size_on_image()[0]
    pbRo.size_on_image.y = ro.get_size_on_image()[1]
    for color in ro.get_colors():
        pbRo.colors.append(color)
    return pbRo


def gatherCameraRecognitionData(robot, name):
    ros = robot.getCamera(name).getRecognitionObjects()
    sd = sim_pb2.SensorData()
    sd.name = name
    for ro in ros:
        sd.camera_recognition_data.objects.append(_cvtRecognitionObject(ro))
    return sd


def gatherDistanceSensorData(robot, name):
    ds = robot.getDistanceSensor(name)
    sd = sim_pb2.SensorData()
    sd.name = name
    sd.distance_sensor_data.value = ds.getValue()
    return sd


def gatherPositionSensorData(robot, name):
    ps = robot.getPositionSensor(name)
    sd = sim_pb2.SensorData()
    sd.name = name
    sd.position_sensor_data.value = ps.getValue()
    return sd


def gatherInertialSensorData(robot, name):
    imu = robot.getInertialUnit(name)
    sd = sim_pb2.SensorData()
    sd.name = name
    rpy = imu.getRollPitchYaw()
    sd.inertial_sensor_data.roll = rpy[0]
    sd.inertial_sensor_data.pitch = rpy[1]
    sd.inertial_sensor_data.yaw = rpy[2]
    return sd


def gatherSensorsData(robot):
    sds = sim_pb2.SensorsData()
    sds.timestamp = robot.getTime()
    for camera in CAMERA_SENSORS:
        sds.data.append(gatherCameraRecognitionData(robot, camera))
    for ds in DISTANCE_SENSORS:
        sds.data.append(gatherDistanceSensorData(robot, ds))
    for ps in POSITION_SENSORS:
        sds.data.append(gatherPositionSensorData(robot, ps))
    for imu in INERTIAL_SENSORS:
        sds.data.append(gatherInertialSensorData(robot, imu))
    return sds


def applyCommands(robot, commands):
    for cmd in commands.commands:
        if cmd.HasField('motor_command'):
            motorCommand = cmd.motor_command
            robot.getMotor(cmd.name).setVelocity(motorCommand.velocity)
        if cmd.HasField('led_command'):
            ledCommand = cmd.led_command
            robot.getLED(cmd.name).set(ledCommand.value)


class WbtTicker(Thread):
    def __init__(self, robot, step, dataChannel, syncChannel, doneFunc,
                 *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.robot = robot
        self.step = step
        self.dataChannel = dataChannel
        self.syncChannel = syncChannel
        self.doneFunc = doneFunc
        self.isRunning = False

    def run(self):
        # init
        # TODO: client-configurable sensor init
        for camera in CAMERA_SENSORS:
            c = self.robot.getCamera(camera)
            c.enable(self.step)
            c.recognitionEnable(self.step)
        for distanceSensor in DISTANCE_SENSORS:
            ds = self.robot.getDistanceSensor(distanceSensor)
            ds.enable(self.step)
        for positionSensor in POSITION_SENSORS:
            ps = self.robot.getPositionSensor(positionSensor)
            ps.enable(self.step)
        for motor in MOTORS:
            m = self.robot.getMotor(motor)
            m.setPosition(float('inf'))  # Velocity control mode
            m.setVelocity(0)

        # run
        self.isRunning = True
        while self.isRunning:
            sdMsg = wb_controller_pb2.WbControllerMessage.ClientMessage()
            sdMsg.sensor_data.CopyFrom(gatherSensorsData(self.robot))
            self.dataChannel.put(sdMsg)

            if self.syncChannel is not None:
                self.syncChannel.get()
            stepVal = self.robot.step(self.step)
            if stepVal == -1:
                self.isRunning = False
                self.doneFunc()

        # deinit (go into idle state)
        for camera in CAMERA_SENSORS:
            c.recognitionDisable()
            c.disable()
        for distanceSensor in DISTANCE_SENSORS:
            ds.disable()
        for positionSensor in POSITION_SENSORS:
            ps = self.robot.getPositionSensor(positionSensor)
            ps.disable()
        for motor in MOTORS:
            m = self.robot.getMotor(motor)
            m.setVelocity(0)


class WbtIdleTicker(Thread):
    def __init__(self, robot, doneFunc, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.robot = robot
        self.doneFunc = doneFunc
        self.isRunning = False

    def run(self):
        step = int(self.robot.getBasicTimeStep())
        self.isRunning = True
        while self.isRunning:
            stepVal = self.robot.step(step)
            if stepVal == -1:
                self.isRunning = False
                self.doneFunc()


def main():
    robot = Robot()
    assert robot.getSynchronization()
    name = robot.getName()
    channel = grpc.insecure_channel(BROKER_ADDRESS)
    stub = wb_controller_pb2_grpc.WbControllerStub(channel)
    sendQueue = Queue(32)
    handshakeMsg = wb_controller_pb2.WbControllerMessage.ClientMessage()
    handshakeMsg.wb_controller_handshake.robot_name = name
    # TODO: fill robot_info
    sendQueue.put(handshakeMsg)
    timestep = float('NaN')
    isIdle = True
    ticker = None
    syncChannel = None
    try:
        call = stub.Session(iter(sendQueue.get, None), wait_for_ready=True)

        def cancel():
            nonlocal call
            nonlocal sendQueue
            call.cancel()
            sendQueue.put(None)

        for serverMsg in call:
            if timestep != timestep:
                # handshake response not received (timestep is nan)
                if serverMsg.HasField('wb_controller_handshake_response'):
                    if serverMsg.wb_controller_handshake_response \
                            .HasField('error'):
                        raise RuntimeError('Failed to handshake: {}'.format(
                            serverMsg.wb_controller_handshake_response.error
                        ))
                    print('Robot connected to broker')
                    timestep = \
                        serverMsg.wb_controller_handshake_response.ok.timestep
                    isIdle = True
                    ticker = WbtIdleTicker(robot, doneFunc=cancel)
                    ticker.start()
            else:
                if serverMsg.HasField('ping'):
                    nonce = serverMsg.ping.nonce
                    pong = \
                        wb_controller_pb2.WbControllerMessage.ClientMessage()
                    pong.pong.nonce = nonce
                    sendQueue.put(pong)
                if isIdle:
                    if serverMsg.HasField('wb_controller_bound'):
                        ticker.isRunning = False
                        ticker.join()
                        isIdle = False
                        isSync = serverMsg.wb_controller_bound.is_sync
                        syncChannel = Queue() if isSync else None
                        ticker = WbtTicker(robot, timestep,
                                           dataChannel=sendQueue,
                                           syncChannel=syncChannel,
                                           doneFunc=cancel)
                        ticker.start()
                        print('Robot bound')
                else:
                    if serverMsg.HasField('wb_controller_unbound'):
                        ticker.isRunning = False
                        if syncChannel is not None:
                            syncChannel.put(True)
                        ticker.join()
                        print('Robot unbound')
                        isIdle = True
                        ticker = WbtIdleTicker(robot, doneFunc=cancel)
                        ticker.start()
                    if serverMsg.HasField('commands'):
                        applyCommands(robot, serverMsg.commands)
                        if syncChannel is not None:
                            syncChannel.put(True)
    finally:
        if ticker is not None:
            ticker.isRunning = False
            ticker.join()


main()
