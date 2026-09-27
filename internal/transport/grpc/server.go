// internal/transport/grpc/server.go
package grpc

import (
    "context"
    "errors"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "grpc-users/proto/userpb"
    "grpc-users/internal/domain"
    "grpc-users/internal/service"
)

type Server struct {
    pb.UnimplementedUserServiceServer
    svc *service.UserService
}

func NewServer(svc *service.UserService) *Server {
    return &Server{svc: svc}
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
    u, err := s.svc.CreateUser(ctx, req.GetName(), req.GetEmail(), int(req.GetAge()))
    if err != nil {
        return nil, toStatus(err)
    }
    return toProto(u), nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
    u, err := s.svc.GetUser(ctx, req.GetId())
    if err != nil {
        return nil, toStatus(err)
    }
    return toProto(u), nil
}

// ... ListUsers, UpdateUser, DeleteUser аналогично

func toStatus(err error) error {
    switch {
    case errors.Is(err, domain.ErrUserNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, domain.ErrEmailExists):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, domain.ErrInvalidName),
        errors.Is(err, domain.ErrInvalidEmail),
        errors.Is(err, domain.ErrInvalidAge):
        return status.Error(codes.InvalidArgument, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}