package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mdp/qrterminal"

	"github.com/AYM1607/godig/pkg/auth"
	"github.com/AYM1607/godig/pkg/tunnel"
	"github.com/AYM1607/godig/types"
)

func main() {
	var (
		serverAddr    = flag.String("server", "godig.xyz:8080", "Tunnel server address")
		localAddr     = flag.String("local", "localhost:3000", "Local service address")
		persistConfig = flag.Bool("persist-config", false, "Persist tunnel configuration to file")
		generateQR    = flag.Bool("generate-qr", false, "generate qr code")
		disableAuth   = flag.Bool("disable-auth", false, "Disable bearer token authentication (insecure)")
	)
	flag.Parse()

	key, err := auth.GetServerKey()
	if err != nil {
		log.Fatalln(err)
	}

	clientConfig := types.TunnelClientConfig{
		PersistConfig: *persistConfig,
		DisableAuth:   *disableAuth,
	}

	client, err := tunnel.NewTunnelClient(*serverAddr, *localAddr, key, clientConfig)
	if err != nil {
		log.Fatalln("Failed to create tunnel client:", err)
	}

	log.Printf("Starting tunnel client...")
	log.Printf("Tunnel ID: %s", client.TunnelID)
	if client.Bearer != nil {
		log.Printf("Bearer token: %s", *client.Bearer)
	} else {
		log.Printf("Authentication: DISABLED (tunnel is publicly accessible)")
	}
	log.Printf("Local service: %s", *localAddr)
	log.Printf("Server: %s", *serverAddr)

	if *generateQR {
		bearerStr := ""
		if client.Bearer != nil {
			bearerStr = *client.Bearer
		}
		qrterminal.GenerateHalfBlock(
			fmt.Sprintf(
				`{"link": "%s", "auth": "%s"}`,
				fmt.Sprintf("https://%s.godig.xyz", client.TunnelID),
				bearerStr,
			),
			qrterminal.L,
			os.Stdout,
		)
	}

	client.Run(context.Background())
}
