package client

import (
	"net"

	"github.com/bamboovir/cs425/lib/mp3/config"
	"github.com/bamboovir/cs425/lib/mp3/transaction"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	logger = log.WithField("src", "client")
)

type ClientManager struct {
	ClientID string
	clients  map[string]transaction.TransactionClient
	conns    map[string]*grpc.ClientConn
}

func New(clientID string, serverConfig *config.Config) (*ClientManager, error) {
	clientManager := &ClientManager{
		ClientID: clientID,
		clients:  map[string]transaction.TransactionClient{},
		conns:    map[string]*grpc.ClientConn{},
	}

	for _, currConfig := range serverConfig.ConfigItems {
		addr := net.JoinHostPort(currConfig.NodeHost, currConfig.NodePort)
		logger.Infof("try connect to server [%s][%s]", currConfig.NodeID, addr)
		conn, err := newConn(addr)
		if err != nil {
			return nil, err
		}
		logger.Infof("connect to server [%s][%s] succ", currConfig.NodeID, addr)
		clientManager.conns[currConfig.NodeID] = conn
		client := transaction.NewTransactionClient(conn)
		clientManager.clients[currConfig.NodeID] = client
	}

	return clientManager, nil
}

func (c *ClientManager) GetClientByID(nid string) (transaction.TransactionClient, bool) {
	client, ok := c.clients[nid]
	return client, ok
}

func (c *ClientManager) GetFirstClient() transaction.TransactionClient {
	for _, client := range c.clients {
		return client
	}
	return nil
}

func (c *ClientManager) GetAllClients() map[string]transaction.TransactionClient {
	return c.clients
}

func (c *ClientManager) Close() (err error) {
	for _, conn := range c.conns {
		err = conn.Close()
	}

	return err
}

func newConn(addr string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		logger.Errorf("client did not connect to server [%s], retry", addr)
		return nil, err
	}

	return conn, nil
}
