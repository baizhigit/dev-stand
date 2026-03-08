package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"syscall"

	"github.com/baizhigit/dev-stand/services/api-gateway/config"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/closer"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/healthcheck"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
)

type App struct {
	// main servers
	grpcServer *grpc.Server
	httpServer *http.Server // grpc-gateway HTTP proxy

	// admin (debug, healthcheck) — separate port, starts before main
	adminListener net.Listener
	adminMux      *chi.Mux

	// downstream gRPC connections
	grpcConns map[string]grpc.ClientConnInterface

	// state flags — typed atomics (Go 1.19+), no unsafe int32 tricks
	started    atomic.Int32
	terminated atomic.Int32

	healthCheck  healthcheck.Handler
	publicCloser *closer.Closer // owns main server + gRPC conns
	adminCloser  *closer.Closer // owns admin server
}

// New initializes and returns the App. Never calls os.Exit.
// All fatal decisions are left to the caller (main).
func New(ctx context.Context) (*App, error) {
	app := &App{
		grpcConns:    make(map[string]grpc.ClientConnInterface),
		publicCloser: closer.New(syscall.SIGTERM, syscall.SIGINT),
		adminCloser:  closer.New(),
	}

	// Admin server starts FIRST — healthcheck is reachable during init.
	// Readiness probe returns not-ready until init completes.
	if err := app.initAdminServer(ctx); err != nil {
		return nil, fmt.Errorf("init admin server: %w", err)
	}
	app.runAdminServer(ctx)

	if err := app.init(ctx); err != nil {
		return nil, fmt.Errorf("init app: %w", err)
	}
	return app, nil
}

func (a *App) init(ctx context.Context) error {
	// Order matters: servers before conns before controllers
	for _, f := range []func(context.Context) error{
		a.initMainServer,
		a.initGrpcConn,
		a.initControllers,
	} {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) Run(_ context.Context) {
	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Instance().GrpcServer.Port))
		if err != nil {
			slog.Error("grpc listen failed", "err", err)
			a.publicCloser.CloseAll()
			return
		}
		if err := a.grpcServer.Serve(lis); err != nil {
			slog.Error("grpc server stopped", "err", err)
			a.publicCloser.CloseAll()
		}
	}()

	// Start HTTP gateway
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http gateway stopped", "err", err)
			a.publicCloser.CloseAll()
		}
	}()

	// Mark ready — readiness probe now returns 200
	a.started.Store(1)

	cfg := config.Instance()
	slog.Info("app started",
		"grpc_port", cfg.GrpcServer.Port,
		"http_port", cfg.HttpServer.Port,
		"admin_port", cfg.HttpServer.AdminPort,
	)

	// Block until signal or server failure triggers publicCloser
	a.publicCloser.Wait()
	a.publicCloser.CloseAll()
	a.adminCloser.CloseAll()
}
