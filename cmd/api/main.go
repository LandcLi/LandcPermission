package main

import (
	"fmt"
	"log"

	landcpermission "github.com/LandcLi/LandcPermission"
)

func main() {
	engine, err := landcpermission.InitDefault("config.yml")
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	port := ":8082"
	fmt.Printf("LandcPermission starting on %s\n", port)
	if err := engine.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
