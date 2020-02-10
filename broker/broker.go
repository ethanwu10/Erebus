package main

import (
	"context"
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"

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

type RobotHandle struct {
	broker   *Broker
	connBind chan RobotConnection
}

type ClientHandle struct {
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
		ctx:               ctx,
		simInfo:           info,
		simState:          pb.SimState{State: pb.SimState_RESET},
		simStateListeners: make(map[chan<- *pb.SimState]struct{}),
	}
}

// RegisterRobot registers a new robot with the given name
func (b *Broker) RegisterRobot(name string) *RobotHandle {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.robots[name]; ok {
		return nil
	}
	connBind := make(chan RobotConnection)
	handle := RobotHandle{
		connBind: connBind,
		broker:   b,
	}
	b.robots[name] = &handle
	log.WithFields(log.Fields{
		"robot": name,
	}).Info("Robot registered")
	return &handle
}

// UnregisterRobot unregisters an already-registered robot with the given name
func (b *Broker) UnregisterRobot(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.robots[name]; !ok {
		return errors.New("Robot not registered")
	}
	delete(b.robots, name)
	// TODO: kill client connections
	log.WithFields(log.Fields{
		"robot": name,
	}).Info("Robot unregistered")
	return nil
}

// RegisterClient registers a new client with the given name
func (b *Broker) RegisterClient(name string, requestsSync bool) *ClientHandle {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[name]; ok {
		return nil
	}
	connBind := make(chan ClientConnection)
	handle := ClientHandle{
		broker:       b,
		requestsSync: requestsSync,
		connBind:     connBind,
	}
	b.clients[name] = &handle
	log.WithFields(log.Fields{
		"client": name,
	}).Info("Client registered")
	return &handle
}

// UnregisterClient unregisters an already-registered client with the given name
func (b *Broker) UnregisterClient(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[name]; !ok {
		return errors.New("Client not registered")
	}
	delete(b.clients, name)
	// TODO: kill connections
	log.WithFields(log.Fields{
		"client": name,
	}).Info("Client unregistered")
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
	b.connections = append(b.connections,
		connectionIdentifier{clientName: clientName, robotName: robotName},
	)
	b.connectionContexts[clientName] = connectionContext{ctx: ctx, cancel: cancel}
	// TODO: handle sync
	// TODO: MITM channels for instrumentation
	sdChan := make(chan *pb.SensorsData)
	cmdChan := make(chan *pb.Commands)
	rConnSSC := make(chan *pb.SimState)
	robot.connBind <- RobotConnection{
		Ctx:            ctx,
		SdOut:          sdChan,
		CmdIn:          cmdChan,
		SimStateChange: rConnSSC,
		IsSync:         isSync,
	}
	cConnSSC := make(chan *pb.SimState)
	client.connBind <- ClientConnection{
		Ctx:            ctx,
		SdIn:           sdChan,
		CmdOut:         cmdChan,
		SimStateChange: cConnSSC,
		IsSync:         isSync,
	}
	b.simStateListeners[rConnSSC] = struct{}{}
	b.simStateListeners[cConnSSC] = struct{}{}
	go func() {
		<-ctx.Done()
		close(rConnSSC)
		close(cConnSSC)
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.simStateListeners, rConnSSC)
		delete(b.simStateListeners, cConnSSC)
	}()
	return nil
}

func (b *Broker) DisconnectClientFromRobot(clientName string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	connCtx, ok := b.connectionContexts[clientName]
	if !ok {
		return errors.New("Client not connnected")
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

// GetClientNames reutnrs the names of all registered clients
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
		listener <- &state
	}
}

// GetConnection blocks until a peer connection is established, returning the
// connection
func (r *RobotHandle) GetConnection() RobotConnection {
	return <-r.connBind
}

// GetConnection blocks until a peer connection is established, returning the
// connection
func (c *ClientHandle) GetConnection() ClientConnection {
	return <-c.connBind
}
