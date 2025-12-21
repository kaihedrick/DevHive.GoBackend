# Project Creation & Invite System - Complete Logic Guide

This document shows the complete flow for creating projects, managing invites, and authorization logic.

---

## 1. Creating a New Project

### Endpoint
**POST** `/api/v1/projects`

### Handler: `CreateProject`

```go
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user ID
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w, "User ID not found in context")
        return
    }

    // 2. Parse request body
    var req CreateProjectRequest
    if !response.Decode(w, r, &req) {
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        response.BadRequest(w, "Invalid user ID")
        return
    }

    // 3. Create project in database
    project, err := h.queries.CreateProject(r.Context(), repo.CreateProjectParams{
        OwnerID:     userUUID,
        Name:        req.Name,
        Description: &req.Description,
    })
    if err != nil {
        response.BadRequest(w, "Failed to create project: "+err.Error())
        return
    }

    // 4. CRITICAL: Insert owner into project_members table (canonical model)
    // This ensures owners appear in member lists and all queries work consistently
    err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
        ProjectID: project.ID,
        UserID:    userUUID,
        Role:      "owner",
    })
    if err != nil {
        // Log error but don't fail the request - project was created successfully
        log.Printf("Warning: Failed to add owner to project_members for project %s: %v", 
            project.ID.String(), err)
    }

    // 5. Return project response
    response.JSON(w, http.StatusCreated, ProjectResponse{
        ID:          project.ID.String(),
        OwnerID:     project.OwnerID.String(),
        Name:        project.Name,
        Description: *project.Description,
        CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
    })
}
```

### Request Body
```json
{
  "name": "My Project",
  "description": "Project description"
}
```

### Response (201 Created)
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "name": "My Project",
  "description": "Project description",
  "createdAt": "2025-01-20T15:30:00Z",
  "updatedAt": "2025-01-20T15:30:00Z"
}
```

### Key Points
- ✅ User must be authenticated
- ✅ Creator automatically becomes project owner
- ✅ Owner is **automatically added to `project_members`** table (canonical model)
- ✅ This ensures owner appears in member lists and can access all project features

---

## 2. Creating an Invite Link

### Endpoint
**POST** `/api/v1/projects/{projectId}/invites`

### Handler: `CreateInvite`

```go
func (h *ProjectHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user ID
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w, "User ID not found in context")
        return
    }

    // 2. Get project ID from URL
    projectID := chi.URLParam(r, "projectId")
    if projectID == "" {
        response.BadRequest(w, "Project ID is required")
        return
    }

    projectUUID, err := uuid.Parse(projectID)
    if err != nil {
        response.BadRequest(w, "Invalid project ID")
        return
    }
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        response.BadRequest(w, "Invalid user ID")
        return
    }

    // 3. AUTHORIZATION: Check if user is owner or admin
    isOwnerOrAdmin, err := h.queries.CheckProjectOwnerOrAdmin(r.Context(), 
        repo.CheckProjectOwnerOrAdminParams{
            ID:      projectUUID,
            OwnerID: userUUID,
        })
    if err != nil || !isOwnerOrAdmin {
        response.Forbidden(w, "Only project owners and admins can create invites")
        return
    }

    // 4. Parse request body (optional fields)
    var req CreateInviteRequest
    if !response.Decode(w, r, &req) {
        return
    }

    // 5. Generate unique invite token (UUID)
    inviteToken, err := uuid.NewRandom()
    if err != nil {
        response.InternalServerError(w, "Failed to generate invite token")
        return
    }

    // 6. Set expiration time (default 30 minutes)
    expiresInMinutes := 30
    if req.ExpiresInMinutes != nil && *req.ExpiresInMinutes > 0 {
        expiresInMinutes = *req.ExpiresInMinutes
    }
    expiresAt := time.Now().Add(time.Duration(expiresInMinutes) * time.Minute)

    // 7. Create invite in database
    invite, err := h.queries.CreateProjectInvite(r.Context(), repo.CreateProjectInviteParams{
        ProjectID:   projectUUID,
        CreatedBy:   userUUID,
        InviteToken: inviteToken.String(),
        ExpiresAt:   expiresAt,
        MaxUses:     req.MaxUses, // Optional: nil = unlimited
    })
    if err != nil {
        response.BadRequest(w, "Failed to create invite: "+err.Error())
        return
    }

    // 8. Return invite details
    response.JSON(w, http.StatusCreated, map[string]interface{}{
        "id":        invite.ID.String(),
        "projectId": invite.ProjectID.String(),
        "token":     invite.InviteToken,
        "expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
        "maxUses":   invite.MaxUses,
        "usedCount": invite.UsedCount,
        "isActive":  invite.IsActive,
        "createdAt": invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
    })
}
```

### Request Body (All fields optional)
```json
{
  "expiresInMinutes": 60,  // Optional, defaults to 30
  "maxUses": 10            // Optional, null = unlimited
}
```

### Response (201 Created)
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "token": "invite-token-uuid",
  "expiresAt": "2025-01-20T16:30:00Z",
  "maxUses": 10,
  "usedCount": 0,
  "isActive": true,
  "createdAt": "2025-01-20T15:30:00Z"
}
```

