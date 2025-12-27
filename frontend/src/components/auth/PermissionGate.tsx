import React from 'react';
import type { User } from '../../types';
import { useHasPermission } from '../../hooks/usePermissions';

interface ProtectedRouteProps {
  user: User | null;
  requiredPermission: 'public' | 'member' | 'moderator' | 'admin';
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * Component to conditionally render children based on user permissions
 */
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  user,
  requiredPermission,
  children,
  fallback = <AccessDenied />,
}) => {
  const hasPermission = useHasPermission(user, requiredPermission);

  if (!hasPermission) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

/**
 * Default access denied component
 */
const AccessDenied: React.FC = () => {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="text-center">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Zugriff verweigert
        </h2>
        <p className="text-gray-600">
          Sie haben keine Berechtigung, diese Seite anzuzeigen.
        </p>
      </div>
    </div>
  );
};

interface PermissionGateProps {
  user: User | null;
  requiredPermission: 'public' | 'member' | 'moderator' | 'admin';
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

/**
 * Component to conditionally render UI elements based on permissions
 * Similar to ProtectedRoute but for inline elements
 */
export const PermissionGate: React.FC<PermissionGateProps> = ({
  user,
  requiredPermission,
  children,
  fallback = null,
}) => {
  const hasPermission = useHasPermission(user, requiredPermission);

  if (!hasPermission) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};
