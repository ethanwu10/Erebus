package main

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

const closeTimeout = 10 * time.Millisecond

func init() {
	log = logrus.New()
}

func TestUnregisterRobot(t *testing.T) {
	globalCtx, globalCtxClose := context.WithCancel(context.Background())
	broker := NewBroker(globalCtx, SimInfo{timestep: 32})
	robotEnclCtx, robotEnclCtxClose := context.WithCancel(context.Background())
	handle := broker.RegisterRobot("robot", robotEnclCtx)
	broker.UnregisterRobot("robot")
	<-handle.ctx.Done()
	time.Sleep(closeTimeout)
	robots := broker.GetRobotNames()
	for _, robot := range robots {
		if robot == "robot" {
			t.Error("Robot was not removed from robots list")
		}
	}
	robotEnclCtxClose()
	globalCtxClose()
}

func TestRobotAutoUnregister(t *testing.T) {
	globalCtx, globalCtxClose := context.WithCancel(context.Background())
	broker := NewBroker(globalCtx, SimInfo{timestep: 32})
	robotEnclCtx, robotEnclCtxClose := context.WithCancel(context.Background())
	broker.RegisterRobot("robot", robotEnclCtx)
	robotEnclCtxClose()
	time.Sleep(closeTimeout)
	robots := broker.GetRobotNames()
	for _, robot := range robots {
		if robot == "robot" {
			t.Error("Robot was not removed from robots list")
		}
	}
	globalCtxClose()
}

func TestUnregisterClient(t *testing.T) {
	globalCtx, globalCtxClose := context.WithCancel(context.Background())
	broker := NewBroker(globalCtx, SimInfo{timestep: 32})
	clientEnclCtx, clientEnclCtxClose := context.WithCancel(context.Background())
	handle := broker.RegisterClient("client", clientEnclCtx, false)
	broker.UnregisterClient("client")
	<-handle.ctx.Done()
	time.Sleep(closeTimeout)
	clients := broker.GetClientNames()
	for _, client := range clients {
		if client == "client" {
			t.Error("Client was not removed from clients list")
		}
	}
	clientEnclCtxClose()
	globalCtxClose()
}

func TestClientAutoUnregister(t *testing.T) {
	globalCtx, globalCtxClose := context.WithCancel(context.Background())
	broker := NewBroker(globalCtx, SimInfo{timestep: 32})
	clientEnclCtx, clientEnclCtxClose := context.WithCancel(context.Background())
	broker.RegisterClient("client", clientEnclCtx, false)
	clientEnclCtxClose()
	time.Sleep(closeTimeout)
	clients := broker.GetClientNames()
	for _, client := range clients {
		if client == "client" {
			t.Error("Client was not removed from clients list")
		}
	}
	globalCtxClose()
}

func TestSimStateListener(t *testing.T) {
	globalCtx, globalCtxClose := context.WithCancel(context.Background())
	broker := NewBroker(globalCtx, SimInfo{timestep: 32})
	listenerCtx, listenerCtxClose := context.WithCancel(context.Background())
	listener := broker.GetSimStateListener(listenerCtx)
	state := pb.SimState{State: pb.SimState_START}
	broker.SetSimState(state)
	if got := <-listener; got.GetState() != state.GetState() {
		t.Errorf("Didn't get expected state: received %s", got)
	}
	listenerCtxClose()
	globalCtxClose()
}
