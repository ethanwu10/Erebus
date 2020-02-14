package main

import (
	"io"

	"github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

type ClientControllerServer struct {
	pb.UnimplementedClientControllerServer

	broker *Broker
}

func NewClientControllerServer(broker *Broker) *ClientControllerServer {
	return &ClientControllerServer{broker: broker}
}

func (s *ClientControllerServer) Session(srv pb.ClientController_SessionServer) error {
	hasInitialized := false
	var clientHandle *ClientHandle
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
		case *pb.ClientControllerMessage_ControllerMessage_ClientControllerHandshake:
			handshake := msg.GetClientControllerHandshake()
			name = handshake.GetClientName()
			logger = log.WithFields(logrus.Fields{
				"client": name,
			})
			clientHandle := s.broker.RegisterClient(name, handshake.GetRequestSync())
			if clientHandle == nil {
				srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerHandshakeResponse{
					ClientControllerHandshakeResponse: &pb.ClientControllerHandshakeResponse{Data: &pb.ClientControllerHandshakeResponse_Error{
						Error: "name in use",
					}},
				}})
				logger.Info("Client rejected for duplicate name")
				return nil
			}
			hasInitialized = true
			srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerHandshakeResponse{
				ClientControllerHandshakeResponse: &pb.ClientControllerHandshakeResponse{Data: &pb.ClientControllerHandshakeResponse_Ok_{
					Ok: &pb.ClientControllerHandshakeResponse_Ok{
						Timestep: int32(s.broker.simInfo.timestep),
					},
				}},
			}})
			logger.Info("Client connected")
			defer s.broker.UnregisterClient(name)
		}
	}
	for {
		connection := clientHandle.GetConnection()
		srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerBound{
			ClientControllerBound: &pb.ClientControllerBound{IsSync: connection.IsSync},
		}})
		cmdOut := make(chan *pb.Commands)
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
					if cmd := msg.GetCommands(); cmd != nil {
						cmdOut <- cmd
					}
				}
			}
		}()
		select {
		case cmd := <-cmdOut:
			connection.CmdOut <- cmd
		case sd := <-connection.SdIn:
			srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_SensorData{SensorData: sd}})
		case ssc := <-connection.SimStateChange:
			srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_SimStateChange{SimStateChange: ssc}})
		case <-connection.Ctx.Done():
			srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerUnbound{}})
		case <-closed:
			// Remote hung up
			logger.Info("Client disconnected")
			return nil
		}
	}
}
