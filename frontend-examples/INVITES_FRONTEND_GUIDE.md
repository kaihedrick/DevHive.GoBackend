# Project Invites - Frontend Implementation Guide

## Overview

This guide shows you how to implement project invites in your React frontend using TanStack Query. All project members (owners, admins, and regular members) can **view** invites, but only owners and admins can **create** and **revoke** invites.

## Quick Start

### 1. Install Dependencies

```bash
npm install @tanstack/react-query axios
```

### 2. Copy Files

Copy these files to your frontend project:
- `src/hooks/useInvites.ts` - React hooks for invite operations
- `src/components/ProjectInvites.tsx` - Example component (optional, customize as needed)

### 3. Use the Hooks

```typescript
import { useProjectInvites, useCreateInvite } from './hooks/useInvites';

function MyProjectPage({ projectId }: { projectId: string }) {
  const { data: invites, isLoading } = useProjectInvites(projectId);
  const createInvite = useCreateInvite();

  // ... your component logic
}
```

---

## API Endpoints

### 1. List Project Invites
**Endpoint**: `GET /api/v1/projects/{projectId}/invites`

**Authentication**: ✅ Required (Bearer Token)

**Permissions**: All project members can view invites

**Response**:
```json
{
  "invites": [
    {
      "id": "uuid",
      "projectId": "uuid",
      "token": "invite-token-uuid",
      "expiresAt": "2025-01-20T15:30:00Z",
      "maxUses": 10,
      "usedCount": 3,
      "isActive": true,
      "createdAt": "2025-01-20T10:00:00Z"
    }
  ],
  "count": 1
}
```

**Error Responses**:
- `401 Unauthorized` - Missing/invalid token
- `403 Forbidden` - User is not a project member
- `404 Not Found` - Project doesn't exist

---

### 2. Create Invite
**Endpoint**: `POST /api/v1/projects/{projectId}/invites`

**Authentication**: ✅ Required (Bearer Token)

**Permissions**: Only owners and admins can create invites

**Request Body**:
```json
{
  "expiresInMinutes": 30,  // Optional, defaults to 30
  "maxUses": 10            // Optional, null = unlimited
}
```

**Response** (201 Created):
```json
{
  "id": "uuid",
  "projectId": "uuid",
  "token": "invite-token-uuid",
  "expiresAt": "2025-01-20T15:30:00Z",
  "maxUses": 10,
  "usedCount": 0,
  "isActive": true,
  "createdAt": "2025-01-20T10:00:00Z"
}
```

**Error Responses**:
- `403 Forbidden` - User is not owner/admin
- `400 Bad Request` - Invalid request data

---

### 3. Revoke Invite
**Endpoint**: `DELETE /api/v1/projects/{projectId}/invites/{inviteId}`

**Authentication**: ✅ Required (Bearer Token)

**Permissions**: Only owners and admins can revoke invites

**Response** (200 OK):
```json
{
  "message": "Invite revoked successfully"
}
```

**Error Responses**:
- `403 Forbidden` - User is not owner/admin
- `404 Not Found` - Invite doesn't exist

---

### 4. Get Invite Details (Public)
**Endpoint**: `GET /api/v1/invites/{inviteToken}`

**Authentication**: ❌ None (Public endpoint)

**Response** (200 OK):
```json
{
  "invite": {
    "id": "uuid",
    "projectId": "uuid",
    "token": "invite-token",
    "expiresAt": "2025-01-20T15:30:00Z",
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

---

### 5. Accept Invite
**Endpoint**: `POST /api/v1/invites/{inviteToken}/accept`

**Authentication**: ✅ Required (Bearer Token)

**Response** (200 OK):
```json
{
  "id": "uuid",
  "ownerId": "uuid",
  "name": "Project Name",
  "description": "Project Description",
  "createdAt": "2025-01-20T10:00:00Z",
  "updatedAt": "2025-01-20T10:00:00Z",
  "owner": {
    "id": "uuid",
    "username": "username",
    "email": "email@example.com",
    "firstName": "First",
    "lastName": "Last"
  }
}
```

---

## React Hooks Usage

### Query: List Invites

```typescript
import { useProjectInvites } from './hooks/useInvites';

