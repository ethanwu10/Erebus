package main

import (
	"context"
	"errors"
	"io"

	log "github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

var _ = log.WithFields // FIXME

type ControlServer struct {
	pb.UnimplementedControlServer

	broker *Broker
}

func NewControlServer(broker *Broker) *ControlServer {
	return &ControlServer{broker: broker}
}

func (s *ControlServer) GetRobots(context.Context, *pb.Null) (*pb.ControlMessage_GetRobotsResponse, error) {
	return &pb.ControlMessage_GetRobotsResponse{RobotNames: s.broker.GetRobotNames()}, nil
}

func (s *ControlServer) GetClientControllers(context.Context, *pb.Null) (*pb.ControlMessage_GetClientControllersResponse, error) {
	return &pb.ControlMessage_GetClientControllersResponse{ControllerNames: s.broker.GetClientNames()}, nil
}

func (s *ControlServer) SubscribeClientControllers(_ *pb.Null, srv pb.Control_SubscribeClientControllersServer) error {
	return errors.New("Not yet implemented")
}

func (s *ControlServer) GetSimulationState(context.Context, *pb.Null) (*pb.SimState, error) {
	ss := s.broker.GetSimState()
	return &ss, nil
}

func (s *ControlServer) SubscribeSimulationState(_ *pb.Null, srv pb.Control_SubscribeSimulationStateServer) error {
	done := make(chan error)
	go func() {
		_, err := srv.Recv()
		if err == nil || err == io.EOF {
			close(done)
		} else {
			done <- err
			close(done)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sscChan := s.broker.GetSimStateListener(ctx)
	for {
		select {
		case ssc := <-sscChan:
			srv.Send(ssc)
		case err := <-done:
			return err
		}
	}
}

func (s *ControlServer) SetSimulationState(_ context.Context, state *pb.SimState) (*pb.Null, error) {
	s.broker.SetSimState(*state)
	return &pb.Null{}, nil
}

func (s *ControlServer) ConnectClientToRobot(_ context.Context, req *pb.ControlMessage_ConnectClientToRobotRequest) (*pb.ControlMessage_ConnectClientToRobotResponse, error) {
	// TODO: sync behavior
	err := s.broker.ConnectClientToRobot(req.GetClientName(), req.GetRobotName(), true)
	if err != nil {
		return &pb.ControlMessage_ConnectClientToRobotResponse{Data: &pb.ControlMessage_ConnectClientToRobotResponse_Error{Error: err.Error()}}, nil
	}
	return &pb.ControlMessage_ConnectClientToRobotResponse{Data: &pb.ControlMessage_ConnectClientToRobotResponse_Ok_{}}, nil
}

func (s *ControlServer) DisconnectClientFromRobot(_ context.Context, req *pb.ControlMessage_DisconnectClientFromRobotRequest) (*pb.ControlMessage_DisconnectClientFromRobotResponse, error) {
	err := s.broker.DisconnectClientFromRobot(req.GetClientName())
	if err != nil {
		return &pb.ControlMessage_DisconnectClientFromRobotResponse{Data: &pb.ControlMessage_DisconnectClientFromRobotResponse_Error{Error: err.Error()}}, nil
	}
	return &pb.ControlMessage_DisconnectClientFromRobotResponse{Data: &pb.ControlMessage_DisconnectClientFromRobotResponse_Ok_{}}, nil
}
