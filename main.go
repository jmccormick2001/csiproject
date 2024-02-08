package main

import (
	"log/slog"
	"os"

	"example.com/csiproject/service"
)

const version = "1.0"

func main() {

	slog.Info("CSI Driver is Starting")

	nodeIP := os.Getenv("NODE_IP")
	if nodeIP == "" {
		slog.Error("NODE_IP not set")
		os.Exit(1)
	}
	driverName := os.Getenv("CSI_DRIVER_NAME")
	if driverName == "" {
		slog.Error("CSI_DRIVER_NAME not set")
		os.Exit(1)
	}
	csiEndpoint := os.Getenv("CSI_ENDPOINT")
	if csiEndpoint == "" {
		slog.Error("CSI_ENDPOINT not set")
		os.Exit(1)
	}

	slog.Info("startup", "NodeIP", nodeIP)
	slog.Info("startup", "DriverName", driverName)
	slog.Info("startup", "Endpoint", csiEndpoint)

	driverOptions := service.DriverOptions{
		NodeID:     nodeIP,
		DriverName: driverName,
		Endpoint:   csiEndpoint,
		Version:    version,
	}
	d := service.NewDriver(&driverOptions)
	d.Run(false)
}
