import React from 'react';
import { useProject } from '../hooks/useProjects';
import { ProjectInvites } from './ProjectInvites';

/**
 * Example: How to integrate ProjectInvites into your project page
 */
export const ProjectPage: React.FC<{ projectId: string }> = ({ projectId }) => {
  const { data: project, isLoading } = useProject(projectId);

  if (isLoading) {
    return <div>Loading project...</div>;
  }

  if (!project) {
    return <div>Project not found</div>;
  }

  return (
    <div className="project-page">
      <h1>{project.name}</h1>
      <p>{project.description}</p>
      
      {/* Project Invites Section */}
      <div className="mt-8">
        <ProjectInvites
          projectId={projectId}
          userRole={project.userRole} // From ProjectResponse
          permissions={project.permissions} // From ProjectResponse
        />
      </div>
    </div>
  );
};




