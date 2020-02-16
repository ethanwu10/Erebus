package main

import (
	"io"

	"github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

type WbControllerServer struct {
	pb.UnimplementedWbControllerServer

	broker *Broker
}

func NewWbControllerServer(broker *Broker) *WbControllerServer {
	return &WbControllerServer{broker: broker}
}

func (s *WbControllerServer) Session(srv pb.WbController_SessionServer) error {
	hasInitialized := false
	var robotHandle *RobotHandle
	var name string
	var logger *logrus.Entry
	for !hasInitialized {
		msg, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch msg.Message.(type) {
		case *pb.WbControllerMessage_ClientMessage_WbControllerHandshake:
			name = msg.GetWbControllerHandshake().GetRobotName()
			logger = log.WithFields(logrus.Fields{
				"robot": name,
			})
			// TODO: handle RobotInfo
			robotHandle = s.broker.RegisterRobot(name, srv.Context())
			if robotHandle == nil {
				srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerHandshakeResponse{
					WbControllerHandshakeResponse: &pb.WbControllerHandshakeResponse{Data: &pb.WbControllerHandshakeResponse_Error{
						Error: "name in use",
					}},
				}})
				logger.Info("Robot rejected for duplicate name")
				return nil
			}
			hasInitialized = true
			srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerHandshakeResponse{
				WbControllerHandshakeResponse: &pb.WbControllerHandshakeResponse{Data: &pb.WbControllerHandshakeResponse_Ok_{
					Ok: &pb.WbControllerHandshakeResponse_Ok{Timestep: int32(s.broker.simInfo.timestep)},
				}},
			}})
			logger.Info("Robot connected")
		}
	}
	incoming := make(chan *pb.WbControllerMessage_ClientMessage)
	go func() {
		for {
			msg, err := srv.Recv()
			if err != nil {
				// TODO: better handle incoming errors
				close(incoming)
				return
			}
			incoming <- msg
		}
	}()
	for {
		var connection RobotConnection
		logger.Debug("Robot waiting for peer")
		select {
		case connection = <-robotHandle.GetConnection():
		case <-srv.Context().Done():
			return nil
		}
		if err := srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerBound{
			WbControllerBound: &pb.WbControllerBound{IsSync: connection.IsSync},
		}}); err != nil {
			logger.Errorf("Couldn't send bound message: %s", err.Error())
			return err
		}
		logger.Debug("Robot got peer")
	LBoundSession:
		for {
			select {
			case controllerMsg, ok := <-incoming:
				if !ok {
					// Remote hung up
					logger.Info("Robot disconnected")
					return nil
				}
				if sd := controllerMsg.GetSensorData(); sd != nil {
					connection.SdOut <- sd
				}
			case cmd, ok := <-connection.CmdIn:
				if !ok {
					continue
				}
				err := srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_Commands{Commands: cmd}})
				if err != nil {
					logger.Errorf("Couldn't send commands message: %s", err.Error())
					return err
				}
			case ssc, ok := <-connection.SimStateChange:
				if !ok {
					continue
				}
				err := srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_SimStateChange{SimStateChange: ssc}})
				if err != nil {
					logger.Errorf("Couldn't send sim state change message: %s", err.Error())
					return err
				}
			case <-connection.Ctx.Done():
				err := srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerUnbound{WbControllerUnbound: &pb.WbControllerUnbound{}}})
				if err != nil {
					logger.Errorf("Couldn't send unbound message: %s", err.Error())
					return err
				}
				break LBoundSession
			}
		}
	}
}
