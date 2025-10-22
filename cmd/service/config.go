package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AYM1607/godig/pkg/config"
)

func handleConfigCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: godig-service config <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  set <key> <value>  Set a configuration value")
		fmt.Println("  get <key>          Get a configuration value")
		fmt.Println("\nKeys:")
		fmt.Printf("  %s      API key for server authentication\n", config.KeyAPIKey)
		fmt.Printf("  %s        Server address (e.g., godig.xyz:8080)\n", config.KeyServer)
		os.Exit(1)
	}

	command := os.Args[2]

	switch command {
	case "set":
		if len(os.Args) < 5 {
			fmt.Println("Usage: godig-service config set <key> <value>")
			os.Exit(1)
		}
		key := config.ConfigKey(os.Args[3])
		value := os.Args[4]

		if err := config.SetConfigValue(key, value); err != nil {
			log.Fatalf("Failed to set config: %v\n", err)
		}

		fmt.Printf("Configuration updated: %s = %s\n", key, value)

	case "get":
		if len(os.Args) < 4 {
			fmt.Println("Usage: godig-service config get <key>")
			os.Exit(1)
		}
		key := config.ConfigKey(os.Args[3])

		value, err := config.GetConfigValue(key)
		if err != nil {
			log.Fatalf("Failed to get config: %v\n", err)
		}

		if value == "" {
			fmt.Printf("%s is not set\n", key)
		} else {
			fmt.Printf("%s = %s\n", key, value)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Valid commands: set, get")
		os.Exit(1)
	}
}
