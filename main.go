package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/KavetiRohith/GoCached/server"
	"github.com/KavetiRohith/GoCached/store"
)

// Command line defaults
const (
	DefaultTCPAddr  = "localhost:3001"
	DefaultRaftAddr = "localhost:4001"
)

// Command line parameters
var inMemmory bool
var tcpAddr string
var raftAddr string
var joinAddr string
var nodeID string
var raftDir string

func init() {
	flag.BoolVar(&inMemmory, "inMemmory", false, "Use in-memory storage for Raft")
	flag.StringVar(&tcpAddr, "tcpAddr", DefaultTCPAddr, "Set the TCP bind address")
	flag.StringVar(&raftAddr, "rAddr", DefaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set join address, if any")
	flag.StringVar(&raftDir, "raftDir", "", "Raft storage directory path")
	flag.StringVar(&nodeID, "id", "", "Node ID")
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	flag.Parse()

	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	os.MkdirAll(raftDir, 0700)

	s := store.New(raftAddr, raftDir, inMemmory)

	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := server.New(tcpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start server: %s", err.Error())
	}

	// If join was specified, make the join request.
	if joinAddr != "" {
		if err := join(joinAddr, nodeID, raftAddr); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("GoCached started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("GoCached exiting")
}

func join(joinAddr, nodeID, raftAddr string) error {

	conn, err := net.Dial("tcp", joinAddr)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = fmt.Fprintf(conn, "JOIN %s %s", nodeID, raftAddr)
	if err != nil {
		return err
	}

	return nil
}
