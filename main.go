package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"sync"
)

const PORT = ":3000"
const MAX_ROOM_PLAYER = 2

var playerAutoInc autoInc
var roomAutoInc autoInc

type autoInc struct {
	sync.Mutex
	id int
}

func (a *autoInc) ID() (id int) {
	a.Lock()
	defer a.Unlock()

	id = a.id
	a.id++
	return
}

type Player struct {
	ID   int
	Conn net.Conn
}

func newPlayer(conn net.Conn) Player {
	return Player{
		ID:   playerAutoInc.ID(),
		Conn: conn,
	}
}

type Room struct {
	server  *http.Server
	mu      sync.RWMutex
	ID      int
	Players []Player
}

type GameHandler struct{}

func newGameHandler() *GameHandler {
	return &GameHandler{}
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello")
}

type ServerHandler struct{}

func newServerHandler() *ServerHandler {
	return &ServerHandler{}
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the matchmaking server")
}

func newRoom() Room {
	s := &http.Server{
		Addr: ":3001",
	}
	return Room{
		server:  s,
		ID:      roomAutoInc.ID(),
		Players: make([]Player, 0, MAX_ROOM_PLAYER),
	}
}

func (r *Room) addPlayer(player Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Players) >= MAX_ROOM_PLAYER {
		return false
	}

	r.Players = append(r.Players, player)
	return true
}

type GameServer struct {
	mu            sync.RWMutex
	PlayerAutoInc int
	RoomAutoInc   int
	Rooms         []*Room
	Players       []*Player
	httpServer    *http.Server
}

func newGameServer() *GameServer {
	return &GameServer{
		httpServer:    &http.Server{},
		PlayerAutoInc: playerAutoInc.ID(),
		RoomAutoInc:   roomAutoInc.ID(),
		Rooms:         make([]*Room, 0),
		Players:       make([]*Player, 0),
	}
}

func (gs *GameServer) Start() {

	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		panic(err)
	}

	slog.Info("Server started", "port", PORT)
	gs.httpServer.Handler = newGameHandler()
	gs.httpServer.Addr = ":3002"
	go gs.httpServer.ListenAndServe()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go gs.handleConnexion(conn)
	}
}

func (gs *GameServer) handleConnexion(conn net.Conn) {
	message := make([]byte, 0, 4096)
	player := newPlayer(conn)
	var destinationPort string

	reader := bufio.NewReader(conn)
	for {
		line, _, _ := reader.ReadLine()
		if len(line) == 0 {
			break
		}
		message = append(message, line...)
	}

	// slog.Info(string(message))

	_ = destinationPort

	// Add player to the global list of players
	gs.mu.Lock()
	gs.Players = append(gs.Players, &player)
	gs.mu.Unlock()

	// Tring to find a room for the player
	gs.mu.Lock()
	roomFound := false
	for i := range gs.Rooms {
		if gs.Rooms[i].addPlayer(player) {
			destinationPort = gs.Rooms[i].server.Addr
			slog.Info("Player added to an existing room",
				"playerID", player.ID,
				"roomID", gs.Rooms[i].ID)
			roomFound = true
			break
		}
	}

	if !roomFound {
		room := newRoom()
		destinationPort = room.server.Addr
		slog.Info("Starting web Server")
		handler := newGameHandler()
		room.server.Handler = handler
		go room.server.ListenAndServe()
		room.addPlayer(player)
		gs.Rooms = append(gs.Rooms, &room)
		slog.Info("Player added to new room",
			"playerID", player.ID,
			"roomID", room.ID)
	}
	gs.mu.Unlock()

	defer func() {
		conn.Close()
		gs.removePlayer(player)
	}()
}

func (gs *GameServer) removePlayer(player Player) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	for i, p := range gs.Players {
		if p.ID == player.ID {
			gs.Players = append(gs.Players[:i], gs.Players[i+1:]...)
			break
		}
	}

	// Remove from room and clean up empty rooms
	for i := len(gs.Rooms) - 1; i >= 0; i-- {
		room := gs.Rooms[i]
		room.mu.Lock()
		for j, p := range room.Players {
			if p.ID == player.ID {
				room.Players = append(room.Players[:j], room.Players[j+1:]...)
				if len(room.Players) == 0 {
					// Remove empty room
					gs.Rooms = append(gs.Rooms[:i], gs.Rooms[i+1:]...)
				}
				break
			}
		}
		room.mu.Unlock()
	}
}

func main() {
	gs := newGameServer()
	gs.Start()
}
