package app

import (
	"fmt"
	"log/slog"
	"net"
	"tcp-proxy/internal/adapter/in"
	"tcp-proxy/internal/service"
)

type App struct {
	port          int
	proxyService  *service.ProxyService
	serverAdapter *in.ServerAdapter
}

func New(port int) *App {
	resolver := service.NewResolver()                        // отвечает только за разрешение доменов
	proxyService := service.NewProxyService(resolver)        // за логику прокси
	serverAdapter := in.NewServerAdapter(port, proxyService) // за взаимодействие с клиентами

	return &App{
		port:          port,
		proxyService:  proxyService,
		serverAdapter: serverAdapter,
	}
}

func (a *App) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", a.port, err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			slog.Error("Failed to close listener", slog.Any("err", err))
		}
	}(listener)

	slog.Info("SOCKS5 proxy server listening", slog.Int("port", a.port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", slog.Any("err", err))
			continue
		}

		go a.serverAdapter.HandleConnection(conn)
	}
}
