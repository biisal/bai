package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/biisal/bai/internal/agent"
	"github.com/biisal/bai/internal/config"
	"github.com/biisal/bai/internal/git"
	"github.com/biisal/bai/internal/logger"
	"github.com/biisal/bai/internal/notifier"
	broker "github.com/biisal/bai/internal/pubsub"
	webui "github.com/biisal/bai/internal/web"
	bundle "github.com/biisal/bai/web/app"
)

func runWeb(cfg *config.Config, dev bool, args []string) int {
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:7777", "listen address (port increments if busy)")
	devFlag := fs.Bool("dev", false, "allow the Vite dev origin")
	noOpen := fs.Bool("no-open", false, "do not open a browser")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	dev = dev || *devFlag

	logLevel := slog.LevelInfo
	if dev {
		logLevel = slog.LevelDebug
	}
	logFile, err := logger.SetUpLogger(cfg.LogFilePath, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up logger: %v\n", err)
		return 1
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error("close log file", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbService, err := startDB(ctx, cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	n := notifier.New(nil)
	defer n.Close(context.Background())

	skillPaths := append([]string{}, cfg.SkillsPaths...)
	skillPaths = append(skillPaths, config.DefaultSkillsPaths()...)
	gitRepo := git.New(cfg.GitDirName)
	b := broker.New()
	gateway, err := agent.NewGateway(ctx, dbService, b, cfg, n, skillPaths, gitRepo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	if initialized, err := gitRepo.CheckIfGitInitialized(); err != nil {
		slog.Warn("could not check shadow git repo", "error", err)
	} else if !initialized {
		slog.Info("shadow git repo is not initialized; runs will continue without checkpoints", "dir", gitRepo.Dir())
	}

	token := os.Getenv("BAI_WEB_TOKEN")
	if token == "" {
		var err error
		token, err = randomToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate web token: %v\n", err)
			return 1
		}
	}

	ln, listenAddr, err := listen(*addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen on %s: %v\n", *addr, err)
		return 1
	}
	if !isLoopback(listenAddr) {
		fmt.Fprintf(os.Stderr, "WARNING: bai web is listening on a non-loopback address %s. Anyone who can reach it and has the token can run commands as you.\n", listenAddr)
	}

	static, err := bundle.Dist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load web assets: %v\n", err)
		return 1
	}
	registry := webui.NewRegistry(ctx)
	server := webui.NewServer(webui.Options{
		Gateway:  gateway,
		Config:   cfg,
		Registry: registry,
		Token:    token,
		Addr:     listenAddr,
		Dev:      dev,
		Static:   static,
	})
	httpServer := &http.Server{
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	events := b.Subscribe()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-events:
			}
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(ln)
	}()

	url := fmt.Sprintf("http://%s/?t=%s", urlHost(listenAddr), token)
	fmt.Printf("bai web: %s\n", url)
	if !*noOpen {
		go openBrowser(url)
	}

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "web server: %v\n", err)
			return 1
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	registry.Shutdown(shutdownCtx)
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown web server", "error", err)
	}
	return 0
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func urlHost(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "0.0.0.0" || host == "::" || host == "" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

// listen binds the requested port, walking forward if it is already taken so a
// second instance does not fail just because the first one is running.
func listen(addr string) (net.Listener, string, error) {
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, "", err
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 0 || port > 65535 {
		return nil, "", fmt.Errorf("invalid port %q", portText)
	}
	if port == 0 {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			return nil, "", err
		}
		return ln, ln.Addr().String(), nil
	}
	for p := port; p < port+100 && p <= 65535; p++ {
		candidate := net.JoinHostPort(host, strconv.Itoa(p))
		if ln, err := net.Listen("tcp", candidate); err == nil {
			return ln, ln.Addr().String(), nil
		}
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return nil, "", fmt.Errorf("no free port available starting from %d: %w", port, err)
	}
	return ln, ln.Addr().String(), nil
}

func openBrowser(url string) {
	if browser := os.Getenv("BROWSER"); browser != "" {
		if cmd := exec.Command(browser, url); cmd.Start() == nil {
			return
		}
	}

	var cmds []*exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmds = []*exec.Cmd{exec.Command("open", url)}
	case "windows":
		cmds = []*exec.Cmd{
			exec.Command("rundll32", "url.dll,FileProtocolHandler", url),
			exec.Command("cmd.exe", "/c", "start", "", url),
		}
	default:
		if isWSL() {
			time.Sleep(500 * time.Millisecond)
			cmds = append(cmds,
				exec.Command("wslview", url),
				exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Start-Process", fmt.Sprintf(`"%s"`, url)),
				exec.Command("cmd.exe", "/c", "start", "", url),
			)
		}
		cmds = append(cmds,
			exec.Command("xdg-open", url),
			exec.Command("sensible-browser", url),
			exec.Command("gio", "open", url),
			exec.Command("google-chrome", url),
			exec.Command("firefox", url),
			exec.Command("chromium", url),
		)
	}
	for _, cmd := range cmds {
		if cmd.Start() == nil {
			return
		}
	}
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	data, err := os.ReadFile("/proc/version")
	return err == nil && strings.Contains(strings.ToLower(string(data)), "microsoft")
}
