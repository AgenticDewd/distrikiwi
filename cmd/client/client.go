package main

import (
	"context"
	"fmt"
	"os"
	"time"

	pb "github.com/AgenticDewd/distrikiwi/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	const address = "localhost:50051"

	// 1. Establish connection to the gRPC database server
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Did not connect: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Printf("Failed to close connection: %v\n", err)
		}
	}()

	// 2. Initialize the generated client
	client := pb.NewDistrikiwiClient(conn)

	// Create a context with a timeout for safe network calls
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fmt.Println("--- Step 1: Putting Data over network ---")
	putRes, err := client.Put(ctx, &pb.PutRequest{
		Key:   "user:99",
		Value: []byte("Developer Bob"),
	})
	if err != nil {
		fmt.Printf("Error Put: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Server Response: %s\n\n", putRes.GetMessage())

	fmt.Println("--- Step 2: Retrieving Data over network ---")
	getRes, err := client.Get(ctx, &pb.GetRequest{
		Key: "user:99",
	})
	if err != nil {
		fmt.Printf("Error Get: %v\n", err)
		os.Exit(1)
	}
	if getRes.GetFound() {
		fmt.Printf("Found Key! Value: %s\n\n", string(getRes.GetValue()))
	} else {
		fmt.Println("Key not found in database.")
	}

	fmt.Println("--- Step 3: Checking missing key ---")
	missingRes, err := client.Get(ctx, &pb.GetRequest{
		Key: "user:missing",
	})
	if err != nil {
		fmt.Printf("Error Get: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Key 'user:missing' found? %v\n", missingRes.GetFound())
}
