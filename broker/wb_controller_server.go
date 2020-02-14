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
			robotHandle = s.broker.RegisterRobot(name)
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
			defer s.broker.UnregisterRobot(name)
		}
	}
	for {
		connection := robotHandle.GetConnection()
		srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerBound{
			WbControllerBound: &pb.WbControllerBound{IsSync: connection.IsSync},
		}})
		sdOut := make(chan *pb.SensorsData)
		closed := make(chan struct{})
		go func() {
			for {
				msg, err := srv.Recv()
				if err != nil {
					// TODO: better handle errors
					close(closed)
					return
				}
				select {
				case <-connection.Ctx.Done():
					return
				default:
					if sd := msg.GetSensorData(); sd != nil {
						sdOut <- sd
					}
				}
			}
		}()
		select {
		case sd := <-sdOut:
			connection.SdOut <- sd
		case cmd := <-connection.CmdIn:
			srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_Commands{Commands: cmd}})
		case ssc := <-connection.SimStateChange:
			srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_SimStateChange{SimStateChange: ssc}})
		case <-connection.Ctx.Done():
			srv.Send(&pb.WbControllerMessage_ServerMessage{Message: &pb.WbControllerMessage_ServerMessage_WbControllerUnbound{}})
		case <-closed:
			// Remote hung up
			logger.Info("Robot disconnected")
			return nil
		}
	}
}
