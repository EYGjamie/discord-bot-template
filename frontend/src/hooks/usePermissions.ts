import { useMemo } from 'react';
import type { User, PermissionCheck } from '../types';

/**
 * Hook to check user permissions based on their roles
 * @param user - The current authenticated user
 * @returns Permission check object with boolean flags for various permissions
 */
export const usePermissions = (user: User | null): PermissionCheck => {
  return useMemo(() => {
    // Default permissions (no access)
    if (!user) {
      return {
        canViewMembers: false,
        canModerate: false,
        canViewAuditLogs: false,
        canManageSettings: false,
        isAdmin: false,
        isModerator: false,
        isMember: false,
      };
    }

    const isAdmin = user.is_admin || false;
    const isModerator = user.is_moderator || false;
    const isMember = (user.roles && user.roles.length > 0) || false;

    return {
      // Member data viewing requires moderator role
      canViewMembers: isModerator || isAdmin,
      
      // Moderation actions require moderator role
      canModerate: isModerator || isAdmin,
      
      // Audit logs require admin role
      canViewAuditLogs: isAdmin,
      
      // Settings management requires admin role
      canManageSettings: isAdmin,
      
      // Role flags
      isAdmin,
      isModerator,
      isMember,
    };
  }, [user]);
};

/**
 * Hook to check if user has a specific permission level
 * @param user - The current authenticated user
 * @param requiredLevel - The required permission level
 * @returns Boolean indicating if user has the required permission
 */
export const useHasPermission = (
  user: User | null,
  requiredLevel: 'public' | 'member' | 'moderator' | 'admin'
): boolean => {
  return useMemo(() => {
    if (requiredLevel === 'public') {
      return true;
    }

    if (!user) {
      return false;
    }

    switch (requiredLevel) {
      case 'member':
        return (user.roles && user.roles.length > 0) || false;
      case 'moderator':
        return user.is_moderator || user.is_admin || false;
      case 'admin':
        return user.is_admin || false;
      default:
        return false;
    }
  }, [user, requiredLevel]);
};
