package GameSpy

import (
	"net"

	log "github.com/ReviveNetwork/GoRevive/Log"
)

// Socket is a basic event-based TCP-Server
type Socket struct {
	Clients   []*Client
	name      string
	port      string
	listen    net.Listener
	eventChan chan SocketEvent
}

// SocketEvent is the generic struct for events
// by this socket
type SocketEvent struct {
	Name string
	Data interface{}
}

// New starts to listen on a new Socket
func (socket *Socket) New(name string, port string) (chan SocketEvent, error) {
	var err error

	socket.name = name
	socket.port = port
	socket.eventChan = make(chan SocketEvent, 1000)

	// Listen for incoming connections.
	socket.listen, err = net.Listen("tcp", "0.0.0.0:"+socket.port)
	if err != nil {
		log.Errorf("%s: Listening on 0.0.0.0:%s threw an error.\n%v", socket.name, socket.port, err)
		return nil, err
	}
	log.Noteln(socket.name + ": Listening on 0.0.0.0:" + socket.port)

	// Accept new connections in a new Goroutine("thread")
	go socket.run()

	return socket.eventChan, nil
}

// Close fires a close-event and closes the socket
func (socket *Socket) Close() {
	// Fire closing event
	log.Noteln(socket.name + " closing. Port " + socket.port)
	socket.eventChan <- SocketEvent{
		Name: "close",
		Data: nil,
	}

	// Close socket
	socket.listen.Close()
}

func (socket *Socket) run() {
	for {
		// Listen for an incoming connection.
		conn, err := socket.listen.Accept()
		if err != nil {
			log.Errorf("%s: A new client connecting threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: err,
			}
		}

		// Create a new Client and add it to our slice
		log.Noteln(socket.name + ": A new client connected")
		newClient := new(Client)
		clientEventSocket, err := newClient.New(socket.name, &conn)
		if err != nil {
			log.Errorf("%s: Creating the new client threw an error.\n%v", socket.name, err)
			socket.eventChan <- SocketEvent{
				Name: "error",
				Data: err,
			}
		}
		go socket.handleClientEvents(newClient, clientEventSocket)

		socket.Clients = append(socket.Clients, newClient)

		// Fire newClient event
		socket.eventChan <- SocketEvent{
			Name: "newClient",
			Data: newClient,
		}
	}
}

func (socket *Socket) handleClientEvents(client *Client, eventsChannel chan ClientEvent) {
	for {
		select {
		case event := <-eventsChannel:
			switch {
			case event.Name == "command":
				command := event.Data.(*Command)
				log.Debugln(command)
			default:
				log.Debugln(event)
			}
		}
	}
}