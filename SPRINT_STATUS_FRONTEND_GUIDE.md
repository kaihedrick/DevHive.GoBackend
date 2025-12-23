# Sprint Status Frontend Guide

## Backend Sprint Structure

The backend provides sprints with the following status fields:

```typescript
interface Sprint {
  id: string;
  projectId: string;
  name: string;
  description: string;
  startDate: string;        // ISO 8601 format
  endDate: string;          // ISO 8601 format
  isStarted: boolean;        // Has the sprint been started?
  isCompleted: boolean;      // Has the sprint been completed?
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
```

## Sprint Status Logic

Based on the backend code, sprints have three states:

1. **Planned** (Inactive): `!isStarted && !isCompleted`
   - Sprint has been created but not started yet
   - Future sprints that are scheduled

2. **Active**: `isStarted && !isCompleted`
   - Sprint is currently running
   - This is the only "active" state

3. **Completed** (Inactive): `isCompleted`
   - Sprint has been finished
   - Past sprints

## Frontend Implementation

### 1. Sprint Status Helper Functions

```typescript
// utils/sprintUtils.ts
export type SprintStatus = 'planned' | 'active' | 'completed';

export interface Sprint {
  id: string;
  projectId: string;
  name: string;
  description: string;
  startDate: string;
  endDate: string;
  isStarted: boolean;
  isCompleted: boolean;
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

/**
 * Get the status of a sprint
 */
export function getSprintStatus(sprint: Sprint): SprintStatus {
  if (sprint.isCompleted) {
    return 'completed';
  }
  if (sprint.isStarted) {
    return 'active';
  }
  return 'planned';
}

/**
 * Check if a sprint is active
 */
export function isSprintActive(sprint: Sprint): boolean {
  return sprint.isStarted && !sprint.isCompleted;
}

/**
 * Check if a sprint is inactive (planned or completed)
 */
export function isSprintInactive(sprint: Sprint): boolean {
  return !isSprintActive(sprint);
}

/**
 * Check if a sprint is planned (not started, not completed)
 */
export function isSprintPlanned(sprint: Sprint): boolean {
  return !sprint.isStarted && !sprint.isCompleted;
}

/**
 * Check if a sprint is completed
 */
export function isSprintCompleted(sprint: Sprint): boolean {
  return sprint.isCompleted;
}

/**
 * Get days remaining in sprint (if active)
 */
export function getDaysRemaining(sprint: Sprint): number | null {
  if (!isSprintActive(sprint)) {
    return null;
  }

  const endDate = new Date(sprint.endDate);
  const now = new Date();
  const diffTime = endDate.getTime() - now.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

  return diffDays > 0 ? diffDays : 0;
}

/**
 * Get sprint progress percentage (based on dates)
 */
export function getSprintProgress(sprint: Sprint): number {
  if (!isSprintActive(sprint)) {
    return sprint.isCompleted ? 100 : 0;
  }

  const startDate = new Date(sprint.startDate);
  const endDate = new Date(sprint.endDate);
  const now = new Date();

  const totalDuration = endDate.getTime() - startDate.getTime();
  const elapsed = now.getTime() - startDate.getTime();

  if (totalDuration <= 0) return 0;
  if (elapsed <= 0) return 0;
  if (elapsed >= totalDuration) return 100;

  return Math.round((elapsed / totalDuration) * 100);
}
```

### 2. Filtering and Grouping Sprints

```typescript
// hooks/useSprints.ts
import { useQuery } from '@tanstack/react-query';
import { fetchSprints } from '../services/sprintService';
import { Sprint, SprintStatus, getSprintStatus, isSprintActive } from '../utils/sprintUtils';

export function useSprints(projectId: string) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['sprints', projectId],
    queryFn: () => fetchSprints(projectId),
    enabled: !!projectId,
    staleTime: 2 * 60 * 1000, // 2 minutes
  });

  const sprints = data?.sprints || [];

  // Group sprints by status
  const activeSprints = sprints.filter(isSprintActive);
  const inactiveSprints = sprints.filter(isSprintInactive);
  const plannedSprints = sprints.filter(isSprintPlanned);
  const completedSprints = sprints.filter(isSprintCompleted);

  // Group by status object
  const sprintsByStatus = sprints.reduce((acc, sprint) => {
    const status = getSprintStatus(sprint);
    if (!acc[status]) {
      acc[status] = [];
    }
    acc[status].push(sprint);
    return acc;
  }, {} as Record<SprintStatus, Sprint[]>);

  return {
    sprints,
    activeSprints,
    inactiveSprints,
    plannedSprints,
    completedSprints,
    sprintsByStatus,
    isLoading,
    error,
  };
}
```

### 3. Sprint List Component

