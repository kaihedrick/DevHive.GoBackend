# API Endpoints Reference - Project & Invite Endpoints

## Request/Response Structures

### 1. POST `/api/v1/projects/join` - Join Project by ID

**Request Body:**
```json
{
  "projectId": "uuid-string"
}
```

**Response (200 OK):**
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "name": "Project Name",
  "description": "Project Description",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "owner": {
    "id": "uuid",
    "username": "username",
    "email": "email@example.com",
    "firstName": "First",
    "lastName": "Last"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid project ID or missing projectId
- `401 Unauthorized` - User not authenticated
- `404 Not Found` - Project doesn't exist
- `500 Internal Server Error` - Server error

---

### 2. POST `/api/v1/invites/{inviteToken}/accept` - Accept Project Invite

**Request:** None (token in URL path)

**Response (200 OK):**
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "name": "Project Name",
  "description": "Project Description",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "owner": {
    "id": "uuid",
    "username": "username",
    "email": "email@example.com",
    "firstName": "First",
    "lastName": "Last"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invite expired, inactive, or reached max uses
- `401 Unauthorized` - User not authenticated
- `404 Not Found` - Invalid or expired invite token

---

### 3. GET `/api/v1/invites/{inviteToken}` - Get Invite Details (Public)

**Request:** None (token in URL path)

**Response (200 OK):**
```json
{
  "invite": {
    "id": "uuid",
    "projectId": "uuid",
    "token": "invite-token",
    "expiresAt": "2024-01-01T00:00:00Z",
    "maxUses": 10,
    "usedCount": 3,
    "isActive": true
  },
  "project": {
    "id": "uuid",
    "name": "Project Name",
    "description": "Project Description"
  }
}
```

**Note:** The `invite` object in this response does not include `createdAt` or `updatedAt` fields.

**Error Responses:**
- `400 Bad Request` - Invite has expired
- `404 Not Found` - Invalid or expired invite token

---

### 4. GET `/api/v1/projects/{projectId}` - Get Project

**Response (200 OK):**
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "name": "Project Name",
  "description": "Project Description",
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "owner": {
    "id": "uuid",
    "username": "username",
    "email": "email@example.com",
    "firstName": "First",
    "lastName": "Last"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid project ID
- `401 Unauthorized` - User not authenticated
- `403 Forbidden` - User doesn't have access to project
- `404 Not Found` - Project doesn't exist

---

### 5. POST `/api/v1/projects/{projectId}/invites` - Create Invite

**Request Body:**
```json
{
  "expiresInMinutes": 30,  // Optional, defaults to 30
  "maxUses": 10            // Optional, null = unlimited
}
```

**Response (201 Created):**
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "token": "generated-uuid-token",
  "expiresAt": "2024-01-01T00:00:00Z",
  "maxUses": 10,
  "usedCount": 0,
  "isActive": true,
  "createdAt": "2024-01-01T00:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - User not authenticated
- `403 Forbidden` - Only project owners and admins can create invites

---

### 6. GET `/api/v1/projects/{projectId}/invites` - List Invites

**Response (200 OK):**
```json
{
  "invites": [
    {
      "id": "uuid",
      "projectId": "uuid",
      "token": "invite-token",
      "expiresAt": "2024-01-01T00:00:00Z",
      "maxUses": 10,
      "usedCount": 3,
      "isActive": true,
      "createdAt": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

**Error Responses:**
- `401 Unauthorized` - User not authenticated
- `403 Forbidden` - Only project owners and admins can view invites

---

### 7. DELETE `/api/v1/projects/{projectId}/invites/{inviteId}` - Revoke Invite

**Response (200 OK):**
```json
{
  "message": "Invite revoked successfully"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid invite ID
- `401 Unauthorized` - User not authenticated
- `403 Forbidden` - Only project owners and admins can revoke invites
- `404 Not Found` - Invite not found

---

## WebSocket Cache Invalidation

### Message Structure

When database changes occur, WebSocket clients receive cache invalidation notifications:

```json
{
  "type": "cache_invalidate",
  "data": {
    "resource": "project_members",
    "id": "project_id:user_id",
    "action": "INSERT",
    "project_id": "uuid",
    "timestamp": "2024-01-01T00:00:00Z"
  },
  "project_id": "uuid"
}
```

**WebSocket Message Structure:**
- `type` - Message type: "cache_invalidate" or "reconnect"
- `data` - The cache invalidation payload (see above)
- `project_id` - The project ID this notification relates to
- `resource` - (optional) Resource type in the message
- `action` - (optional) Action type in the message

**Resource Types:**
- `projects` - Project changes
- `sprints` - Sprint changes
- `tasks` - Task changes
- `project_members` - Member join/leave changes

**Actions:**
- `INSERT` - New record created
- `UPDATE` - Record updated
- `DELETE` - Record deleted

---

## Important Notes

1. **All project responses now include the `owner` field** - This ensures consistency across all endpoints and prevents frontend cache mismatches.

2. **Cache Invalidation Flow:**
   - When a user joins a project, a `project_members` INSERT notification is sent
   - Frontend should handle this gracefully and not immediately refetch if the data is already available
   - The complete project response is returned immediately, so refetching is optional

3. **Race Condition Prevention:**
   - The backend returns the complete project data immediately after adding the member
   - Frontend should use this data directly rather than immediately refetching
   - WebSocket notifications are sent after the transaction commits