### Authorization Rules
- ✅ **Only owners and admins** can create invites
- ❌ Regular members cannot create invites
- ✅ Uses `CheckProjectOwnerOrAdmin` query to verify permissions

---

## 3. Listing Invite Links (Viewing Invites)

### Endpoint
**GET** `/api/v1/projects/{projectId}/invites`

### Handler: `ListInvites`

```go
func (h *ProjectHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user ID
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w, "User ID not found in context")
        return
    }

    // 2. Get project ID from URL
    projectID := chi.URLParam(r, "projectId")
    if projectID == "" {
        response.BadRequest(w, "Project ID is required")
        return
    }

    projectUUID, err := uuid.Parse(projectID)
    if err != nil {
        response.BadRequest(w, "Invalid project ID")
        return
    }
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        response.BadRequest(w, "Invalid user ID")
        return
    }

    // 3. AUTHORIZATION: Check if user has access to project
    // ANY project member (owner, admin, or regular member) can view invites
    hasAccess, err := h.queries.CheckProjectAccess(r.Context(), 
        repo.CheckProjectAccessParams{
            ProjectID: projectUUID,
            UserID:    userUUID,
        })
    if err != nil || !hasAccess {
        response.Forbidden(w, "Access denied to project")
        return
    }

    // 4. Get all active invites for the project
    invites, err := h.queries.ListProjectInvites(r.Context(), projectUUID)
    if err != nil {
        response.BadRequest(w, "Failed to list invites: "+err.Error())
        return
    }

    // 5. Convert to response format
    var inviteResponses []map[string]interface{}
    for _, invite := range invites {
        inviteResponses = append(inviteResponses, map[string]interface{}{
            "id":        invite.ID.String(),
            "projectId": invite.ProjectID.String(),
            "token":     invite.InviteToken,
            "expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
            "maxUses":   invite.MaxUses,
            "usedCount": invite.UsedCount,
            "isActive":  invite.IsActive,
            "createdAt": invite.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        })
    }

    // 6. Ensure invites is always an array, never null
    if inviteResponses == nil {
        inviteResponses = []map[string]interface{}{}
    }

    // 7. Return invites list
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "invites": inviteResponses,
        "count":   len(inviteResponses),
    })
}
```

### Response (200 OK)
```json
{
  "invites": [
    {
      "id": "uuid",
      "projectId": "uuid",
      "token": "invite-token-uuid",
      "expiresAt": "2025-01-20T16:30:00Z",
      "maxUses": 10,
      "usedCount": 3,
      "isActive": true,
      "createdAt": "2025-01-20T15:30:00Z"
    }
  ],
  "count": 1
}
```

### Authorization Rules
- ✅ **All project members** (owners, admins, and regular members) can **view** invites
- ✅ Uses `CheckProjectAccess` query (checks `project_members` table)
- ✅ Returns empty array if no invites exist (never null)

---

## 4. Revoking an Invite Link

### Endpoint
**DELETE** `/api/v1/projects/{projectId}/invites/{inviteId}`

### Handler: `RevokeInvite`

