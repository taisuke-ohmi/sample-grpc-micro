package main

import (
	pb "sample-grpc-micro/proto/user"
	"sample-grpc-micro/shared/md"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"golang.org/x/net/context"
)

type UserService struct {
	store Store
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	if req.Email == "" || len(req.PasswordHash) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty email or password")
	}

	passwordHash, err := bcrypt.GenerateFromPassword(req.PasswordHash, bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.store.CreateUser(&pb.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    ptypes.TimestampNow(),
	})
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}

	ctx = md.AddUserIDToContext(ctx, user.Id)

	return &pb.CreateUserResponse{User: user}, nil
}

func (s *UserService) FindUser(ctx context.Context, req *pb.FindUserRequest) (*pb.FindUserResponse, error) {
	user, err := s.store.FindUser(req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.FindUserResponse{User: user}, nil
}

func (s *UserService) VerifyUser(ctx context.Context, req *pb.VerifyUserRequest) (*pb.VerifyUserResponse, error) {
	user, err := s.store.FindUserByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, req.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pb.VerifyUserResponse{User: user}, nil
}
