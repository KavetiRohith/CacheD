package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

// Store is the interface Raft-backed key-value stores must implement.
type Store interface {
	// Get returns the value for the given key.
	Get(key string) (string, error)

	// Set sets the value for the given key, via distributed consensus.
	Set(key, value string) error

	// Delete removes the given key, via distributed consensus.
	Del(key string) error

	// Join joins the node, identitifed by nodeID and reachable at addr, to the cluster.
	Join(nodeID string, addr string) error
}

// Server provides HTTP server.
type Server struct {
	addr string
	ln   net.Listener

	store Store
}

// New returns an uninitialized HTTP server.
func New(addr string, store Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
	}
}

// Start starts the server.
func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("accept error %s\n", err)
				continue
			}

			go s.handleConn(conn)
		}
	}()

	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Printf("connection closed: %s\n", conn.RemoteAddr())
	}()

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		s.parseCommand(conn, scanner.Bytes())
	}
}

func (s *Server) parseCommand(conn net.Conn, rawCmd []byte) {
	var (
		parts   = strings.Fields(string(rawCmd))
		len_cmd = len(parts)
	)

	if len_cmd < 2 {
		conn.Write([]byte("message must atleast have command, key or nodeID and remoteAddr\n"))
		return
	}

	cmd := parts[0]

	switch cmd {
	case "JOIN":
		if len_cmd < 3 {
			conn.Write([]byte("JOIN message must have nodeID and remoteAddr\n"))
			return
		}

		var (
			nodeID     = parts[1]
			remoteAddr = parts[2]
		)

		s.handleJoin(conn, nodeID, remoteAddr)

	case "GET":
		key := parts[1]

		s.handleGet(conn, key)

	case "SET":
		if len_cmd < 3 {
			conn.Write([]byte("SET message must have key and value\n"))
			return
		}

		var (
			key = parts[1]
			val = parts[2]
		)
		s.handleSet(conn, key, val)

	case "DEL":
		key := parts[1]

		s.handleDel(conn, key)

	default:
		conn.Write([]byte("Invalid command in message"))
	}
}

func (s *Server) handleSet(conn net.Conn, key string, val string) {
	err := s.store.Set(key, val)
	if err != nil {
		log.Println("SET Error: ", err)
		conn.Write([]byte(err.Error() + "\n"))
		return
	}

	conn.Write([]byte("Success\n"))
	log.Printf("SET %s %s\n", key, val)
}

func (s *Server) handleGet(conn net.Conn, key string) {
	val, err := s.store.Get(key)
	if err != nil {
		log.Println("GET Error: ", err)
		conn.Write([]byte(err.Error() + "\n"))
		return
	}

	conn.Write([]byte(fmt.Sprintf("%s\n", val)))
	log.Printf("GET %s %s\n", key, val)
}

func (s *Server) handleDel(conn net.Conn, key string) {
	err := s.store.Del(key)
	if err != nil {
		log.Println("DEL Error: ", err)
		conn.Write([]byte(err.Error() + "\n"))
		return
	}

	conn.Write([]byte("Success\n"))
	log.Printf("DEL %s\n", key)
}

func (s *Server) handleJoin(conn net.Conn, nodeID, remoteAddr string) {
	err := s.store.Join(nodeID, remoteAddr)
	if err != nil {
		log.Println("JOIN Error: ", err)
		conn.Write([]byte(err.Error() + "\n"))
		return
	}

	conn.Write([]byte("Success\n"))
}

// Addr returns the address on which the Server is listening
func (s *Server) Addr() net.Addr {
	return s.ln.Addr()
}

// Close closes the server.
func (s *Server) Close() {
	s.ln.Close()
}
