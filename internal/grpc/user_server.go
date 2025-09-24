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

// UserServer implements the UserService gRPC service
type UserServer struct {
	v1.UnimplementedUserServiceServer
	queries *repo.Queries
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.User, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &v1.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarUrl: getStringValue(user.AvatarUrl),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.User, error) {
	// Hash password (you'll need to implement this)
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	user, err := s.queries.CreateUser(ctx, repo.CreateUserParams{
		Username:  req.Username,
		Email:     req.Email,
		PasswordH: hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &v1.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarUrl: getStringValue(user.AvatarUrl),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.User, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	user, err := s.queries.UpdateUser(ctx, repo.UpdateUserParams{
		ID:        userID,
		Email:     req.Email,
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarUrl: &req.AvatarUrl,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &v1.User{
		Id:        user.ID.String(),
		Email:     user.Email,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Active:    user.Active,
		AvatarUrl: getStringValue(user.AvatarUrl),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*v1.Empty, error) {
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID: %v", err)
	}

	err = s.queries.DeactivateUser(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &v1.Empty{}, nil
}

// ListUsers lists users with pagination
func (s *UserServer) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	users, err := s.queries.ListUsers(ctx, repo.ListUsersParams{
		Limit:  req.Limit,
		Offset: req.Offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var responseUsers []*v1.User
	for _, user := range users {
		responseUsers = append(responseUsers, &v1.User{
			Id:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Active:    user.Active,
			AvatarUrl: getStringValue(user.AvatarUrl),
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		})
	}

	return &v1.ListUsersResponse{
		Users: responseUsers,
		Total: int32(len(responseUsers)),
	}, nil
}

// Helper functions
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func hashPassword(password string) (string, error) {
	// Implement password hashing (bcrypt, scrypt, etc.)
	// For now, return the password as-is (NOT for production!)
	return password, nil
}
