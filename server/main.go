package main

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "grpc-homework/proto"
)

type server struct {
	pb.UnimplementedEntityServiceServer
	store *Storage
}

func (s *server) CreateEntity(ctx context.Context, req *pb.CreateEntityRequest) (*pb.EntityResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	e, err := s.store.Create(req.GetName(), req.GetContent())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create: %v", err)
	}
	return &pb.EntityResponse{Entity: toProto(e)}, nil
}

func (s *server) GetEntity(ctx context.Context, req *pb.GetEntityRequest) (*pb.EntityResponse, error) {
	e, err := s.store.Get(req.GetId())
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, "entity not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get: %v", err)
	}
	return &pb.EntityResponse{Entity: toProto(e)}, nil
}

func (s *server) UpdateEntity(ctx context.Context, req *pb.UpdateEntityRequest) (*pb.EntityResponse, error) {
	e, err := s.store.Update(req.GetId(), req.GetName(), req.GetContent())
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, "entity not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update: %v", err)
	}
	return &pb.EntityResponse{Entity: toProto(e)}, nil
}

func (s *server) DeleteEntity(ctx context.Context, req *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error) {
	err := s.store.Delete(req.GetId())
	if errors.Is(err, ErrNotFound) {
		return nil, status.Error(codes.NotFound, "entity not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete: %v", err)
	}
	return &pb.DeleteEntityResponse{Success: true}, nil
}

func (s *server) ListEntities(ctx context.Context, req *pb.ListEntitiesRequest) (*pb.ListEntitiesResponse, error) {
	items, total := s.store.List(req.GetPage(), req.GetPageSize())
	out := make([]*pb.Entity, 0, len(items))
	for _, e := range items {
		out = append(out, toProto(e))
	}
	return &pb.ListEntitiesResponse{Entities: out, Total: total}, nil
}

func toProto(e *Entity) *pb.Entity {
	return &pb.Entity{Id: e.ID, Name: e.Name, Content: e.Content, Version: e.Version}
}

func main() {
	store, err := NewStorage("entities.csv")
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEntityServiceServer(grpcServer, &server{store: store})

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}