```go
func (h *ProjectHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user ID
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w, "User ID not found in context")
        return
    }

    // 2. Get project ID and invite ID from URL
    projectID := chi.URLParam(r, "projectId")
    inviteID := chi.URLParam(r, "inviteId")
    if projectID == "" || inviteID == "" {
        response.BadRequest(w, "Project ID and Invite ID are required")
        return
    }

    projectUUID, err := uuid.Parse(projectID)
    if err != nil {
        response.BadRequest(w, "Invalid project ID")
        return
    }
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        response.BadRequest(w, "Invalid user ID")
        return
    }

    // 3. AUTHORIZATION: Check if user is owner or admin
    isOwnerOrAdmin, err := h.queries.CheckProjectOwnerOrAdmin(r.Context(), 
        repo.CheckProjectOwnerOrAdminParams{
            ID:      projectUUID,
            OwnerID: userUUID,
        })
    if err != nil || !isOwnerOrAdmin {
        response.Forbidden(w, "Only project owners and admins can revoke invites")
        return
    }

    // 4. Verify invite exists and belongs to this project
    inviteUUID, err := uuid.Parse(inviteID)
    if err != nil {
        response.BadRequest(w, "Invalid invite ID")
        return
    }

    invite, err := h.queries.GetProjectInviteByID(r.Context(), inviteUUID)
    if err != nil {
        response.NotFound(w, "Invite not found")
        return
    }

    if invite.ProjectID != projectUUID {
        response.Forbidden(w, "Invite does not belong to this project")
        return
    }

    // 5. Deactivate the invite (sets is_active = false)
    err = h.queries.DeactivateInvite(r.Context(), inviteUUID)
    if err != nil {
        response.BadRequest(w, "Failed to revoke invite: "+err.Error())
        return
    }

    // 6. Return success
    response.JSON(w, http.StatusOK, map[string]string{
        "message": "Invite revoked successfully"
    })
}
```

### Authorization Rules
- ✅ **Only owners and admins** can revoke invites
- ❌ Regular members cannot revoke invites
- ✅ Verifies invite belongs to the project before revoking

---

## 5. Getting Invite Details (Public Endpoint)

### Endpoint
**GET** `/api/v1/invites/{inviteToken}`

### Handler: `GetInviteDetails`

```go
func (h *ProjectHandler) GetInviteDetails(w http.ResponseWriter, r *http.Request) {
    // 1. Get invite token from URL (NO AUTH REQUIRED - public endpoint)
    inviteToken := chi.URLParam(r, "inviteToken")
    if inviteToken == "" {
        response.BadRequest(w, "Invite token is required")
        return
    }

    // 2. Get invite from database
    invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
    if err != nil {
        response.NotFound(w, "Invite not found or expired")
        return
    }

    // 3. Validate invite is active
    if !invite.IsActive {
        response.BadRequest(w, "Invite has been revoked")
        return
    }

    // 4. Check if invite has expired
    if invite.ExpiresAt.Before(time.Now()) {
        response.BadRequest(w, "Invite has expired")
        return
    }

    // 5. Get project details
    project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
    if err != nil {
        response.NotFound(w, "Project not found")
        return
    }

    // 6. Return invite and project details
    response.JSON(w, http.StatusOK, map[string]interface{}{
        "invite": map[string]interface{}{
            "id":        invite.ID.String(),
            "projectId": invite.ProjectID.String(),
            "token":     invite.InviteToken,
            "expiresAt": invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
            "maxUses":   invite.MaxUses,
            "usedCount": invite.UsedCount,
            "isActive":  invite.IsActive,
        },
        "project": map[string]interface{}{
            "id":          project.ID.String(),
            "name":        project.Name,
            "description": *project.Description,
        },
    })
}
```

### Response (200 OK)
```json
{
  "invite": {
    "id": "uuid",
    "projectId": "uuid",
    "token": "invite-token-uuid",
    "expiresAt": "2025-01-20T16:30:00Z",
    "maxUses": 10,
    "usedCount": 3,
    "isActive": true
  },
  "project": {
    "id": "uuid",
    "name": "My Project",
    "description": "Project description"
  }
}
```

### Key Points
- ✅ **No authentication required** (public endpoint)
- ✅ Validates invite is active and not expired
- ✅ Returns project details so user can see what they're joining

---

## 6. Accepting an Invite

### Endpoint
**POST** `/api/v1/invites/{inviteToken}/accept`

### Handler: `AcceptInvite`