function InvitesList({ projectId }: { projectId: string }) {
  const { data, isLoading, error } = useProjectInvites(projectId);

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      <h2>Invites ({data?.count || 0})</h2>
      {data?.invites.map(invite => (
        <div key={invite.id}>
          <p>Token: {invite.token}</p>
          <p>Expires: {invite.expiresAt}</p>
          <p>Uses: {invite.usedCount} / {invite.maxUses || 'Unlimited'}</p>
        </div>
      ))}
    </div>
  );
}
```

### Mutation: Create Invite

```typescript
import { useCreateInvite } from './hooks/useInvites';

function CreateInviteButton({ projectId }: { projectId: string }) {
  const createInvite = useCreateInvite();

  const handleCreate = async () => {
    try {
      const newInvite = await createInvite.mutateAsync({
        projectId,
        data: {
          expiresInMinutes: 60, // 1 hour
          maxUses: 5, // Limit to 5 uses
        },
      });
      console.log('Created invite:', newInvite);
    } catch (error) {
      console.error('Failed to create invite:', error);
    }
  };

  return (
    <button 
      onClick={handleCreate}
      disabled={createInvite.isPending}
    >
      {createInvite.isPending ? 'Creating...' : 'Create Invite'}
    </button>
  );
}
```

### Mutation: Revoke Invite

```typescript
import { useRevokeInvite } from './hooks/useInvites';

function RevokeInviteButton({ 
  projectId, 
  inviteId 
}: { 
  projectId: string; 
  inviteId: string;
}) {
  const revokeInvite = useRevokeInvite();

  const handleRevoke = async () => {
    if (!confirm('Are you sure you want to revoke this invite?')) {
      return;
    }

    try {
      await revokeInvite.mutateAsync({ projectId, inviteId });
      console.log('Invite revoked');
    } catch (error) {
      console.error('Failed to revoke invite:', error);
    }
  };

  return (
    <button 
      onClick={handleRevoke}
      disabled={revokeInvite.isPending}
    >
      {revokeInvite.isPending ? 'Revoking...' : 'Revoke'}
    </button>
  );
}
```

### Query: Get Invite Details (Public)

```typescript
import { useInviteDetails } from './hooks/useInvites';
import { useParams } from 'react-router-dom';

function InviteAcceptPage() {
  const { token } = useParams<{ token: string }>();
  const { data, isLoading, error } = useInviteDetails(token || null);

  if (isLoading) return <div>Loading invite...</div>;
  if (error) return <div>Invalid or expired invite</div>;

  return (
    <div>
      <h1>Join {data?.project.name}</h1>
      <p>{data?.project.description}</p>
      <button>Accept Invite</button>
    </div>
  );
}
```

### Mutation: Accept Invite

```typescript
import { useAcceptInvite } from './hooks/useInvites';
import { useNavigate } from 'react-router-dom';

function AcceptInviteButton({ token }: { token: string }) {
  const acceptInvite = useAcceptInvite();
  const navigate = useNavigate();

  const handleAccept = async () => {
    try {
      const project = await acceptInvite.mutateAsync(token);
      // Redirect to project page
      navigate(`/projects/${project.id}`);
    } catch (error) {
      console.error('Failed to accept invite:', error);
    }
  };

  return (
    <button 
      onClick={handleAccept}
      disabled={acceptInvite.isPending}
    >
      {acceptInvite.isPending ? 'Joining...' : 'Accept Invite'}
    </button>
  );
}
```

---

## Helper Functions

The `useInvites.ts` file includes helpful utility functions:

```typescript
import {
  isInviteExpired,
  isInviteMaxedOut,
  isInviteValid,
  getInviteUrl,
  formatInviteExpiration,
} from './hooks/useInvites';

// Check if invite is expired
if (isInviteExpired(invite)) {
  console.log('This invite has expired');
}

// Check if invite reached max uses
if (isInviteMaxedOut(invite)) {
  console.log('This invite has reached its usage limit');
}

