package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/biisal/bai/internal/config"
)

func main() {
	configFilePath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	dev := flag.Bool("dev", false, "enable development mode")
	flag.Parse()
	if err := start(*configFilePath, *dev); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
