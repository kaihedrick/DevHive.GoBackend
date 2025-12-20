import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../lib/apiClient';

// ============================================================================
// TypeScript Interfaces
// ============================================================================

export interface ProjectInvite {
  id: string;
  projectId: string;
  token: string;
  expiresAt: string; // ISO 8601 format
  maxUses: number | null; // null = unlimited
  usedCount: number;
  isActive: boolean;
  createdAt: string; // ISO 8601 format
}

export interface InvitesResponse {
  invites: ProjectInvite[];
  count: number;
}

export interface CreateInviteRequest {
  expiresInMinutes?: number; // Optional, defaults to 30
  maxUses?: number; // Optional, null = unlimited
}

export interface CreateInviteResponse {
  id: string;
  projectId: string;
  token: string;
  expiresAt: string;
  maxUses: number | null;
  usedCount: number;
  isActive: boolean;
  createdAt: string;
}

// ============================================================================
// React Hooks
// ============================================================================

/**
 * Query: List all invites for a project
 * 
 * All project members (owners, admins, and regular members) can view invites.
 * Returns empty array if user doesn't have permission (graceful degradation).
 * 
 * @param projectId - The project ID to fetch invites for
 */
export const useProjectInvites = (projectId: string | null) => {
  return useQuery<InvitesResponse>({
    queryKey: ['projectInvites', projectId],
    queryFn: async () => {
      if (!projectId) {
        throw new Error('Project ID is required');
      }
      
      const response = await apiClient.get<InvitesResponse>(
        `/projects/${projectId}/invites`
      );
      return response.data;
    },
    enabled: !!projectId, // Only run query if projectId is provided
    staleTime: 1 * 60 * 1000, // 1 minute - invites change frequently
    gcTime: 5 * 60 * 1000, // 5 minutes cache retention
    retry: (failureCount, error: any) => {
      // Don't retry on 403 (permission denied) or 404 (not found)
      if (error?.response?.status === 403 || error?.response?.status === 404) {
        return false;
      }
      // Retry up to 2 times for other errors
      return failureCount < 2;
    },
  });
};

/**
 * Query: Get invite details by token (public endpoint, no auth required)
 * 
 * Used when a user clicks an invite link to see project details before accepting.
 * 
 * @param inviteToken - The invite token from the URL
 */
export const useInviteDetails = (inviteToken: string | null) => {
  return useQuery({
    queryKey: ['inviteDetails', inviteToken],
    queryFn: async () => {
      if (!inviteToken) {
        throw new Error('Invite token is required');
      }
      
      const response = await apiClient.get(`/invites/${inviteToken}`);
      return response.data;
    },
    enabled: !!inviteToken,
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 10 * 60 * 1000, // 10 minutes
    retry: (failureCount, error: any) => {
      // Don't retry on 404 (invalid/expired invite)
      if (error?.response?.status === 404) {
        return false;
      }
      return failureCount < 2;
    },
  });
};

/**
 * Mutation: Accept an invite
 * 
 * Adds the current user to the project.
 * 
 * @param inviteToken - The invite token to accept
 */
export const useAcceptInvite = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (inviteToken: string) => {
      const response = await apiClient.post(`/invites/${inviteToken}/accept`);
      return response.data;
    },
    onSuccess: (data) => {
      // Invalidate and refetch relevant queries
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      queryClient.invalidateQueries({ queryKey: ['projectInvites', data.id] });
      queryClient.invalidateQueries({ queryKey: ['inviteDetails'] });
    },
  });
};

/**
 * Mutation: Create a new invite
 * 
 * Only owners and admins can create invites.
 * 
 * @param projectId - The project ID to create an invite for
 */
export const useCreateInvite = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ 
      projectId, 
      data 
    }: { 
      projectId: string; 
      data?: CreateInviteRequest 
    }) => {
      const response = await apiClient.post<CreateInviteResponse>(
        `/projects/${projectId}/invites`,
        data || {}
      );
      return response.data;
    },
    onSuccess: (data, variables) => {
      // Invalidate invites list to show the new invite
      queryClient.invalidateQueries({ 
        queryKey: ['projectInvites', variables.projectId] 
      });
    },
  });
};

/**
 * Mutation: Revoke (deactivate) an invite
 * 
 * Only owners and admins can revoke invites.
 * 
 * @param projectId - The project ID
 * @param inviteId - The invite ID to revoke
 */
export const useRevokeInvite = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ 
      projectId, 
      inviteId 
    }: { 
      projectId: string; 
      inviteId: string 
    }) => {
      const response = await apiClient.delete(
        `/projects/${projectId}/invites/${inviteId}`
      );
      return response.data;
    },
    onSuccess: (data, variables) => {
      // Invalidate invites list to remove the revoked invite
      queryClient.invalidateQueries({ 
        queryKey: ['projectInvites', variables.projectId] 
      });
    },
  });
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Check if an invite is expired
 */
export const isInviteExpired = (invite: ProjectInvite): boolean => {
  const expiresAt = new Date(invite.expiresAt);
  return expiresAt < new Date();
};

/**
 * Check if an invite has reached max uses
 */
export const isInviteMaxedOut = (invite: ProjectInvite): boolean => {
  if (invite.maxUses === null) {
    return false; // Unlimited uses
  }
  return invite.usedCount >= invite.maxUses;
};

/**
 * Check if an invite is valid (not expired, not maxed out, and active)
 */
export const isInviteValid = (invite: ProjectInvite): boolean => {
  return invite.isActive && !isInviteExpired(invite) && !isInviteMaxedOut(invite);
};

/**
 * Generate the full invite URL
 */
export const getInviteUrl = (invite: ProjectInvite, baseUrl?: string): string => {
  const base = baseUrl || window.location.origin;
  return `${base}/invite/${invite.token}`;
};

/**
 * Format invite expiration time for display
 */
export const formatInviteExpiration = (expiresAt: string): string => {
  const date = new Date(expiresAt);
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  
  if (diffMs < 0) {
    return 'Expired';
  }
  if (diffDays > 0) {
    return `${diffDays} day${diffDays > 1 ? 's' : ''} remaining`;
  }
  if (diffHours > 0) {
    return `${diffHours} hour${diffHours > 1 ? 's' : ''} remaining`;
  }
  if (diffMins > 0) {
    return `${diffMins} minute${diffMins > 1 ? 's' : ''} remaining`;
  }
  return 'Expiring soon';
};

