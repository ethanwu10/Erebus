package main

import (
	"context"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"

	pb "github.com/ethanwu10/erebus/broker/gen"
)

type Broker struct {
	ctx      context.Context
	mu       sync.RWMutex
	robots   map[string]*RobotHandle
	clients  map[string]*ClientHandle
	simInfo  SimInfo
	simState pb.SimState

	connections        []connectionIdentifier
	connectionContexts map[string]connectionContext

	simStateListeners map[chan<- *pb.SimState]struct{}
}

type connectionIdentifier struct {
	robotName  string
	clientName string
}

type connectionContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type SimInfo struct {
	timestep int
}

// RobotHandle represents a connected robot to the broker
type RobotHandle struct {
	ctx      context.Context
	cancel   context.CancelFunc
	broker   *Broker
	connBind chan RobotConnection
}

// ClientHandle represents a connected client to the broker
type ClientHandle struct {
	ctx          context.Context
	cancel       context.CancelFunc
	broker       *Broker
	requestsSync bool
	connBind     chan ClientConnection
}

// RobotConnection represents an active connection with a robot
type RobotConnection struct {
	Ctx            context.Context
	SdOut          chan<- *pb.SensorsData
	CmdIn          <-chan *pb.Commands
	SimStateChange <-chan *pb.SimState
	IsSync         bool
}

// ClientConnection represents an active connection with a client
type ClientConnection struct {
	Ctx            context.Context
	SdIn           <-chan *pb.SensorsData
	CmdOut         chan<- *pb.Commands
	SimStateChange <-chan *pb.SimState
	IsSync         bool
}

// NewBroker creates a new broker instance
func NewBroker(ctx context.Context, info SimInfo) *Broker {
	return &Broker{
		ctx:                ctx,
		simInfo:            info,
		robots:             make(map[string]*RobotHandle),
		clients:            make(map[string]*ClientHandle),
		simState:           pb.SimState{State: pb.SimState_RESET},
		connectionContexts: make(map[string]connectionContext),
		simStateListeners:  make(map[chan<- *pb.SimState]struct{}),
	}
}

// RegisterRobot registers a new robot with the given name
func (b *Broker) RegisterRobot(name string, ctx context.Context) *RobotHandle {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.robots[name]; ok {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	connBind := make(chan RobotConnection)
	handle := RobotHandle{
		ctx:      ctx,
		cancel:   cancel,
		connBind: connBind,
		broker:   b,
	}
	b.robots[name] = &handle
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.robots, name)
		log.WithFields(logrus.Fields{
			"robot": name,
		}).Info("Robot unregistered")
	}()
	log.WithFields(logrus.Fields{
		"robot": name,
	}).Info("Robot registered")
	return &handle
}

// UnregisterRobot unregisters an already-registered robot with the given name
func (b *Broker) UnregisterRobot(name string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if _, ok := b.robots[name]; !ok {
		return errors.New("Robot not registered")
	}
	b.robots[name].cancel()
	return nil
}

// RegisterClient registers a new client with the given name
func (b *Broker) RegisterClient(name string, ctx context.Context, requestsSync bool) *ClientHandle {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[name]; ok {
		return nil
	}
	ctx, cancel := context.WithCancel(ctx)
	connBind := make(chan ClientConnection)
	handle := ClientHandle{
		ctx:          ctx,
		cancel:       cancel,
		broker:       b,
		requestsSync: requestsSync,
		connBind:     connBind,
	}
	b.clients[name] = &handle
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.clients, name)
		log.WithFields(logrus.Fields{
			"client": name,
		}).Info("Client unregistered")
	}()
	log.WithFields(logrus.Fields{
		"client": name,
	}).Info("Client registered")
	return &handle
}

// UnregisterClient unregisters an already-registered client with the given name
func (b *Broker) UnregisterClient(name string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if _, ok := b.clients[name]; !ok {
		return errors.New("Client not registered")
	}
	b.clients[name].cancel()
	return nil
}

func (b *Broker) ConnectClientToRobot(clientName string, robotName string, isSync bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	robot, ok := b.robots[robotName]
	if !ok {
		return errors.New("Robot not found")
	}
	client, ok := b.clients[clientName]
	if !ok {
		return errors.New("Client not found")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-ctx.Done(): // prevent leaking goroutine
		case <-robot.ctx.Done():
			cancel()
		case <-client.ctx.Done():
			cancel()
		}
	}()
	b.connections = append(b.connections,
		connectionIdentifier{clientName: clientName, robotName: robotName},
	)
	go func() {
		<-ctx.Done()
		// TODO: remove entry from b.connections
	}()
	b.connectionContexts[clientName] = connectionContext{ctx: ctx, cancel: cancel}
	// TODO: handle sync
	// TODO: MITM channels for instrumentation
	sdChan := make(chan *pb.SensorsData)
	cmdChan := make(chan *pb.Commands)
	b.mu.Unlock()
	rConnSSC := b.GetSimStateListener(ctx)
	b.mu.Lock()
	robot.connBind <- RobotConnection{
		Ctx:            ctx,
		SdOut:          sdChan,
		CmdIn:          cmdChan,
		SimStateChange: rConnSSC,
		IsSync:         isSync,
	}
	b.mu.Unlock()
	cConnSSC := b.GetSimStateListener(ctx)
	b.mu.Lock()
	client.connBind <- ClientConnection{
		Ctx:            ctx,
		SdIn:           sdChan,
		CmdOut:         cmdChan,
		SimStateChange: cConnSSC,
		IsSync:         isSync,
	}
	return nil
}

func (b *Broker) DisconnectClientFromRobot(clientName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	connCtx, ok := b.connectionContexts[clientName]
	if !ok {
		return errors.New("Client not connected")
	}
	connCtx.cancel()
	return nil
}

func (b *Broker) GetSimStateListener(ctx context.Context) <-chan *pb.SimState {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan *pb.SimState)
	b.simStateListeners[ch] = struct{}{}
	go func() {
		<-ctx.Done()
		close(ch)
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.simStateListeners, ch)
	}()
	return ch
}

// GetRobotNames returns the names of all registered robots
func (b *Broker) GetRobotNames() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	names := make([]string, 0, len(b.robots))
	for name := range b.robots {
		names = append(names, name)
	}
	return names
}

// GetClientNames returns the names of all registered clients
func (b *Broker) GetClientNames() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	names := make([]string, 0, len(b.clients))
	for name := range b.clients {
		names = append(names, name)
	}
	return names
}

// GetSimState gets the current simulation state
func (b *Broker) GetSimState() pb.SimState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.simState
}

// SetSimState sets the simulation state
func (b *Broker) SetSimState(state pb.SimState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.simState = state
	for listener := range b.simStateListeners {
		go func(listener chan<- *pb.SimState) {
			listener <- &state
		}(listener)
	}
}

// GetConnection returns a channel where the connection will be sent once it is
// established
func (r *RobotHandle) GetConnection() <-chan RobotConnection {
	return r.connBind
}

// GetConnection returns a channel where the peer connection will be sent once
// it is established
func (c *ClientHandle) GetConnection() <-chan ClientConnection {
	return c.connBind
}
