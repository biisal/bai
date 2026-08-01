package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/biisal/bai/internal/config"
)

func main() {
	configFilePath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	flag.Parse()
	if err := start(*configFilePath); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
