package main

func main() {
	// signal.NotifyContext: clean, modern, cancels ctx on SIGTERM/SIGINT
	// replaces manual signal channels and os.Signal handling
	// ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	// defer stop()

	// app, err := internal.New(ctx)
	// if err != nil {
	// 	slog.Error("failed to initialize app", "err", err)
	// 	os.Exit(1) // os.Exit only in main — never in library code
	// }
	// app.Run(ctx)
}
