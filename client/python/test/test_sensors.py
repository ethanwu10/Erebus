import pytest
import math

from erebus.client import Sensors, XYPair, RecognitionObject
from erebus.client import sim_pb2


def testRecognitionObjectPBRepr():
    ro = RecognitionObject(
        idNumber=1,
        positionOnImage=XYPair(10, 20),
        sizeOnImage=XYPair(60, 80),
        colors=[1, 2, 3]
    )
    pb = ro.getPBRepr()
    assert pb.id == 1
    assert pb.position_on_image.x == 10
    assert pb.position_on_image.y == 20
    assert pb.size_on_image.x == 60
    assert pb.size_on_image.y == 80
    assert pb.colors == [1, 2, 3]


@pytest.fixture
def omnibusData():
    sensorsData = sim_pb2.SensorsData()
    sensorsData.timestamp = 0.1
    sensorsDataElems = dict()
    for name, value in [
            ('ds1', 1.2),
            ('ds2', 2.1),
            ('ds3', 0.1),
            ('ds4', 0.0)
    ]:
        sensorsDataElems[name] = sim_pb2.SensorData()
        sensorsDataElems[name].name = name
        sensorsDataElems[name].distance_sensor_data.value = value
        sensorsData.data.append(sensorsDataElems[name])
    for name, value in [
            ('ps1', 0),
            ('ps2', 1.2)
    ]:
        sensorsDataElems[name] = sim_pb2.SensorData()
        sensorsDataElems[name].name = name
        sensorsDataElems[name].position_sensor_data.value = value
        sensorsData.data.append(sensorsDataElems[name])
    for name, value in [
            ('imu1', (0, 0, 0)),
            ('imu2', (math.pi/2, 0, math.pi/4))
    ]:
        sensorsDataElems[name] = sim_pb2.SensorData()
        sensorsDataElems[name].name = name
        sensorsDataElems[name].inertial_sensor_data.yaw = value[0]
        sensorsDataElems[name].inertial_sensor_data.pitch = value[1]
        sensorsDataElems[name].inertial_sensor_data.roll = value[2]
        sensorsData.data.append(sensorsDataElems[name])
    for name, value in [
            ('cam1', [RecognitionObject(
                idNumber=1,
                positionOnImage=XYPair(10, 20),
                sizeOnImage=XYPair(60, 80),
                colors=[1, 2, 3]
            ), RecognitionObject(
                idNumber=2,
                positionOnImage=XYPair(30, 35),
                sizeOnImage=XYPair(10, 12),
                colors=[2, 3]
            )])
    ]:
        sde = sim_pb2.SensorData()
        sde.name = name
        for ro in value:
            sde.camera_recognition_data.objects.append(ro.getPBRepr())
        sensorsDataElems[name] = value
        sensorsData.data.append(sde)
    sensors = Sensors(sensorsData)
    return (sensors, sensorsData.timestamp, sensorsDataElems)


def testTimestamp(omnibusData):
    sensors, timestamp, sde = omnibusData
    assert timestamp == sensors.getTimestamp()


def testDistanceSensors(omnibusData):
    sensors, timestamp, sde = omnibusData
    for name in ['ds1', 'ds2', 'ds3', 'ds4']:
        assert sde[name].distance_sensor_data.value == \
            sensors.getDistanceSensorReading(name)


def testNonexistantDistanceSensor(omnibusData):
    sensors, timestamp, sde = omnibusData
    with pytest.raises(KeyError):
        sensors.getDistanceSensorReading('nonexistant')


def testPositionSensors(omnibusData):
    sensors, timestamp, sde = omnibusData
    for name in ['ps1', 'ps2']:
        assert sde[name].position_sensor_data.value == \
            sensors.getPositionSensorReading(name)


def testNonexistantPositionSensor(omnibusData):
    sensors, timestamp, sde = omnibusData
    with pytest.raises(KeyError):
        sensors.getPositionSensorReading('nonexistant')


def testInertialSensors(omnibusData):
    sensors, timestamp, sde = omnibusData
    for name in ['imu1', 'imu2']:
        imuData = sde[name].inertial_sensor_data
        ypr = (imuData.yaw, imuData.pitch, imuData.roll)
        assert ypr == sensors.getInertialSensorReading(name)


def testNonexistantInertialSensor(omnibusData):
    sensors, timestamp, sde = omnibusData
    with pytest.raises(KeyError):
        sensors.getInertialSensorReading('nonexistant')


def testCameraRecognition(omnibusData):
    sensors, timestamp, sde = omnibusData
    for name in ['cam1']:
        ros = sde[name]
        for ro in sensors.getRecognitionObjects(name):
            assert [x for x in ros if x.idNumber == ro.idNumber][0] == ro


def testNonexistantCameraRecognition(omnibusData):
    sensors, timestamp, sde = omnibusData
    with pytest.raises(KeyError):
        sensors.getRecognitionObjects('nonexistant')
