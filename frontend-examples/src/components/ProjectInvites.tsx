import React, { useState } from 'react';
import {
  useProjectInvites,
  useCreateInvite,
  useRevokeInvite,
  isInviteValid,
  formatInviteExpiration,
  getInviteUrl,
  type ProjectInvite,
} from '../hooks/useInvites';

interface ProjectInvitesProps {
  projectId: string;
  userRole?: 'owner' | 'admin' | 'member';
  permissions?: {
    canCreateInvites?: boolean;
    canRevokeInvites?: boolean;
  };
}

/**
 * Component to display and manage project invites
 * 
 * All project members can view invites.
 * Only owners/admins can create and revoke invites.
 */
export const ProjectInvites: React.FC<ProjectInvitesProps> = ({
  projectId,
  userRole,
  permissions,
}) => {
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [expiresInMinutes, setExpiresInMinutes] = useState(30);
  const [maxUses, setMaxUses] = useState<number | null>(null);

  // Fetch invites
  const { data, isLoading, error, refetch } = useProjectInvites(projectId);
  const createInvite = useCreateInvite();
  const revokeInvite = useRevokeInvite();

  // Check if user can create/revoke invites
  const canCreate = permissions?.canCreateInvites ?? (userRole === 'owner' || userRole === 'admin');
  const canRevoke = permissions?.canRevokeInvites ?? (userRole === 'owner' || userRole === 'admin');

  const handleCreateInvite = async () => {
    try {
      await createInvite.mutateAsync({
        projectId,
        data: {
          expiresInMinutes: expiresInMinutes > 0 ? expiresInMinutes : undefined,
          maxUses: maxUses && maxUses > 0 ? maxUses : undefined,
        },
      });
      setShowCreateForm(false);
      setExpiresInMinutes(30);
      setMaxUses(null);
    } catch (error) {
      console.error('Failed to create invite:', error);
    }
  };

  const handleRevokeInvite = async (inviteId: string) => {
    if (!confirm('Are you sure you want to revoke this invite?')) {
      return;
    }
    
    try {
      await revokeInvite.mutateAsync({ projectId, inviteId });
    } catch (error) {
      console.error('Failed to revoke invite:', error);
    }
  };

  const handleCopyInviteLink = (invite: ProjectInvite) => {
    const url = getInviteUrl(invite);
    navigator.clipboard.writeText(url);
    // You might want to show a toast notification here
    alert('Invite link copied to clipboard!');
  };

  if (isLoading) {
    return <div>Loading invites...</div>;
  }

  if (error) {
    // Handle 403 gracefully - user might not have permission
    if ((error as any)?.response?.status === 403) {
      return (
        <div className="text-gray-500">
          You don't have permission to view invites.
        </div>
      );
    }
    
    return (
      <div className="text-red-500">
        Error loading invites: {(error as Error).message}
        <button onClick={() => refetch()}>Retry</button>
      </div>
    );
  }

  const invites = data?.invites || [];

  return (
    <div className="project-invites">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Project Invites</h2>
        {canCreate && (
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            {showCreateForm ? 'Cancel' : 'Create Invite'}
          </button>
        )}
      </div>

      {/* Create Invite Form */}
      {showCreateForm && canCreate && (
        <div className="mb-6 p-4 border rounded">
          <h3 className="font-semibold mb-3">Create New Invite</h3>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium mb-1">
                Expires in (minutes)
              </label>
              <input
                type="number"
                value={expiresInMinutes}
                onChange={(e) => setExpiresInMinutes(parseInt(e.target.value) || 30)}
                min="1"
                className="w-full px-3 py-2 border rounded"
                placeholder="30"
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                Max Uses (leave empty for unlimited)
              </label>
              <input
                type="number"
                value={maxUses || ''}
                onChange={(e) => setMaxUses(e.target.value ? parseInt(e.target.value) : null)}
                min="1"
                className="w-full px-3 py-2 border rounded"
                placeholder="Unlimited"
              />
            </div>
            <button
              onClick={handleCreateInvite}
              disabled={createInvite.isPending}
              className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600 disabled:opacity-50"
            >
              {createInvite.isPending ? 'Creating...' : 'Create Invite'}
            </button>
          </div>
          {createInvite.isError && (
            <div className="mt-2 text-red-500 text-sm">
              Failed to create invite: {(createInvite.error as Error).message}
            </div>
          )}
        </div>
      )}

      {/* Invites List */}
      {invites.length === 0 ? (
        <div className="text-gray-500 py-4">
          No active invites. {canCreate && 'Create one to get started!'}
        </div>
      ) : (
        <div className="space-y-3">
          {invites.map((invite) => (
            <InviteCard
              key={invite.id}
              invite={invite}
              canRevoke={canRevoke}
              onRevoke={() => handleRevokeInvite(invite.id)}
              onCopyLink={() => handleCopyInviteLink(invite)}
            />
          ))}
        </div>
      )}
    </div>
  );
};

/**
 * Individual invite card component
 */
interface InviteCardProps {
  invite: ProjectInvite;
  canRevoke: boolean;
  onRevoke: () => void;
  onCopyLink: () => void;
}

const InviteCard: React.FC<InviteCardProps> = ({
  invite,
  canRevoke,
  onRevoke,
  onCopyLink,
}) => {
  const isValid = isInviteValid(invite);
  const expirationText = formatInviteExpiration(invite.expiresAt);
  const inviteUrl = getInviteUrl(invite);

  return (
    <div className={`p-4 border rounded ${!isValid ? 'opacity-50' : ''}`}>
      <div className="flex justify-between items-start">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-2">
            <span className="font-mono text-sm bg-gray-100 px-2 py-1 rounded">
              {invite.token.substring(0, 8)}...
            </span>
            {isValid ? (
              <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                Active
              </span>
            ) : (
              <span className="px-2 py-1 bg-red-100 text-red-800 text-xs rounded">
                Invalid
              </span>
            )}
          </div>
          
          <div className="text-sm text-gray-600 space-y-1">
            <div>Expires: {expirationText}</div>
            <div>
              Uses: {invite.usedCount}
              {invite.maxUses !== null ? ` / ${invite.maxUses}` : ' / Unlimited'}
            </div>
            <div className="mt-2">
              <button
                onClick={onCopyLink}
                className="text-blue-500 hover:text-blue-700 underline text-xs"
              >
                Copy invite link
              </button>
            </div>
          </div>
        </div>

        {canRevoke && (
          <button
            onClick={onRevoke}
            className="px-3 py-1 bg-red-500 text-white text-sm rounded hover:bg-red-600"
          >
            Revoke
          </button>
        )}
      </div>
    </div>
  );
};

