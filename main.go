package main

import (
	"os"
)

func main() {
	config := ServerConfig{
		Host: "localhost",
		Port: "3000",
	}

	err := Run(config)
	if err != nil {
		os.Exit(1)
	}
}
