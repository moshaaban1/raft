package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/mohamedshaaban/raft/config"
	"github.com/mohamedshaaban/raft/raft"
	"github.com/mohamedshaaban/raft/rpc"
)

func main() {
	cfg := loadConfig()

	node := raft.NewRaftNode(cfg)

	grpcServer := rpc.NewGRPCServer(cfg, node)

	node.Bootstrap()

	grpcServer.Start(cfg.Port, grpcServer)
}

func loadConfig() *config.Config {
	filePath := flag.String("config", "./config/config-a.yaml", "config file path")
	flag.Parse()

	slog.Info(fmt.Sprintf("Loading config file %s", *filePath))

	cfg, err := config.Load(*filePath)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
