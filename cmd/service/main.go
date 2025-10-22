package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mdp/qrterminal"

	"github.com/AYM1607/godig/pkg/config"
	"github.com/AYM1607/godig/pkg/tunnel"
	"github.com/AYM1607/godig/types"
)

func main() {
	// Check for config subcommand.
	if len(os.Args) > 1 && os.Args[1] == "config" {
		handleConfigCommand()
		return
	}

	var (
		serverAddrFlag = flag.String("server", "", "Tunnel server address")
		apiKeyFlag     = flag.String("api-key", "", "API key for server authentication")
		localAddr      = flag.String("local", "localhost:3000", "Local service address")
		persistConfig  = flag.Bool("persist-config", false, "Persist tunnel configuration to file")
		generateQR     = flag.Bool("generate-qr", false, "generate qr code")
		disableAuth    = flag.Bool("disable-auth", false, "Disable bearer token authentication (insecure)")
	)
	flag.Parse()

	// Load global config.
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		log.Fatalf("Failed to load global config: %v\n", err)
	}

	// Resolve API key with priority: CLI flag > env var > global config.
	var apiKey string
	if *apiKeyFlag != "" {
		apiKey = *apiKeyFlag
	} else if envKey := os.Getenv("GODIG_API_KEY"); envKey != "" {
		apiKey = envKey
	} else if globalConfig.APIKey != "" {
		apiKey = globalConfig.APIKey
	} else {
		log.Fatalln("API key must be provided via --api-key flag, GODIG_API_KEY environment variable, or global config")
	}

	// Resolve server address with priority: CLI flag > env var > global config > default.
	serverAddr := "godig.xyz:8080"
	if *serverAddrFlag != "" {
		serverAddr = *serverAddrFlag
	} else if envServer := os.Getenv("GODIG_SERVER"); envServer != "" {
		serverAddr = envServer
	} else if globalConfig.Server != "" {
		serverAddr = globalConfig.Server
	}

	clientConfig := types.TunnelClientConfig{
		PersistConfig: *persistConfig,
		DisableAuth:   *disableAuth,
	}

	client, err := tunnel.NewTunnelClient(serverAddr, *localAddr, apiKey, clientConfig)
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
	log.Printf("Server: %s", serverAddr)

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
