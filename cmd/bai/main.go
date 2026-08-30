package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/biisal/bai/internal/config"
)

func main() {
	configFilePath := flag.String("config", config.DefaultConfigPath(), "path to config file")
	themeConfigPath := flag.String("theme", config.ThemeConfigPath(), "path to theme config file")
	dev := flag.Bool("dev", false, "enable development mode")
	flag.Parse()
	if err := start(*configFilePath, *themeConfigPath, *dev); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}