```go
func (h *ProjectHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
    // 1. Get authenticated user ID
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Unauthorized(w, "User ID not found in context")
        return
    }

    // 2. Get invite token from URL
    inviteToken := chi.URLParam(r, "inviteToken")
    if inviteToken == "" {
        response.BadRequest(w, "Invite token is required")
        return
    }

    // 3. Get invite from database
    invite, err := h.queries.GetProjectInviteByToken(r.Context(), inviteToken)
    if err != nil {
        response.NotFound(w, "Invite not found or expired")
        return
    }

    // 4. Validate invite is active
    if !invite.IsActive {
        response.BadRequest(w, "Invite has been revoked")
        return
    }

    // 5. Check if invite has expired
    if invite.ExpiresAt.Before(time.Now()) {
        response.BadRequest(w, "Invite has expired")
        return
    }

    // 6. Check if invite has reached max uses
    if invite.MaxUses != nil && invite.UsedCount >= *invite.MaxUses {
        response.BadRequest(w, "Invite has reached maximum uses")
        return
    }

    userUUID, err := uuid.Parse(userID)
    if err != nil {
        response.BadRequest(w, "Invalid user ID")
        return
    }

    // 7. Add user to project_members table
    log.Printf("AcceptInvite: Adding user %s to project %s", 
        userUUID.String(), invite.ProjectID.String())
    
    err = h.queries.AddProjectMember(r.Context(), repo.AddProjectMemberParams{
        ProjectID: invite.ProjectID,
        UserID:    userUUID,
        Role:      "member", // Default role for invited members
    })
    if err != nil {
        response.BadRequest(w, "Failed to join project: "+err.Error())
        return
    }

    log.Printf("AcceptInvite: Successfully added user %s to project %s (should trigger cache invalidation)", 
        userUUID.String(), invite.ProjectID.String())

    // 8. Increment invite use count
    err = h.queries.IncrementInviteUseCount(r.Context(), invite.ID)
    if err != nil {
        // Log but don't fail - user was already added
        log.Printf("Warning: Failed to increment invite use count: %v", err)
    }

    // 9. Get project details and return
    project, err := h.queries.GetProjectByID(r.Context(), invite.ProjectID)
    if err != nil {
        response.NotFound(w, "Project not found")
        return
    }

    // Get user's role and permissions
    userRole, permissions := h.getUserRoleAndPermissions(r.Context(), invite.ProjectID, userUUID)

    response.JSON(w, http.StatusOK, ProjectResponse{
        ID:          project.ID.String(),
        OwnerID:     project.OwnerID.String(),
        Name:        project.Name,
        Description: *project.Description,
        CreatedAt:   project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        UpdatedAt:   project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
        UserRole:    userRole,
        Permissions: permissions,
    })
}
```

### Key Points
- ✅ User must be authenticated
- ✅ Validates invite is active, not expired, and hasn't reached max uses
- ✅ Adds user to `project_members` table with role "member"
- ✅ Increments invite use count
- ✅ Returns full project response with user role and permissions

---

## 7. Authorization & Permissions System

### Permission Calculation: `getUserRoleAndPermissions`

```go
func (h *ProjectHandler) getUserRoleAndPermissions(ctx context.Context, projectID, userID uuid.UUID) 
    (*string, struct {
        CanViewInvites   bool `json:"canViewInvites"`
        CanCreateInvites bool `json:"canCreateInvites"`
        CanRevokeInvites bool `json:"canRevokeInvites"`
        CanManageMembers bool `json:"canManageMembers"`
    }) {
    
    // 1. Check if user is owner
    isOwner, err := h.queries.CheckProjectOwner(ctx, repo.CheckProjectOwnerParams{
        ID:      projectID,
        OwnerID: userID,
    })
    if err != nil {
        // On error, return nil role and no permissions
        return nil, struct {
            CanViewInvites   bool `json:"canViewInvites"`
            CanCreateInvites bool `json:"canCreateInvites"`
            CanRevokeInvites bool `json:"canRevokeInvites"`
            CanManageMembers bool `json:"canManageMembers"`
        }{}
    }

    var userRole string
    var permissions struct {
        CanViewInvites   bool `json:"canViewInvites"`
        CanCreateInvites bool `json:"canCreateInvites"`
        CanRevokeInvites bool `json:"canRevokeInvites"`
        CanManageMembers bool `json:"canManageMembers"`
    }

    // 2. If owner, grant all permissions
    if isOwner {
        userRole = "owner"
        permissions.CanViewInvites = true
        permissions.CanCreateInvites = true
        permissions.CanRevokeInvites = true
        permissions.CanManageMembers = true
    } else {
        // 3. Get user's role from project_members table
        roleResult, err := h.queries.GetUserProjectRole(ctx, repo.GetUserProjectRoleParams{
            ID:      projectID,
            OwnerID: userID,
        })
        if err != nil {
            // Default to member if query fails
            userRole = "member"
        } else {
            if roleStr, ok := roleResult.(string); ok && roleStr != "" {
                userRole = roleStr
            } else {
                userRole = "member"
            }
        }

        // 4. Set permissions based on role
        switch userRole {
        case "admin":
            permissions.CanViewInvites = true
            permissions.CanCreateInvites = true
            permissions.CanRevokeInvites = true
            permissions.CanManageMembers = true
        case "member", "viewer":
            // All members can view invites (matches ListInvites handler behavior)
            permissions.CanViewInvites = true   // ✅ All members can VIEW invites
            permissions.CanCreateInvites = false // ❌ Only owners/admins can CREATE
            permissions.CanRevokeInvites = false // ❌ Only owners/admins can REVOKE
            permissions.CanManageMembers = false
        default:
            // No permissions for unknown roles
            permissions.CanViewInvites = false
            permissions.CanCreateInvites = false
            permissions.CanRevokeInvites = false
            permissions.CanManageMembers = false
        }
    }

    return &userRole, permissions
}
```