```typescript
// components/SprintList.tsx
import React from 'react';
import { useSprints } from '../hooks/useSprints';
import { SprintCard } from './SprintCard';
import { getSprintStatus } from '../utils/sprintUtils';

interface SprintListProps {
  projectId: string;
}

export function SprintList({ projectId }: SprintListProps) {
  const {
    activeSprints,
    plannedSprints,
    completedSprints,
    isLoading,
  } = useSprints(projectId);

  if (isLoading) {
    return <div>Loading sprints...</div>;
  }

  return (
    <div className="sprint-list">
      {/* Active Sprints Section */}
      {activeSprints.length > 0 && (
        <section className="active-sprints">
          <h2>Active Sprints</h2>
          <div className="sprint-grid">
            {activeSprints.map(sprint => (
              <SprintCard key={sprint.id} sprint={sprint} />
            ))}
          </div>
        </section>
      )}

      {/* Planned Sprints Section */}
      {plannedSprints.length > 0 && (
        <section className="planned-sprints">
          <h2>Planned Sprints</h2>
          <div className="sprint-grid">
            {plannedSprints.map(sprint => (
              <SprintCard key={sprint.id} sprint={sprint} />
            ))}
          </div>
        </section>
      )}

      {/* Completed Sprints Section (Collapsible) */}
      {completedSprints.length > 0 && (
        <details className="completed-sprints">
          <summary>
            <h2>Completed Sprints ({completedSprints.length})</h2>
          </summary>
          <div className="sprint-grid">
            {completedSprints.map(sprint => (
              <SprintCard key={sprint.id} sprint={sprint} />
            ))}
          </div>
        </details>
      )}

      {/* Empty State */}
      {activeSprints.length === 0 && 
       plannedSprints.length === 0 && 
       completedSprints.length === 0 && (
        <div className="empty-state">
          <p>No sprints found. Create your first sprint to get started!</p>
        </div>
      )}
    </div>
  );
}
```

### 4. Sprint Card Component

```typescript
// components/SprintCard.tsx
import React from 'react';
import { Sprint, getSprintStatus, getDaysRemaining, getSprintProgress } from '../utils/sprintUtils';
import { format } from 'date-fns'; // or your date library

interface SprintCardProps {
  sprint: Sprint;
}

export function SprintCard({ sprint }: SprintCardProps) {
  const status = getSprintStatus(sprint);
  const daysRemaining = getDaysRemaining(sprint);
  const progress = getSprintProgress(sprint);

  const statusColors = {
    active: 'bg-green-100 text-green-800 border-green-300',
    planned: 'bg-blue-100 text-blue-800 border-blue-300',
    completed: 'bg-gray-100 text-gray-800 border-gray-300',
  };

  const statusLabels = {
    active: 'Active',
    planned: 'Planned',
    completed: 'Completed',
  };

  return (
    <div className={`sprint-card border rounded-lg p-4 ${statusColors[status]}`}>
      <div className="flex justify-between items-start mb-2">
        <h3 className="font-semibold text-lg">{sprint.name}</h3>
        <span className={`px-2 py-1 rounded text-xs font-medium ${statusColors[status]}`}>
          {statusLabels[status]}
        </span>
      </div>

      {sprint.description && (
        <p className="text-sm text-gray-600 mb-3">{sprint.description}</p>
      )}

      <div className="text-sm mb-3">
        <div>
          <span className="font-medium">Start:</span>{' '}
          {format(new Date(sprint.startDate), 'MMM d, yyyy')}
        </div>
        <div>
          <span className="font-medium">End:</span>{' '}
          {format(new Date(sprint.endDate), 'MMM d, yyyy')}
        </div>
      </div>

      {/* Progress Bar for Active Sprints */}
      {status === 'active' && (
        <div className="mb-2">
          <div className="flex justify-between text-xs mb-1">
            <span>Progress</span>
            <span>{progress}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-green-600 h-2 rounded-full transition-all"
              style={{ width: `${progress}%` }}
            />
          </div>
          {daysRemaining !== null && (
            <div className="text-xs mt-1 text-gray-600">
              {daysRemaining === 0
                ? 'Ends today'
                : daysRemaining === 1
                ? '1 day remaining'
                : `${daysRemaining} days remaining`}
            </div>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-2 mt-3">
        <button className="text-sm text-blue-600 hover:text-blue-800">
          View Details
        </button>
        {status === 'planned' && (
          <button className="text-sm text-green-600 hover:text-green-800">
            Start Sprint
          </button>
        )}
        {status === 'active' && (
          <button className="text-sm text-orange-600 hover:text-orange-800">
            Complete Sprint
          </button>
        )}
      </div>
    </div>
  );
}
```

### 5. Tabbed Sprint View

