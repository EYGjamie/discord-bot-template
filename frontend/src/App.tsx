import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './hooks/useAuth';
import LoginPage from './pages/LoginPage';
import AuthCallbackPage from './pages/AuthCallbackPage';
import DashboardPage from './pages/DashboardPage';
import MembersPage from './pages/MembersPage';
import UserProfilePage from './pages/UserProfilePage';
import DashboardLayout from './components/layout/DashboardLayout';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
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
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/auth/callback" element={<AuthCallbackPage />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardLayout>
                <DashboardPage />
              </DashboardLayout>
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/dashboard" />} />
        
        {/* Placeholder routes */}
        <Route path="/members" element={<ProtectedRoute><DashboardLayout><MembersPage /></DashboardLayout></ProtectedRoute>} />
        <Route path="/members/:userId" element={<ProtectedRoute><DashboardLayout><UserProfilePage /></DashboardLayout></ProtectedRoute>} />
        <Route path="/events" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Events Page</div></DashboardLayout></ProtectedRoute>} />
        <Route path="/tasks" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Tasks Page</div></DashboardLayout></ProtectedRoute>} />
        <Route path="/matches" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Matches Page</div></DashboardLayout></ProtectedRoute>} />
        <Route path="/discord" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Discord Page</div></DashboardLayout></ProtectedRoute>} />
        <Route path="/notifications" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Notifications Page</div></DashboardLayout></ProtectedRoute>} />
        <Route path="/settings" element={<ProtectedRoute><DashboardLayout><div className="p-6 text-white">Settings Page</div></DashboardLayout></ProtectedRoute>} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;