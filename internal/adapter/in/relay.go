package in

import (
	"io"
	"log/slog"
	"net"
)

func (s *ServerAdapter) relay(client, target net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		_, err := io.Copy(target, client)
		if err != nil {
			slog.Error("Error relaying data from client to target",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		}
		if tcp, ok := target.(*net.TCPConn); ok {
			if err := tcp.CloseWrite(); err != nil {
				slog.Error("Error closing write on target", slog.Any("err", err))
			}
		}
		done <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(client, target)
		if err != nil {
			slog.Error("Error relaying data from target to client",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		}
		if tcp, ok := client.(*net.TCPConn); ok {
			if err := tcp.CloseWrite(); err != nil {
				slog.Error("Error closing write on client", slog.Any("err", err))
			}
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}
