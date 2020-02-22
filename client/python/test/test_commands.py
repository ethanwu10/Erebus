import pytest
from erebus.client import Commands
from erebus.client import sim_pb2


def extractNamedCommand(commands, name):
    for cmd in commands.commands[::-1]:
        if cmd.name == name:
            return cmd
    return None


def testMotorBasic():
    cmd = Commands()
    cmd.setMotor('m1', 1.2)
    cmd.setMotor('m2', 2.1)
    pb_cmds = cmd.getPBMessage()
    assert extractNamedCommand(pb_cmds, 'm1').HasField('motor_command')
    assert extractNamedCommand(pb_cmds, 'm1').motor_command.velocity == 1.2
    assert extractNamedCommand(pb_cmds, 'm2').HasField('motor_command')
    assert extractNamedCommand(pb_cmds, 'm2').motor_command.velocity == 2.1


def testLEDBasic():
    cmd = Commands()
    cmd.setLED('l1', 0)
    cmd.setLED('l2', 1024)
    pb_cmds = cmd.getPBMessage()
    assert extractNamedCommand(pb_cmds, 'l1').HasField('led_command')
    assert extractNamedCommand(pb_cmds, 'l1').led_command.state == 0
    assert extractNamedCommand(pb_cmds, 'l2').HasField('led_command')
    assert extractNamedCommand(pb_cmds, 'l2').led_command.state == 1024


def testCombined():
    cmd = Commands()
    cmd.setMotor('m1', 1.2)
    cmd.setLED('l1', 0)
    cmd.setMotor('m2', 2.1)
    cmd.setLED('l2', 1024)
    pb_cmds = cmd.getPBMessage()
    assert extractNamedCommand(pb_cmds, 'm1').HasField('motor_command')
    assert extractNamedCommand(pb_cmds, 'm1').motor_command.velocity == 1.2
    assert extractNamedCommand(pb_cmds, 'm2').HasField('motor_command')
    assert extractNamedCommand(pb_cmds, 'm2').motor_command.velocity == 2.1
    assert extractNamedCommand(pb_cmds, 'l1').HasField('led_command')
    assert extractNamedCommand(pb_cmds, 'l1').led_command.state == 0
    assert extractNamedCommand(pb_cmds, 'l2').HasField('led_command')
    assert extractNamedCommand(pb_cmds, 'l2').led_command.state == 1024