```typescript
// components/SprintTabs.tsx
import React, { useState } from 'react';
import { useSprints } from '../hooks/useSprints';
import { SprintCard } from './SprintCard';

interface SprintTabsProps {
  projectId: string;
}

type Tab = 'active' | 'planned' | 'completed' | 'all';

export function SprintTabs({ projectId }: SprintTabsProps) {
  const [activeTab, setActiveTab] = useState<Tab>('active');
  const {
    sprints,
    activeSprints,
    plannedSprints,
    completedSprints,
    isLoading,
  } = useSprints(projectId);

  const tabs = [
    { id: 'active' as Tab, label: 'Active', count: activeSprints.length },
    { id: 'planned' as Tab, label: 'Planned', count: plannedSprints.length },
    { id: 'completed' as Tab, label: 'Completed', count: completedSprints.length },
    { id: 'all' as Tab, label: 'All', count: sprints.length },
  ];

  const getSprintsForTab = () => {
    switch (activeTab) {
      case 'active':
        return activeSprints;
      case 'planned':
        return plannedSprints;
      case 'completed':
        return completedSprints;
      case 'all':
        return sprints;
      default:
        return [];
    }
  };

  if (isLoading) {
    return <div>Loading sprints...</div>;
  }

  return (
    <div className="sprint-tabs">
      {/* Tab Navigation */}
      <div className="flex border-b mb-4">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 font-medium text-sm ${
              activeTab === tab.id
                ? 'border-b-2 border-blue-600 text-blue-600'
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            {tab.label}
            {tab.count > 0 && (
              <span className="ml-2 px-2 py-0.5 text-xs bg-gray-200 rounded-full">
                {tab.count}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Sprint List */}
      <div className="sprint-grid">
        {getSprintsForTab().length === 0 ? (
          <div className="empty-state text-center py-8 text-gray-500">
            No {activeTab === 'all' ? '' : activeTab} sprints found
          </div>
        ) : (
          getSprintsForTab().map(sprint => (
            <SprintCard key={sprint.id} sprint={sprint} />
          ))
        )}
      </div>
    </div>
  );
}
```

### 6. Sprint Status Badge Component

```typescript
// components/SprintStatusBadge.tsx
import React from 'react';
import { Sprint, getSprintStatus } from '../utils/sprintUtils';

interface SprintStatusBadgeProps {
  sprint: Sprint;
  size?: 'sm' | 'md' | 'lg';
}

export function SprintStatusBadge({ sprint, size = 'md' }: SprintStatusBadgeProps) {
  const status = getSprintStatus(sprint);

  const styles = {
    active: {
      sm: 'bg-green-100 text-green-800 text-xs px-2 py-0.5',
      md: 'bg-green-100 text-green-800 text-sm px-2.5 py-1',
      lg: 'bg-green-100 text-green-800 text-base px-3 py-1.5',
    },
    planned: {
      sm: 'bg-blue-100 text-blue-800 text-xs px-2 py-0.5',
      md: 'bg-blue-100 text-blue-800 text-sm px-2.5 py-1',
      lg: 'bg-blue-100 text-blue-800 text-base px-3 py-1.5',
    },
    completed: {
      sm: 'bg-gray-100 text-gray-800 text-xs px-2 py-0.5',
      md: 'bg-gray-100 text-gray-800 text-sm px-2.5 py-1',
      lg: 'bg-gray-100 text-gray-800 text-base px-3 py-1.5',
    },
  };

  const labels = {
    active: 'Active',
    planned: 'Planned',
    completed: 'Completed',
  };

  return (
    <span className={`inline-flex items-center rounded-full font-medium ${styles[status][size]}`}>
      {labels[status]}
    </span>
  );
}
```

### 7. Update Sprint Status

```typescript
// services/sprintService.ts
import apiClient from './apiClient';

export async function updateSprintStatus(
  sprintId: string,
  isStarted: boolean,
  isCompleted: boolean
) {
  const response = await apiClient.patch(`/projects/{projectId}/sprints/${sprintId}/status`, {
    isStarted,
    isCompleted,
  });
  return response.data;
}

// Usage in component
function SprintActions({ sprint }: { sprint: Sprint }) {
  const queryClient = useQueryClient();
  const updateMutation = useMutation({
    mutationFn: ({ isStarted, isCompleted }: { isStarted: boolean; isCompleted: boolean }) =>
      updateSprintStatus(sprint.id, isStarted, isCompleted),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sprints'] });
    },
  });

  const handleStartSprint = () => {
    updateMutation.mutate({ isStarted: true, isCompleted: false });
  };

  const handleCompleteSprint = () => {
    updateMutation.mutate({ isStarted: true, isCompleted: true });
  };

  return (
    <div>
      {isSprintPlanned(sprint) && (
        <button onClick={handleStartSprint}>Start Sprint</button>
      )}
      {isSprintActive(sprint) && (
        <button onClick={handleCompleteSprint}>Complete Sprint</button>
      )}
    </div>
  );
}
```

## Summary

### Sprint States

- **Planned** (`!isStarted && !isCompleted`): Future sprints, not yet started
- **Active** (`isStarted && !isCompleted`): Currently running sprints
- **Completed** (`isCompleted`): Finished sprints

### Key Functions

- `getSprintStatus(sprint)`: Returns 'planned' | 'active' | 'completed'
- `isSprintActive(sprint)`: Returns true if sprint is active
- `isSprintInactive(sprint)`: Returns true if sprint is planned or completed
- `getDaysRemaining(sprint)`: Returns days left (null if not active)
- `getSprintProgress(sprint)`: Returns progress percentage (0-100)

### UI Patterns

1. **Grouped View**: Separate sections for Active, Planned, and Completed
2. **Tabbed View**: Tabs to switch between different sprint states
3. **Status Badges**: Visual indicators for sprint status
4. **Progress Bars**: Show progress for active sprints
5. **Actions**: Context-sensitive buttons (Start/Complete) based on status


