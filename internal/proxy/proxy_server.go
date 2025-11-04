package proxy

import (
	"log/slog"
	"net"
)

func Start(port int) {
	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{Port: port})
	if err != nil {
		slog.Error("Listener creation error", slog.Any("err", err))
		return
	}
	slog.Info("Proxy server is listening", slog.Any("addr", listener.Addr()))

	defer func(listener *net.TCPListener) {
		err := listener.Close()
		if err != nil {
			slog.Error("Listener closing error", slog.Any("err", err))
		}
		slog.Info("Proxy server stopped", slog.Any("addr", listener.Addr()))
	}(listener)

	for {
		client, err := listener.AcceptTCP()
		if err != nil {
			slog.Error("Client accepting error", slog.Any("err", err))
			return
		}
		slog.Info("Accepted client", slog.Any("remote_addr", client.RemoteAddr()))

		go handleClient(client)
	}
}

func handleClient(client *net.TCPConn) {
	defer func(client *net.TCPConn) {
		err := client.Close()
		if err != nil {
			slog.Error("Client connection closing error", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		}
		slog.Info("Client connection closed", slog.Any("remote_addr", client.RemoteAddr()))
	}(client)

	authReply, err := authenticate(client)
	if err != nil {
		slog.Error("Authentication error", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		if err := sendAuthReply(client, authReply); err != nil {
			slog.Error("Failed to send auth reply", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		}
		return
	}
	if err := sendAuthReply(client, authReply); err != nil {
		slog.Error("Failed to send auth reply", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		return
	}
	slog.Info("Client is authenticated", slog.Any("remote_addr", client.RemoteAddr()))

	// Handle command
	peer, commandReply, err := connectCommand(client)
	if err != nil {
		slog.Error("Command execution error", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		if err := sendCommandReply(client, commandReply); err != nil {
			slog.Error("Failed to send command reply", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		}
		return
	}
	if err := sendCommandReply(client, commandReply); err != nil {
		slog.Error("Failed to send command reply", slog.Any("remote_addr", client.RemoteAddr()), slog.Any("err", err))
		return
	}
	slog.Info("Proxy server is connected to peer", slog.Any("remote_addr", client.RemoteAddr()))

	// Transfer data
	transferData(client, peer)
}
