package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"runtime"
	"time"

	ad "github.com/baizhigit/dev-stand/services/api-gateway/internal/app/ad/v1"
	order "github.com/baizhigit/dev-stand/services/api-gateway/internal/app/order/v1"

	"github.com/baizhigit/dev-stand/services/api-gateway/config"
	"github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/healthcheck"

	adV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/ad/v1"
	orderV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/api_gateway/order/v1"
	extAdV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/ad_service/ad/v1"
	extOrdV1 "github.com/baizhigit/dev-stand/services/api-gateway/internal/pkg/pb/external/order_service/order/v1"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const maxGoroutines = 1000 // named constant — never magic numbers inline

func (a *App) initMainServer(ctx context.Context) error {
	cfg := config.Instance() // assign once — not called 8 times inline

	// gRPC server with keepalive and interceptors
	a.grpcServer = grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.GrpcServer.MaxConnectionIdle,
			MaxConnectionAge:  cfg.GrpcServer.MaxConnectionAge,
			Time:              cfg.GrpcServer.Time,
			Timeout:           cfg.GrpcServer.Timeout,
		}),
		grpc.ChainUnaryInterceptor(
		// Add: recoveryInterceptor, loggingInterceptor, authInterceptor
		),
	)

	// HTTP gateway mux — routes HTTP/JSON to gRPC handlers
	gwMux := gw.NewServeMux()
	grpcAddr := fmt.Sprintf(":%d", cfg.GrpcServer.Port)
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := adV1.RegisterAdServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, dialOpts); err != nil {
		return fmt.Errorf("register ad gateway: %w", err)
	}
	if err := orderV1.RegisterOrderServiceHandlerFromEndpoint(ctx, gwMux, grpcAddr, dialOpts); err != nil {
		return fmt.Errorf("register order gateway: %w", err)
	}

	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HttpServer.Port),
		Handler: gwMux,
	}

	// Register shutdown — graceful drain then hard stop
	a.publicCloser.Add(func() error {
		shutCtx, cancel := context.WithTimeout(context.Background(), cfg.Graceful.Timeout)
		defer cancel()

		stopped := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			slog.Info("grpc server gracefully stopped")
		case <-shutCtx.Done():
			a.grpcServer.Stop() // force stop after timeout
			slog.Warn("grpc server force stopped after timeout")
		}

		if err := a.httpServer.Shutdown(shutCtx); err != nil {
			return fmt.Errorf("http gateway shutdown: %w", err)
		}
		return nil
	})

	return nil
}

func (a *App) initGrpcConn(_ context.Context) error {
	cfg := config.Instance()

	for _, srv := range []string{config.AdService, config.OrderService} {
		target, ok := cfg.Targets[srv]
		if !ok {
			// Explicit check — don't silently connect to an empty address
			return fmt.Errorf("no target configured for service %q", srv)
		}

		conn, err := grpc.NewClient(target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return fmt.Errorf("init grpc conn to %s: %w", srv, err)
		}

		a.grpcConns[srv] = conn
		// Always use publicCloser — never global closer.Add()
		a.publicCloser.Add(conn.Close)
	}
	return nil
}

func (a *App) initControllers(_ context.Context) error {
	// Register gRPC service implementations — plain, no framework magic
	adV1.RegisterAdServiceServer(a.grpcServer,
		ad.New(extAdV1.NewAdServiceClient(a.grpcConns[config.AdService])),
	)
	orderV1.RegisterOrderServiceServer(a.grpcServer,
		order.New(extOrdV1.NewOrderServiceClient(a.grpcConns[config.OrderService])),
	)
	return nil
}

func (a *App) initAdminServer(ctx context.Context) error {
	cfg := config.Instance()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.HttpServer.AdminPort))
	if err != nil {
		return fmt.Errorf("admin listener: %w", err)
	}
	a.adminListener = lis

	if err = a.initHealthCheck(ctx); err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}

	a.adminMux = chi.NewMux()
	a.adminMux.Mount("/debug", chimw.Profiler())
	// Register all three probe paths — don't leave StartupPath unreachable
	a.adminMux.HandleFunc(healthcheck.LivenessPath, a.healthCheck.LiveEndpoint)
	a.adminMux.HandleFunc(healthcheck.ReadinessPath, a.healthCheck.ReadyEndpoint)
	a.adminMux.HandleFunc(healthcheck.StartupPath, a.healthCheck.StartupEndpoint)
	return nil
}

func (a *App) runAdminServer(_ context.Context) {
	adminServer := &http.Server{Handler: a.adminMux}
	go func() {
		if err := adminServer.Serve(a.adminListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("admin server stopped unexpectedly", "err", err)
			a.adminCloser.CloseAll()
		}
	}()
	a.adminCloser.Add(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		adminServer.SetKeepAlivesEnabled(false)
		if err := adminServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("admin server shutdown: %w", err)
		}
		return nil
	})
}

func (a *App) initHealthCheck(_ context.Context) error {
	a.healthCheck = healthcheck.NewHandler()

	// Liveness: is the process healthy? (not leaked, not deadlocked)
	a.healthCheck.AddLivenessCheck("goroutines", func() error {
		if n := runtime.NumGoroutine(); n >= maxGoroutines {
			return fmt.Errorf("goroutine count too high: %d", n)
		}
		return nil
	})

	// Readiness: is the process ready to serve traffic?
	a.healthCheck.AddReadinessCheck("started", func() error {
		if a.started.Load() == 0 {
			return errors.New("application not started yet")
		}
		return nil
	})
	a.healthCheck.AddReadinessCheck("terminating", func() error {
		if a.terminated.Load() != 0 {
			return errors.New("application is terminating")
		}
		return nil
	})

	// First closer func: flip termination flag so load balancer stops routing
	// before we begin draining. Must be registered before server shutdown.
	a.publicCloser.Add(func() error {
		slog.Warn("termination signal received",
			"graceful_timeout", config.Instance().Graceful.Timeout,
		)
		a.terminated.Store(1)
		return nil
	})

	return nil
}