// Check if invite is valid (not expired, not maxed out, and active)
if (isInviteValid(invite)) {
  console.log('This invite can be used');
}

// Get the full invite URL
const inviteUrl = getInviteUrl(invite);
// Returns: "https://yourapp.com/invite/abc123..."

// Format expiration for display
const expirationText = formatInviteExpiration(invite.expiresAt);
// Returns: "2 hours remaining" or "Expired"
```

---

## Permission-Based UI

Use the `userRole` and `permissions` from the project response to conditionally show UI:

```typescript
import { useProject } from './hooks/useProjects';
import { useProjectInvites, useCreateInvite } from './hooks/useInvites';

function ProjectSettingsPage({ projectId }: { projectId: string }) {
  const { data: project } = useProject(projectId);
  const { data: invites } = useProjectInvites(projectId);
  const createInvite = useCreateInvite();

  const canCreateInvites = project?.permissions?.canCreateInvites ?? false;
  const canRevokeInvites = project?.permissions?.canRevokeInvites ?? false;

  return (
    <div>
      <h2>Project Invites</h2>
      
      {/* All members can see invites */}
      <InvitesList invites={invites?.invites || []} />
      
      {/* Only owners/admins can create invites */}
      {canCreateInvites && (
        <button onClick={() => createInvite.mutate({ projectId })}>
          Create Invite
        </button>
      )}
    </div>
  );
}
```

---

## Error Handling

The hooks include built-in error handling, but you can customize it:

```typescript
const { data, error, isLoading } = useProjectInvites(projectId);

if (error) {
  // Handle specific error codes
  if (error.response?.status === 403) {
    // User doesn't have permission (shouldn't happen for viewing, but handle gracefully)
    return <div>You don't have permission to view invites</div>;
  }
  
  if (error.response?.status === 404) {
    // Project doesn't exist
    return <div>Project not found</div>;
  }
  
  // Other errors
  return <div>Error: {error.message}</div>;
}
```

---

## Complete Example Component

See `src/components/ProjectInvites.tsx` for a complete, production-ready component that:
- Lists all invites
- Shows invite status (active/expired/maxed out)
- Allows creating invites (owners/admins only)
- Allows revoking invites (owners/admins only)
- Copies invite links to clipboard
- Handles loading and error states
- Shows expiration time in human-readable format

---

## TypeScript Interfaces

All types are exported from `useInvites.ts`:

```typescript
import type {
  ProjectInvite,
  InvitesResponse,
  CreateInviteRequest,
  CreateInviteResponse,
} from './hooks/useInvites';
```

---

## Caching Strategy

The hooks use TanStack Query with the following caching:

- **List Invites**: 1 minute stale time, 5 minutes cache retention
- **Invite Details**: 5 minutes stale time, 10 minutes cache retention
- **Mutations**: Automatically invalidate relevant queries on success

This ensures:
- Fresh data when needed
- Reduced API calls
- Automatic refetching when data changes

---

## Best Practices

1. **Always check permissions** before showing create/revoke buttons
2. **Handle 403 errors gracefully** - show empty state instead of error
3. **Use helper functions** to check invite validity before displaying
4. **Copy invite links** instead of showing full tokens
5. **Show expiration time** in human-readable format
6. **Invalidate queries** after mutations to keep UI in sync

---

## Troubleshooting

### 403 Forbidden when viewing invites
- **Cause**: User is not a project member
- **Fix**: Ensure user has access to the project first

### 403 Forbidden when creating/revoking invites
- **Cause**: User is not owner/admin
- **Fix**: Check `userRole` or `permissions.canCreateInvites` before showing UI

### Invites not updating after creation
- **Cause**: Query cache not invalidated
- **Fix**: The hooks automatically invalidate, but you can manually refetch:
  ```typescript
  const { refetch } = useProjectInvites(projectId);
  // After mutation
  refetch();
  ```

---

## Next Steps

1. Copy `useInvites.ts` to your hooks directory
2. Copy `ProjectInvites.tsx` as a starting point (customize as needed)
3. Integrate into your project settings page
4. Add invite acceptance flow for public invite links
5. Style components to match your design system



