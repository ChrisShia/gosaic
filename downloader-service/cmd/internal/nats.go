package internal

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func NatsServer() *server.Server {
	serverOpts := &server.Options{
		ServerName:      "embedded-nats",
		DontListen:      false,
		JetStream:       true,
		JetStreamDomain: "download",
	}

	s, _ := server.NewServer(serverOpts)

	return s
}

func NatsClient(ns *server.Server) (*nats.Conn, error) {
	clientConnection, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return nil, err
	}

	return clientConnection, nil
}