### Permission Matrix

| Role | View Invites | Create Invites | Revoke Invites | Manage Members |
|------|--------------|----------------|----------------|----------------|
| **Owner** | ✅ | ✅ | ✅ | ✅ |
| **Admin** | ✅ | ✅ | ✅ | ✅ |
| **Member** | ✅ | ❌ | ❌ | ❌ |
| **Viewer** | ✅ | ❌ | ❌ | ❌ |

### Key Authorization Rules

1. **View Invites** (`ListInvites`):
   - ✅ All project members (owner, admin, member, viewer)
   - Uses: `CheckProjectAccess` (checks `project_members` table)

2. **Create Invites** (`CreateInvite`):
   - ✅ Only owners and admins
   - ❌ Regular members cannot create
   - Uses: `CheckProjectOwnerOrAdmin`

3. **Revoke Invites** (`RevokeInvite`):
   - ✅ Only owners and admins
   - ❌ Regular members cannot revoke
   - Uses: `CheckProjectOwnerOrAdmin`

4. **Accept Invites** (`AcceptInvite`):
   - ✅ Any authenticated user (public invite links)
   - Validates invite token, expiration, and max uses

---

## 8. Complete Flow Example

### Scenario: Owner creates project and invite, member views and uses it

1. **Owner creates project**:
   ```
   POST /api/v1/projects
   → Project created
   → Owner added to project_members with role "owner"
   ```

2. **Owner creates invite**:
   ```
   POST /api/v1/projects/{projectId}/invites
   → Authorization: CheckProjectOwnerOrAdmin ✅
   → Invite created with token
   → Returns: { token: "abc123...", expiresAt: "...", ... }
   ```

3. **Owner/Member views invites**:
   ```
   GET /api/v1/projects/{projectId}/invites
   → Authorization: CheckProjectAccess ✅ (any member)
   → Returns: { invites: [...], count: 1 }
   ```

4. **New user gets invite details** (public):
   ```
   GET /api/v1/invites/{token}
   → No auth required
   → Returns: { invite: {...}, project: {...} }
   ```

5. **New user accepts invite**:
   ```
   POST /api/v1/invites/{token}/accept
   → User authenticated
   → Validates invite (active, not expired, max uses)
   → Adds user to project_members with role "member"
   → Increments invite use count
   → Returns: { project: {...}, userRole: "member", permissions: {...} }
   ```

6. **New member views invites**:
   ```
   GET /api/v1/projects/{projectId}/invites
   → Authorization: CheckProjectAccess ✅ (now a member)
   → Returns: { invites: [...], count: 1 }
   → Can see invite links but cannot create/revoke
   ```

---

## Summary

### Project Creation
- ✅ Creates project with owner
- ✅ Automatically adds owner to `project_members` table

### Invite Management
- ✅ **View**: All members can view invites
- ✅ **Create**: Only owners/admins can create
- ✅ **Revoke**: Only owners/admins can revoke
- ✅ **Accept**: Any authenticated user can accept valid invites

### Authorization
- ✅ Uses `CheckProjectAccess` for viewing (any member)
- ✅ Uses `CheckProjectOwnerOrAdmin` for creating/revoking (owners/admins only)
- ✅ Permissions returned in `ProjectResponse` for frontend UI control

