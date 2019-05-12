package main

import (
	"context"

	pbTask "sample-grpc-micro/proto/task"
	"sample-grpc-micro/shared/md"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskService struct {
	store Store
}

func (s *TaskService) CreateTask(
	ctx context.Context,
	req *pbTask.CreateTaskRequest,
) (*pbTask.CreateTaskResponse, error) {
	if req.GetName() == "" {
		// gRPCのステータスコード付きのerrorを生成する
		return nil, status.Error(codes.InvalidArgument,
			"empty task name")
	}

	// メタデータからUserIDを取得する
	userID := md.GetUserIDFromContext(ctx)
	// protobufのTimestamp型で現在日時を取得する
	now := ptypes.TimestampNow()

	// タスクを保存する
	task, err := s.store.CreateTask(&pbTask.Task{
		Name:      req.GetName(),
		Status:    pbTask.Status_WAITING,
		UserId:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}

	return &pbTask.CreateTaskResponse{Task: task}, nil
}

func (s *TaskService) FindTasks(
	ctx context.Context,
	_ *empty.Empty,
) (*pbTask.FindTasksResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	tasks, err := s.store.FindTasks(userID)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}
	return &pbTask.FindTasksResponse{Tasks: tasks}, nil
}
