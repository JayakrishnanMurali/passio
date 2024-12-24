package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jayakrishnanMurali/passio/internal/app"
	"github.com/jayakrishnanMurali/passio/internal/cmd"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	app, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	cleanup := func() {
		if err := app.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error during cleanup: %v\n", err)
		}
	}

	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		cleanup()
		os.Exit(0)
	}()

	defer cleanup()

	rootCmd := cmd.NewRootCmd(app)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
