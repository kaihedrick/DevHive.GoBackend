import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient from '../lib/apiClient';

interface Project {
  id: string;
  ownerId: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
  owner: {
    id: string;
    username: string;
    email: string;
    firstName: string;
    lastName: string;
  };
}

interface ProjectsResponse {
  projects: Project[];
  limit: number;
  offset: number;
}

interface CreateProjectRequest {
  name: string;
  description: string;
}

interface UpdateProjectRequest {
  name?: string;
  description?: string;
}

// Query: List projects
export const useProjects = (limit = 20, offset = 0) => {
  return useQuery<ProjectsResponse>({
    queryKey: ['projects', limit, offset],
    queryFn: async () => {
      const response = await apiClient.get('/projects', {
        params: { limit, offset },
      });
      return response.data;
    },
    staleTime: 2 * 60 * 1000, // 2 minutes - projects list changes more frequently
    gcTime: 10 * 60 * 1000, // 10 minutes cache retention
  });
};

// Query: Get single project
export const useProject = (projectId: string | null) => {
  return useQuery<Project>({
    queryKey: ['projects', projectId],
    queryFn: async () => {
      const response = await apiClient.get(`/projects/${projectId}`);
      return response.data;
    },
    enabled: !!projectId, // Only fetch if projectId is provided
    staleTime: 5 * 60 * 1000, // 5 minutes - single project changes less frequently
    gcTime: 15 * 60 * 1000, // 15 minutes cache retention
  });
};

// Mutation: Create project
export const useCreateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: CreateProjectRequest) => {
      const response = await apiClient.post('/projects', data);
      return response.data;
    },
    onSuccess: () => {
      // Invalidate projects list to refetch
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

// Mutation: Update project
export const useUpdateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ projectId, data }: { projectId: string; data: UpdateProjectRequest }) => {
      const response = await apiClient.patch(`/projects/${projectId}`, data);
      return response.data;
    },
    onSuccess: (data, variables) => {
      // Update cache optimistically
      queryClient.setQueryData(['projects', variables.projectId], data);
      // Invalidate list to ensure consistency
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

// Mutation: Delete project
export const useDeleteProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (projectId: string) => {
      await apiClient.delete(`/projects/${projectId}`);
    },
    onSuccess: (_, projectId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: ['projects', projectId] });
      // Invalidate list
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

// Mutation: Join project
export const useJoinProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (projectId: string) => {
      const response = await apiClient.post('/projects/join', { projectId });
      return response.data;
    },
    onSuccess: (data) => {
      // Add project to cache
      queryClient.setQueryData(['projects', data.id], data);
      // Invalidate list to show new project
      queryClient.invalidateQueries({ queryKey: ['projects'] });
    },
  });
};

