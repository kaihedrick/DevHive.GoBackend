package grpc

import (
	"context"

	v1 "devhive-backend/api/v1"
	"devhive-backend/internal/repo"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskServer implements the TaskService gRPC service
type TaskServer struct {
	v1.UnimplementedTaskServiceServer
	queries *repo.Queries
}

// GetTask retrieves a task by ID
func (s *TaskServer) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.Task, error) {
	taskID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	task, err := s.queries.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}

	var sprintID, assigneeID string
	if task.SprintID.Valid {
		sprintID = uuid.UUID(task.SprintID.Bytes).String()
	}
	if task.AssigneeID.Valid {
		assigneeID = uuid.UUID(task.AssigneeID.Bytes).String()
	}

	return &v1.Task{
		Id:          task.ID.String(),
		ProjectId:   task.ProjectID.String(),
		SprintId:    sprintID,
		AssigneeId:  assigneeID,
		Title:       task.Title,
		Description: getStringValue(task.Description),
		Status:      int32(task.Status),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}, nil
}

// CreateTask creates a new task
func (s *TaskServer) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.Task, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	var sprintID pgtype.UUID
	if req.SprintId != "" {
		sprintUUID, err := uuid.Parse(req.SprintId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid sprint ID: %v", err)
		}
		sprintID = pgtype.UUID{Bytes: sprintUUID, Valid: true}
	}

	task, err := s.queries.CreateTask(ctx, repo.CreateTaskParams{
		ProjectID:   projectID,
		SprintID:    sprintID,
		Title:       req.Title,
		Description: &req.Description,
		Status:      1, // Default status: TODO
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}

	var sprintIDStr, assigneeID string
	if task.SprintID.Valid {
		sprintIDStr = uuid.UUID(task.SprintID.Bytes).String()
	}
	if task.AssigneeID.Valid {
		assigneeID = uuid.UUID(task.AssigneeID.Bytes).String()
	}

	return &v1.Task{
		Id:          task.ID.String(),
		ProjectId:   task.ProjectID.String(),
		SprintId:    sprintIDStr,
		AssigneeId:  assigneeID,
		Title:       task.Title,
		Description: getStringValue(task.Description),
		Status:      int32(task.Status),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}, nil
}

// UpdateTask updates an existing task
func (s *TaskServer) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.Task, error) {
	taskID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	task, err := s.queries.UpdateTask(ctx, repo.UpdateTaskParams{
		ID:          taskID,
		Title:       req.Title,
		Description: &req.Description,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	var sprintID, assigneeID string
	if task.SprintID.Valid {
		sprintID = uuid.UUID(task.SprintID.Bytes).String()
	}
	if task.AssigneeID.Valid {
		assigneeID = uuid.UUID(task.AssigneeID.Bytes).String()
	}

	return &v1.Task{
		Id:          task.ID.String(),
		ProjectId:   task.ProjectID.String(),
		SprintId:    sprintID,
		AssigneeId:  assigneeID,
		Title:       task.Title,
		Description: getStringValue(task.Description),
		Status:      int32(task.Status),
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}, nil
}

// DeleteTask deletes a task
func (s *TaskServer) DeleteTask(ctx context.Context, req *v1.DeleteTaskRequest) (*v1.Empty, error) {
	taskID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	err = s.queries.DeleteTask(ctx, taskID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	return &v1.Empty{}, nil
}

// ListTasks lists tasks with filters
func (s *TaskServer) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksResponse, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	tasks, err := s.queries.ListTasksByProject(ctx, repo.ListTasksByProjectParams{
		ProjectID: projectID,
		Limit:     req.Limit,
		Offset:    req.Offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	var responseTasks []*v1.Task
	for _, task := range tasks {
		var sprintID, assigneeID string
		if task.SprintID.Valid {
			sprintID = uuid.UUID(task.SprintID.Bytes).String()
		}
		if task.AssigneeID.Valid {
			assigneeID = uuid.UUID(task.AssigneeID.Bytes).String()
		}

		responseTasks = append(responseTasks, &v1.Task{
			Id:          task.ID.String(),
			ProjectId:   task.ProjectID.String(),
			SprintId:    sprintID,
			AssigneeId:  assigneeID,
			Title:       task.Title,
			Description: getStringValue(task.Description),
			Status:      int32(task.Status),
			CreatedAt:   timestamppb.New(task.CreatedAt),
			UpdatedAt:   timestamppb.New(task.UpdatedAt),
		})
	}

	return &v1.ListTasksResponse{
		Tasks: responseTasks,
		Total: int32(len(responseTasks)),
	}, nil
}

// AssignTask assigns a task to a user
func (s *TaskServer) AssignTask(ctx context.Context, req *v1.AssignTaskRequest) (*v1.Empty, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	assigneeID, err := uuid.Parse(req.AssigneeId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid assignee ID: %v", err)
	}

	assigneeUUID := pgtype.UUID{Bytes: assigneeID, Valid: true}
	_, err = s.queries.UpdateTask(ctx, repo.UpdateTaskParams{
		ID:         taskID,
		AssigneeID: assigneeUUID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign task: %v", err)
	}

	return &v1.Empty{}, nil
}

// UpdateTaskStatus updates the status of a task
func (s *TaskServer) UpdateTaskStatus(ctx context.Context, req *v1.UpdateTaskStatusRequest) (*v1.Empty, error) {
	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task ID: %v", err)
	}

	_, err = s.queries.UpdateTaskStatus(ctx, repo.UpdateTaskStatusParams{
		ID:     taskID,
		Status: int32(req.Status),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task status: %v", err)
	}

	return &v1.Empty{}, nil
}
