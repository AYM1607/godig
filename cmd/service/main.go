package main

import (
	"context"
	"flag"
	"log"

	"github.com/AYM1607/godig/pkg/auth"
	"github.com/AYM1607/godig/pkg/tunnel"
	"github.com/AYM1607/godig/types"
)

func main() {
	var (
		serverAddr    = flag.String("server", "localhost:8080", "Tunnel server address")
		localAddr     = flag.String("local", "localhost:3000", "Local service address")
		persistConfig = flag.Bool("persist-config", false, "Persist tunnel configuration to file")
	)
	flag.Parse()

	key, err := auth.GetServerKey()
	if err != nil {
		log.Fatalln(err)
	}

	clientConfig := types.TunnelClientConfig{
		PersistConfig: *persistConfig,
	}

	client, err := tunnel.NewTunnelClient(*serverAddr, *localAddr, key, clientConfig)
	if err != nil {
		log.Fatalln("Failed to create tunnel client:", err)
	}

	log.Printf("Starting tunnel client...")
	log.Printf("Tunnel ID: %s", client.TunnelID)
	log.Printf("Bearer token: %s", client.Bearer)
	log.Printf("Local service: %s", *localAddr)
	log.Printf("Server: %s", *serverAddr)

	client.Run(context.Background())
}
