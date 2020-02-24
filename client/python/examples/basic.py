from erebus import client


class Behavior(client.Behavior):
    TURN_SPEED = 3  # rad/s

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def tick(self, sensorData, commands):
        print('Distance reading: {}cm'.format(sensorData.getDistanceSensorReading('so3')))
        print('Encoder reading: {}rad'.format(sensorData.getPositionSensorReading('left wheel sensor')))
        if (sensorData.getTimestamp() % 2) > 1:
            commands.setMotor('left wheel', +self.TURN_SPEED)
            commands.setMotor('right wheel', -self.TURN_SPEED)
        else:
            commands.setMotor('left wheel', -self.TURN_SPEED)
            commands.setMotor('right wheel', +self.TURN_SPEED)


if __name__ == '__main__':
    client.Client(Behavior, 'ExampleController').run()
