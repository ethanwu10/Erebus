import abc
from .sensors import Sensors
from .commands import Commands


class Behavior(abc.ABC):

    @abc.abstractmethod
    def tick(self, sensorData: Sensors, commands: Commands):
        pass
