from . import sim_pb2


# TODO: consider deduplicating commands


class Commands:
    def __init__(self):
        self.msg = sim_pb2.Commands()

    def setLED(self, name: str, state: int) -> None:
        cmd = sim_pb2.Command()
        cmd.name = name
        cmd.led_command.state = state
        self.msg.commands.append(cmd)

    def setMotor(self, name: str, velocity: float) -> None:
        cmd = sim_pb2.Command()
        cmd.name = name
        cmd.motor_command.velocity = velocity
        self.msg.commands.append(cmd)

    def getPBMessage(self) -> sim_pb2.Commands:
        return self.msg
