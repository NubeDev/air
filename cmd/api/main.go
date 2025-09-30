package main

import (
	"flag"

	"github.com/NubeDev/air/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	dataDir      = flag.String("data", "data", "Path to data directory containing config files")
	configFile   = flag.String("config", "config.yaml", "Configuration file name (relative to data dir)")
	authDisabled = flag.Bool("auth", false, "Disable authentication (development only)")
)

func main() {
	flag.Parse()

	// Build full config path
	configPath := *dataDir + "/" + *configFile

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal().Err(err).Str("config_path", configPath).Msg("Failed to load configuration")
	}

	// Override auth setting if flag is provided
	if *authDisabled {
		cfg.Server.Auth.Enabled = false
		log.Info().Msg("Authentication disabled via --auth flag")
	}

	// Create and start server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}
	defer server.Close()

	// Start server
	if err := server.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
