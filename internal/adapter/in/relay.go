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
			return
		}
		target.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(client, target)
		client.(*net.TCPConn).CloseWrite()
		done <- struct{}{}
	}()

	<-done
	<-done
}
