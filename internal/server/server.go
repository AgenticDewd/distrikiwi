package server

import (
	"context"
	"errors"

	"github.com/harhitosw/distrikiwi/internal/engine"
	"github.com/harhitosw/distrikiwi/internal/wal"
	pb "github.com/harhitosw/distrikiwi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	pb.UnimplementedDistrikiwiServer
	engine engine.Engine
	wal    wal.WAL
}

func NewGrpcServer(eng engine.Engine, wl wal.WAL) *GrpcServer {
	return &GrpcServer{
		engine: eng,
		wal:    wl,
	}
}

func (s *GrpcServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	// Implement the Put method logic here
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "Key cannot be empty")
	}
	// 1. WAL first to achieve durability
	if err := s.wal.WriteOp(0, req.GetKey(), req.GetValue()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to write to WAL: %v", err)
	}
	// 2. Then write to the engine
	if err := s.engine.Put(req.GetKey(), req.GetValue()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to put key-value pair: %v", err)
	}
	return &pb.PutResponse{Success: true, Message: "Key stored successfully"}, nil
}

func (s *GrpcServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// Implement the Get method logic here
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "Key cannot be empty")
	}
	val, err := s.engine.Get(req.GetKey())
	if err != nil {
		if errors.Is(err, engine.ErrKeyNotFound) {
			return &pb.GetResponse{Found: false}, nil
		}
		return nil, status.Errorf(codes.Internal, "Failed to get value: %v", err)
	}

	return &pb.GetResponse{
		Value: val,
		Found: true,
	}, nil
}

func (s *GrpcServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "Key cannot be empty")
	}
	// 1. WAL first to achieve durability
	if err := s.wal.WriteOp(1, req.GetKey(), nil); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to write to WAL: %v", err)
	}
	//2. Then delete from the engine
	if err := s.engine.Delete(req.GetKey()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete key: %v", err)
	}
	return &pb.DeleteResponse{Success: true, Message: "Key deleted successfully"}, nil
}
