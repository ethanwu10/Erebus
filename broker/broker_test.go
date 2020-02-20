package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

const closeTimeout = 10 * time.Millisecond

func init() {
	if log == nil {
		log = logrus.New()
	}
}

type BrokerSuite struct {
	suite.Suite
	globalCtx      context.Context
	globalCtxClose context.CancelFunc
	broker         *Broker
}

func (suite *BrokerSuite) SetupTest() {
	suite.globalCtx, suite.globalCtxClose = context.WithCancel(context.Background())
	suite.broker = NewBroker(suite.globalCtx, SimInfo{timestep: 32})
}

func (suite *BrokerSuite) TestRegisterDuplicateRobot() {
	robotEnclCtx, robotEnclCtxClose := context.WithCancel(context.Background())
	suite.Require().NotNil(suite.broker.RegisterRobot("robot", robotEnclCtx))
	suite.Nil(suite.broker.RegisterRobot("robot", robotEnclCtx))
	robotEnclCtxClose()
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestUnregisterNonexistantRobot() {
	suite.Error(suite.broker.UnregisterRobot("nonexistant"))
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestUnregisterRobot() {
	robotEnclCtx, robotEnclCtxClose := context.WithCancel(context.Background())
	handle := suite.broker.RegisterRobot("robot", robotEnclCtx)
	suite.Require().NotNil(handle)
	suite.broker.UnregisterRobot("robot")
	<-handle.ctx.Done()
	time.Sleep(closeTimeout)
	robots := suite.broker.GetRobotNames()
	suite.NotContainsf(robots, "robot", "Robot was not removed from robots list")
	robotEnclCtxClose()
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestRobotAutoUnregister() {
	robotEnclCtx, robotEnclCtxClose := context.WithCancel(context.Background())
	suite.Require().NotNil(suite.broker.RegisterRobot("robot", robotEnclCtx))
	robotEnclCtxClose()
	time.Sleep(closeTimeout)
	robots := suite.broker.GetRobotNames()
	suite.NotContainsf(robots, "robot", "Robot was not removed from robots list")
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestRegisterDuplicateClient() {
	clientEnclCtx, clientEnclCtxClose := context.WithCancel(context.Background())
	suite.Require().NotNil(suite.broker.RegisterClient("client", clientEnclCtx, false))
	suite.Nil(suite.broker.RegisterClient("client", clientEnclCtx, false))
	clientEnclCtxClose()
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestUnregisterNonexistantClient() {
	suite.Error(suite.broker.UnregisterClient("nonexistant"))
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestUnregisterClient() {
	clientEnclCtx, clientEnclCtxClose := context.WithCancel(context.Background())
	handle := suite.broker.RegisterClient("client", clientEnclCtx, false)
	suite.Require().NotNil(handle)
	suite.broker.UnregisterClient("client")
	<-handle.ctx.Done()
	time.Sleep(closeTimeout)
	clients := suite.broker.GetClientNames()
	suite.NotContainsf(clients, "client", "Client was not removed from clients list")
	clientEnclCtxClose()
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestClientAutoUnregister() {
	clientEnclCtx, clientEnclCtxClose := context.WithCancel(context.Background())
	suite.Require().NotNil(suite.broker.RegisterClient("client", clientEnclCtx, false))
	clientEnclCtxClose()
	time.Sleep(closeTimeout)
	clients := suite.broker.GetClientNames()
	suite.NotContainsf(clients, "client", "Client was not removed from clients list")
	suite.globalCtxClose()
}

func (suite *BrokerSuite) TestSimStateListener() {
	listenerCtx, listenerCtxClose := context.WithCancel(context.Background())
	listener := suite.broker.GetSimStateListener(listenerCtx)
	suite.Require().NotNil(listener)
	state := pb.SimState{State: pb.SimState_START}
	suite.broker.SetSimState(state)
	suite.Equal(state.GetState(), (<-listener).GetState())
	listenerCtxClose()
	suite.globalCtxClose()
}

func TestBrokerSuite(t *testing.T) {
	suite.Run(t, new(BrokerSuite))
}
