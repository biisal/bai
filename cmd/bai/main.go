package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/git"
)

func main() {
	os.Exit(run())
}

func run() int {
	configFilePath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	dev := flag.Bool("dev", false, "enable development mode")
	flag.Parse()

	cfg, err := config.Load(*configFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		return 1
	}

	if args := flag.Args(); len(args) > 0 && args[0] == "git" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		err := git.New(cfg.GitDirName).Passthrough(ctx, args[1:]...)

		var exitErr *exec.ExitError
		switch {
		case err == nil:
			return 0
		case errors.As(err, &exitErr):
			return exitErr.ExitCode()
		default:
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
	}

	if err := start(cfg, *dev); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
		return 1
	}
	return 0
}
