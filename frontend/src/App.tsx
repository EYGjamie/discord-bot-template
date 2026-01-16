import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import DashboardPage from './pages/DashboardPage';
import MembersPage from './pages/MembersPage';
import UserProfilePage from './pages/UserProfilePage';
import EventsPage from './pages/EventsPage';
import MatchesPage from './pages/MatchesPage';
import DiscordPage from './pages/DiscordPage';
import SettingsPage from './pages/SettingsPage';
import BoardsListPage from './pages/BoardsListPage';
import CreateBoardPage from './pages/CreateBoardPage';
import KanbanBoardPage from './pages/KanbanBoardPage';
import TaskGroupsPage from './pages/TaskGroupsPage';
import GroupPermissionsPage from './pages/GroupPermissionsPage';
import DashboardLayout from './components/layout/DashboardLayout';
import { ProtectedRoute } from './components/auth/PermissionGate';

function AuthProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0f1419] flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  return isAuthenticated ? <>{children}</> : <Navigate to="/login" />;
}

function App() {
  const { user } = useAuth();
  
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route
          path="/dashboard"
          element={
            <AuthProtectedRoute>
              <DashboardLayout>
                <DashboardPage />
              </DashboardLayout>
            </AuthProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/dashboard" />} />
        
        {/* Member routes - requires moderator permission */}
        <Route 
          path="/members" 
          element={
            <AuthProtectedRoute>
              <DashboardLayout>
                <ProtectedRoute user={user} requiredPermission="moderator">
                  <MembersPage />
                </ProtectedRoute>
              </DashboardLayout>
            </AuthProtectedRoute>
          } 
        />
        <Route 
          path="/members/:userId" 
          element={
            <AuthProtectedRoute>
              <DashboardLayout>
                <ProtectedRoute user={user} requiredPermission="moderator">
                  <UserProfilePage />
                </ProtectedRoute>
              </DashboardLayout>
            </AuthProtectedRoute>
          } 
        />
        
        {/* Public routes for guild members */}
        <Route path="/events" element={<AuthProtectedRoute><DashboardLayout><EventsPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/tasks" element={<AuthProtectedRoute><DashboardLayout><BoardsListPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/tasks/boards/new" element={<AuthProtectedRoute><DashboardLayout><CreateBoardPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/tasks/boards/:boardId" element={<AuthProtectedRoute><DashboardLayout><KanbanBoardPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/tasks/groups" element={<AuthProtectedRoute><DashboardLayout><TaskGroupsPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/tasks/groups/:groupId/permissions" element={<AuthProtectedRoute><DashboardLayout><GroupPermissionsPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route path="/matches" element={<AuthProtectedRoute><DashboardLayout><MatchesPage /></DashboardLayout></AuthProtectedRoute>} />
        <Route 
          path="/discord" 
          element={
            <AuthProtectedRoute>
              <DashboardLayout>
                <ProtectedRoute user={user} requiredPermission="moderator">
                  <DiscordPage />
                </ProtectedRoute>
              </DashboardLayout>
            </AuthProtectedRoute>
          } 
        />
        <Route path="/notifications" element={<AuthProtectedRoute><DashboardLayout><div className="p-6 text-white">Notifications Page</div></DashboardLayout></AuthProtectedRoute>} />
        
        {/* Settings - requires admin permission */}
        <Route 
          path="/settings" 
          element={
            <AuthProtectedRoute>
              <DashboardLayout>
                <ProtectedRoute user={user} requiredPermission="admin">
                  <SettingsPage />
                </ProtectedRoute>
              </DashboardLayout>
            </AuthProtectedRoute>
          } 
        />
      </Routes>
    </BrowserRouter>
  );
}

export default App;