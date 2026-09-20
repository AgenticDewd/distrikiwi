package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/AgenticDewd/distrikiwi/internal/engine"
	"github.com/AgenticDewd/distrikiwi/internal/server"
	"github.com/AgenticDewd/distrikiwi/internal/wal"
	pb "github.com/AgenticDewd/distrikiwi/proto"
	"google.golang.org/grpc"
)

const (
	defaultHost = "127.0.0.1"
	port        = "50051"
)

func main() {
	port, host := os.Getenv("PORT"), os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}
	const walPath = "distrikiwi.wal"

	fmt.Println("Starting Distrikiwi gRPC server on", net.JoinHostPort(host, port))
	dbEngine := engine.NewMemEngine()

	dbWal, err := wal.NewFileWAL(walPath)
	if err != nil {
		fmt.Println("Failed to create WAL:", err)
		return
	}
	defer func() {
		if err := dbWal.Close(); err != nil {
			fmt.Println("Failed to close WAL:", err)
		}
	}()

	// replay WAL to restore state
	logs, err := dbWal.ReadWAL()
	if err != nil {
		fmt.Println("Failed to read WAL:", err)
		os.Exit(1)
	}

	if len(logs) > 0 {
		fmt.Printf("Found %d transaction logs. Replaying state...\n", len(logs))
		for _, entry := range logs {
			switch entry.OpType {
			case 0: // Put
				_ = dbEngine.Put(entry.Key, entry.Value)
			case 1: // Delete
				_ = dbEngine.Delete(entry.Key)
			}
		}
		fmt.Println("State recovery complete!")
	} else {
		fmt.Println("No logs found. Starting with a clean slate.")
	}

	// 4. Start TCP Listener
	lis, err := net.Listen("tcp", net.JoinHostPort(host, port))
	if err != nil {
		fmt.Printf("Fatal: failed to listen on port %s: %v\n", port, err)
		os.Exit(1)
	}

	// 5. Create and Register our gRPC Server
	grpcServer := grpc.NewServer()
	kvService := server.NewGrpcServer(dbEngine, dbWal)
	pb.RegisterDistrikiwiServer(grpcServer, kvService)

	// Handle graceful shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nGracefully shutting down DistriKV server...")
		grpcServer.GracefulStop()
		if err := dbWal.Close(); err != nil {
			fmt.Println("Failed to close WAL:", err)
		}
		fmt.Println("Shutdown complete. Goodbye!")
		os.Exit(0)
	}()

	fmt.Printf("DistriKV Server successfully running on Port %s...\n", port)
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Printf("Fatal: server exited with error: %v\n", err)
	}
}
