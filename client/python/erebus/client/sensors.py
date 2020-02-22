from . import sim_pb2
from dataclasses import dataclass  # TODO: make compatible for Py3.5
from collections import namedtuple


class Sensors:
    def __init__(self, sensorsData: sim_pb2.SensorsData):
        self.sensorsData = sensorsData
        self.timestamp = sensorsData.timestamp
        self.distanceSensors = dict()
        self.positionSensors = dict()
        self.inertialSensors = dict()
        self.cameraRecognitions = dict()
        for data in sensorsData.data:
            name = data.name
            if data.HasField('distance_sensor_data'):
                self.distanceSensors[name] = data.distance_sensor_data
            if data.HasField('position_sensor_data'):
                self.positionSensors[name] = data.position_sensor_data
            if data.HasField('inertial_sensor_data'):
                self.inertialSensors[name] = data.inertial_sensor_data
            if data.HasField('camera_recognition_data'):
                self.cameraRecognitions[name] = data.camera_recognition_data

    def getTimestamp(self):
        return self.timestamp

    def getDistanceSensorReading(self, name):
        try:
            return self.distanceSensors[name].value
        except KeyError as e:
            raise e

    def getPositionSensorReading(self, name):
        try:
            return self.positionSensors[name].value
        except KeyError as e:
            raise e

    def getInertialSensorReading(self, name):
        """
        Returns the readings from the inertial sensor in (yaw, pitch, roll)
        format in radians
        """
        try:
            sensorData = self.inertialSensors[name]
            return (sensorData.yaw, sensorData.pitch, sensorData.roll)
        except KeyError as e:
            raise e

    def getRecognitionObjects(self, cameraName):
        try:
            return [RecognitionObject(
                idNumber=raw.id,
                positionOnImage=XYPair(
                    raw.position_on_image.x, raw.position_on_image.y
                ),
                sizeOnImage=XYPair(raw.size_on_image.x, raw.size_on_image.y),
                colors=raw.colors
            ) for raw in self.cameraRecognitions[cameraName].objects]
        except KeyError as e:
            raise e


XYPair = namedtuple('XYPair', 'x y')


@dataclass  # TODO: implement manually
class RecognitionObject:
    idNumber: int
    positionOnImage: XYPair
    sizeOnImage: XYPair
    colors: [float]

    def getPBRepr(self) -> \
            sim_pb2.SensorData.CameraRecognitionData.WbCameraRecognitionObject:
        pb = sim_pb2.SensorData.CameraRecognitionData. \
                WbCameraRecognitionObject()
        pb.id = self.idNumber
        pb.position_on_image.x = self.positionOnImage.x
        pb.position_on_image.y = self.positionOnImage.y
        pb.size_on_image.x = self.sizeOnImage.x
        pb.size_on_image.y = self.sizeOnImage.y
        for color in self.colors:
            pb.colors.append(color)
        return pb
