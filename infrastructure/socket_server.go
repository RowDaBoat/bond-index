package infrastructure

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"bond/actions"
)

type SocketServer struct {
	listener     net.Listener
	syncRequests chan actions.SyncRequest
}

func NewSocketServer(dataDir string, syncRequests chan actions.SyncRequest) *SocketServer {
	os.MkdirAll(dataDir, 0755)
	socketPath := filepath.Join(dataDir, "bond.sock")
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Failed to listen on %s: %v\n", socketPath, err)
		os.Exit(1)
	}

	return &SocketServer{
		listener:     listener,
		syncRequests: syncRequests,
	}
}

func (ss *SocketServer) Start() {
	for {
		conn, err := ss.listener.Accept()
		if err != nil {
			return
		}
		go ss.handleConnection(conn)
	}
}

func (ss *SocketServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	msg, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return
	}

	switch strings.TrimSpace(msg) {
	case actions.CmdSync:
		reply := make(actions.SyncRequest, 1)
		ss.syncRequests <- reply
		blockHeight := <-reply
		fmt.Fprintf(conn, "%d\n", blockHeight)
	default:
		fmt.Fprintf(conn, "unknown command: %s\n", strings.TrimSpace(msg))
	}
}

func (ss *SocketServer) Close() error {
	return ss.listener.Close()
}

func SocketPath(dataDir string) string {
	return filepath.Join(dataDir, "bond.sock")
}
