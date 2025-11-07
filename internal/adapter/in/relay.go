package in

import (
	"errors"
	"io"
	"log/slog"
	"net"
	"strings"
)

func (s *ServerAdapter) relay(client, target net.Conn) {
	done := make(chan struct{}, 2)

	go func() {
		_, err := io.Copy(target, client)
		if err != nil && !isIgnorableError(err) {
			slog.Error("Error relaying data from client to target",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		} else if err != nil && isIgnorableError(err) {
			slog.Info("Error relaying data from client to target",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		}
		if tcp, ok := target.(*net.TCPConn); ok {
			if err := tcp.CloseWrite(); err != nil && !isIgnorableCloseError(err) {
				slog.Error("Error closing write on target", slog.Any("err", err))
			}
		}
		done <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(client, target)
		if err != nil && !isIgnorableError(err) {
			slog.Error("Error relaying data from target to client",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		} else if err != nil && isIgnorableError(err) {
			slog.Info("Error relaying data from client to target",
				slog.Any("client_addr", client.RemoteAddr()),
				slog.Any("target_addr", target.RemoteAddr()),
				slog.Any("err", err),
			)
		}
		if tcp, ok := client.(*net.TCPConn); ok {
			if err := tcp.CloseWrite(); err != nil && !isIgnorableCloseError(err) {
				slog.Error("Error closing write on client", slog.Any("err", err))
			}
		}
		done <- struct{}{}
	}()

	<-done
	<-done
}

func isIgnorableError(err error) bool {
	return errors.Is(err, io.EOF) || strings.Contains(err.Error(), "connection reset by peer")
}

func isIgnorableCloseError(err error) bool {
	return strings.Contains(err.Error(), "socket is not connected")
}
