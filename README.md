# Erebus

[![Travis CI](https://img.shields.io/travis/com/ethanwu10/erebus?style=flat-square)](https://travis-ci.com/ethanwu10/erebus)
[![CodeCov](https://img.shields.io/codecov/c/gh/ethanwu10/erebus?style=flat-square)](https://codecov.io/gh/ethanwu10/erebus)

Erebus is a rescue simulation competition environment designed for
semi-experienced to highly experienced programmers.

![Erebus environment screenshot](https://github.com/Shadow149/Erebus/raw/Orion/images/environment.JPG)

# Getting started

## Requirements

- Python 3.6+
- [Webots](https://cyberbotics.com)

Currently installable automated builds are not yet available; all components
must be [built](#building) before running. The Webots controllers have
additional runtime Python dependencies; these can be installed by their
respective `requirements.txt` files:

```sh
# from repository root, after builds are performed
$ pip -r game/controllers/erebus-robot-controller/requirements.txt \
      -r game/controllers/erebus-supervisor-controller/requirements.txt
```

Once these dependencies are installed, launch Webots and open
`game/worlds/GeneratedWorld.wbt`, and press the "Run the simulation in
real-time" button located on the toolbar next to the clock if the world has not
started yet (if the console is empty except for a line stating `INFO:
MainSupervisor: Starting controller: python -u "MainSupervisor.py"`, then it has
not started yet). The simulation should pause immediately.

Ignore the controls in the left-hand panel; they are currently not-implemented,
and the starting, stopping, and resetting of the simulation, as well as the
connecting of competitors' controllers, is handled by the CLI.

## CLI

The broker control CLI is currently the primary way of controlling the
simulation, which includes starting, stopping, and resetting the world, and
connecting a competitor's controller to a virtual robot.

Its source is located at `broker-control-cli/`, and running `make` will produce a
binary `broker-control-cli`. Run `broker-control-cli help` to learn how to use
it (better documentation coming soon).

## Writing a controller

### Python

First, install the client library (see the section on [building it](#python-2)).

Docs are available in the [client library's
directory](https://github.com/ethanwu10/erebus/blob/master/client/python/README.rst),
and an example controller demonstrating reading sensors and controlling motors
is also available in the [examples
directory](https://github.com/ethanwu10/erebus/tree/master/client/python/examples)

# Hacking

## Additional Requirements

- Go 1.13+
- [Poetry](https://python-poetry.org)
- [GRPC tools](https://grpc.io/docs/quickstart/go/#prerequisites)
- [`gox`](https://github.com/mitchellh/gox)

Assuming `$GOPATH/bin` is in your `PATH`, the GRPC tools and gox can be
installed like so:

```sh
$ GO111MODULE=off go get -u google.golang.org/grpc github.com/golang/protobuf/protoc-gen-go github.com/mitchellh/gox
```

## Architecture

Erebus runs in a client-server architecture, with the broker as the master
server that everything else connects to. The broker is responsible for relaying
data between a virtual robot and a client controller (what a competitor writes -
referred to hereafter as simply a "client"), and also handles starting,
stopping, and resetting the environment (although currently it just forwards
this to all connected services).

Each virutal robot runs a Webots controller which connects to the broker and
awaits a peer client to be connected to it. Likewise, each client connects to
the broker and awaits a robot controller to be connected to it. The broker
provides an interface to connect pairs of Webots robot controllers and client
controllers together by their names, at which point the controllers exchange
messages for sensor data and commands.

The supervisor controller (`wb-controllers/erebus-supervisor-controller`) is
responsible for making sure that the state of the Webots simulation matches the
state in the broker by starting, pausing, and resetting the simulation when
appropriate.

The broker control CLI issues commands to the broker, and allows a game
administrator to list connected clients and robots, connect / disconnect robots
to clients, and set the state of the simulation.

## Building

### GRPC / Protobuf

A top-level Makefile exists with a convenience target `proto` to generate code
for all components depending on protobuf:

```sh
$ make proto
```

This is necessary for developing all Python components (the generated code is
excluded from Git), and you should always run this after editing the proto
definitions (in `shared/proto`) to ensure that generated code is up-to-date. Be
sure to commit changes to generated Go sources (`*/gen/*.pb.go`)

### Broker

The broker (`broker/`) is a go-modules-enabled project, and includes a Makefile
for generating protobuf files (`make proto`). After proto sources are generated,
it can be built using the standard `go build`, and a convenience target is also
set up in the Makefile so you can build using `make`.

### Broker control CLI

The broker control CLI (`broker-control-cli/`) is set up the same way as the
main broker itself. Use `make` to build.

### Clients

#### Python

The Python client library (`client/python`) is managed using
[Poetry](https://python-poetry.org) and a Makefile for handling GRPC / Protobuf
generated code (see [GRPC / Protobuf](#grpc-%2F-protobuf)). To build an
installable library, first run `make`, then run `poetry build`; the resulting
packages can be found in `client/python/dist` and installed using `pip install`.

### Webots Controllers

The Webots controllers are located at `wb-controllers` and are symlinked into
the Webots project directory (`game/controllers`) as needed. As they depend on
GRPC / Protobuf generated code, each controller has a Makefile to generate this
code. The controllers' sources are *not* managed by Poetry, thus the GRPC Python
tools must be available in the Python path when the Makefile is run. These
dependencies are listed in the `requirements-dev.txt` files of each controller.

The `erebus-supervisor-controller` has an additional build step (invoked by
`make build`) which performs a cross-build for the [broker](#broker), and copies
the resulting executables into the build output along-side the Python code for
the controller.
