# Frontend Implementation Files

These files implement caching, auth persistence, and lazy loading for your React frontend.

## Setup Instructions

1. **Install Dependencies**
   ```bash
   npm install @tanstack/react-query @tanstack/react-query-persist-client axios
   ```

2. **Copy Files to Your Frontend Repository**
   - Copy `src/lib/queryClient.ts` to your `src/lib/` directory
   - Copy `src/lib/apiClient.ts` to your `src/lib/` directory  
   - Copy `src/contexts/AuthContext.tsx` to your `src/contexts/` directory
   - Copy `src/hooks/useAuth.ts` to your `src/hooks/` directory
   - Copy `src/hooks/useProjects.ts` to your `src/hooks/` directory

3. **Update Your App.tsx**
   - Use the example in `src/App.example.tsx` as a reference
   - Wrap your app with `QueryClientProvider` and `AuthProvider`
   - Implement lazy loading for routes

4. **Update Your index.tsx**
   - See `src/index.example.tsx` for reference
   - Ensure `QueryClientProvider` wraps your App

5. **Environment Variables**
   Add to your `.env` file:
   ```
   REACT_APP_API_URL=https://devhive-go-backend.fly.dev/api/v1
   ```

## Usage Examples

### Using Projects Hook
```typescript
import { useProjects, useProject, useCreateProject } from './hooks/useProjects';

function ProjectsPage() {
  const { data, isLoading } = useProjects();
  const createProject = useCreateProject();

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      {data?.projects.map(project => (
        <div key={project.id}>{project.name}</div>
      ))}
    </div>
  );
}
```

### Using Auth Hook
```typescript
import { useAuth } from './hooks/useAuth';

function LoginPage() {
  const { login, isAuthenticated } = useAuth();

  if (isAuthenticated) {
    return <Navigate to="/projects" />;
  }

  const handleLogin = async () => {
    await login(username, password);
  };

  return <button onClick={handleLogin}>Login</button>;
}
```

## Key Features

- **Caching**: API responses are cached for 5 minutes (configurable per query)
- **No Refetching**: Data doesn't refetch on window focus, reconnect, or remount
- **Persistent Auth**: Users stay logged in across browser sessions via refresh tokens
- **Lazy Loading**: Routes are code-split for faster initial load
- **Automatic Token Refresh**: Access tokens refresh automatically on 401 errors



