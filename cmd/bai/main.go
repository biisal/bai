package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/biisal/bai/internal/config"
)

func init() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}

func main() {
	configFilePath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	dev := flag.Bool("dev", false, "enable development mode")
	flag.Parse()
	if err := start(*configFilePath, *dev); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
