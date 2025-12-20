package grpc

import (
	"context"

	v1 "devhive-backend/api/v1"
	"devhive-backend/internal/repo"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProjectServer implements the ProjectService gRPC service
type ProjectServer struct {
	v1.UnimplementedProjectServiceServer
	queries *repo.Queries
}

// GetProject retrieves a project by ID
func (s *ProjectServer) GetProject(ctx context.Context, req *v1.GetProjectRequest) (*v1.Project, error) {
	projectID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	project, err := s.queries.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "project not found: %v", err)
	}

	return &v1.Project{
		Id:          project.ID.String(),
		OwnerId:     project.OwnerID.String(),
		Name:        project.Name,
		Description: getStringValue(project.Description),
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}, nil
}

// CreateProject creates a new project
func (s *ProjectServer) CreateProject(ctx context.Context, req *v1.CreateProjectRequest) (*v1.Project, error) {
	// Get user ID from context (you'll need to implement this)
	userID := uuid.New() // Placeholder - should come from auth context

	project, err := s.queries.CreateProject(ctx, repo.CreateProjectParams{
		OwnerID:     userID,
		Name:        req.Name,
		Description: &req.Description,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}

	// CRITICAL: Insert owner into project_members table for consistency
	// This ensures owners appear in member lists and all queries work consistently
	err = s.queries.AddProjectMember(ctx, repo.AddProjectMemberParams{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "owner",
	})
	if err != nil {
		// Log error but don't fail the request - project was created successfully
		// The owner can still access via projects.owner_id, but member queries will be inconsistent
		// TODO: Add proper logging
	}

	return &v1.Project{
		Id:          project.ID.String(),
		OwnerId:     project.OwnerID.String(),
		Name:        project.Name,
		Description: getStringValue(project.Description),
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}, nil
}

// UpdateProject updates an existing project
func (s *ProjectServer) UpdateProject(ctx context.Context, req *v1.UpdateProjectRequest) (*v1.Project, error) {
	projectID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	project, err := s.queries.UpdateProject(ctx, repo.UpdateProjectParams{
		ID:          projectID,
		Name:        req.Name,
		Description: &req.Description,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update project: %v", err)
	}

	return &v1.Project{
		Id:          project.ID.String(),
		OwnerId:     project.OwnerID.String(),
		Name:        project.Name,
		Description: getStringValue(project.Description),
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}, nil
}

// DeleteProject deletes a project
func (s *ProjectServer) DeleteProject(ctx context.Context, req *v1.DeleteProjectRequest) (*v1.Empty, error) {
	projectID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	err = s.queries.DeleteProject(ctx, projectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete project: %v", err)
	}

	return &v1.Empty{}, nil
}

// ListProjects lists projects with pagination
func (s *ProjectServer) ListProjects(ctx context.Context, req *v1.ListProjectsRequest) (*v1.ListProjectsResponse, error) {
	// Get user ID from context (you'll need to implement this)
	userID := uuid.New() // Placeholder - should come from auth context

	projects, err := s.queries.ListProjectsByUser(ctx, repo.ListProjectsByUserParams{
		OwnerID: userID,
		Limit:   req.Limit,
		Offset:  req.Offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}

	var responseProjects []*v1.Project
	for _, project := range projects {
		responseProjects = append(responseProjects, &v1.Project{
			Id:          project.ID.String(),
			OwnerId:     project.OwnerID.String(),
			Name:        project.Name,
			Description: getStringValue(project.Description),
			CreatedAt:   timestamppb.New(project.CreatedAt),
			UpdatedAt:   timestamppb.New(project.UpdatedAt),
		})
	}

	return &v1.ListProjectsResponse{
		Projects: responseProjects,
		Total:    int32(len(responseProjects)),
	}, nil
}

// AddMember adds a member to a project
func (s *ProjectServer) AddMember(ctx context.Context, req *v1.AddMemberRequest) (*v1.Empty, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	err = s.queries.AddProjectMember(ctx, repo.AddProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
		Role:      req.Role,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add member: %v", err)
	}

	return &v1.Empty{}, nil
}

// RemoveMember removes a member from a project
func (s *ProjectServer) RemoveMember(ctx context.Context, req *v1.RemoveMemberRequest) (*v1.Empty, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	err = s.queries.RemoveProjectMember(ctx, repo.RemoveProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove member: %v", err)
	}

	return &v1.Empty{}, nil
}

// ListMembers lists project members
func (s *ProjectServer) ListMembers(ctx context.Context, req *v1.ListMembersRequest) (*v1.ListMembersResponse, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	members, err := s.queries.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list members: %v", err)
	}

	var responseMembers []*v1.ProjectMember
	for _, member := range members {
		responseMembers = append(responseMembers, &v1.ProjectMember{
			ProjectId: projectID.String(),
			UserId:    member.ID.String(),
			Role:      member.Role,
			JoinedAt:  timestamppb.New(member.JoinedAt),
		})
	}

	return &v1.ListMembersResponse{
		Members: responseMembers,
		Total:   int32(len(responseMembers)),
	}, nil
}
