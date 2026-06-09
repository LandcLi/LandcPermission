package main

import (
	"fmt"
	"log"

	landcpermission "github.com/LandcLi/LandcPermission"
)

func main() {
	server, err := landcpermission.InitDefault("config.yml")
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	fmt.Println("LandcPermission starting...")
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
