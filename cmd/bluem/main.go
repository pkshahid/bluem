package main

import (
	"fmt"
	"os"

	"bluem/internal/cli"

	// register proxies via init()
	_ "bluem/internal/proxy"
)

func main() {
	root := cli.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}