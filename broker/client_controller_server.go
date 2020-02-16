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
			clientHandle = s.broker.RegisterClient(name, srv.Context(), handshake.GetRequestSync())
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
			if err := srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerHandshakeResponse{
				ClientControllerHandshakeResponse: &pb.ClientControllerHandshakeResponse{Data: &pb.ClientControllerHandshakeResponse_Ok_{
					Ok: &pb.ClientControllerHandshakeResponse_Ok{
						Timestep: int32(s.broker.simInfo.timestep),
					},
				}},
			}}); err != nil {
				logger.Errorf("Couldn't send handshake response: %s", err.Error())
				return err
			}
			logger.Info("Client connected")
		}
	}
	incoming := make(chan *pb.ClientControllerMessage_ControllerMessage)
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
		var connection ClientConnection
		logger.Debug("Client waiting for peer")
		select {
		case connection = <-clientHandle.GetConnection():
		case <-srv.Context().Done():
			return nil
		}
		if err := srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerBound{
			ClientControllerBound: &pb.ClientControllerBound{IsSync: connection.IsSync},
		}}); err != nil {
			logger.Errorf("Couldn't send bound message: %s", err.Error())
			return err
		}
		logger.Debug("Client got peer")
	LBoundSession:
		for {
			select {
			case controllerMsg, ok := <-incoming:
				if !ok {
					// Remote hung up
					logger.Info("Client disconnected")
					return nil
				}
				if cmd := controllerMsg.GetCommands(); cmd != nil {
					connection.CmdOut <- cmd
				}
			case sd, ok := <-connection.SdIn:
				if !ok {
					continue
				}
				err := srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_SensorData{SensorData: sd}})
				if err != nil {
					logger.Errorf("Couldn't send sensor data message: %s", err.Error())
					return err
				}
			case ssc, ok := <-connection.SimStateChange:
				if !ok {
					continue
				}
				err := srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_SimStateChange{SimStateChange: ssc}})
				if err != nil {
					logger.Errorf("Couldn't send sim state change message: %s", err.Error())
					return err
				}
			case <-connection.Ctx.Done():
				err := srv.Send(&pb.ClientControllerMessage_ServerMessage{Message: &pb.ClientControllerMessage_ServerMessage_ClientControllerUnbound{ClientControllerUnbound: &pb.ClientControllerUnbound{}}})
				if err != nil {
					logger.Errorf("Couldn't send unbound message: %s", err.Error())
					return err
				}
				break LBoundSession
			}
		}
	}
}
