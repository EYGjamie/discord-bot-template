import { Link, useLocation } from 'react-router-dom';
import { 
  LayoutDashboard, 
  Users, 
  Calendar, 
  ListTodo, 
  Trophy,
  MessageSquare,
  Bell,
  Settings,
  LogOut
} from 'lucide-react';
import { cn } from '../../utils/cn';
import { useAuth } from '../../hooks/useAuth';
import { usePermissions } from '../../hooks/usePermissions';
import { PermissionGate } from '../auth/PermissionGate';

interface NavItem {
  name: string;
  icon: any;
  href: string;
  requiredPermission?: 'public' | 'member' | 'moderator' | 'admin';
}

const navigation: NavItem[] = [
  { name: 'Dashboard', icon: LayoutDashboard, href: '/dashboard', requiredPermission: 'member' },
  { name: 'Members', icon: Users, href: '/members', requiredPermission: 'moderator' },
  { name: 'Events', icon: Calendar, href: '/events', requiredPermission: 'member' },
  { name: 'Tasks', icon: ListTodo, href: '/tasks', requiredPermission: 'member' },
  { name: 'Matches', icon: Trophy, href: '/matches', requiredPermission: 'member' },
  { name: 'Discord', icon: MessageSquare, href: '/discord', requiredPermission: 'member' },
  { name: 'Notifications', icon: Bell, href: '/notifications', requiredPermission: 'member' },
  { name: 'Settings', icon: Settings, href: '/settings', requiredPermission: 'admin' },
];

export default function Sidebar() {
  const location = useLocation();
  const { user, logout } = useAuth();
  const permissions = usePermissions(user);

  const handleLogout = () => {
    logout();
    window.location.href = '/login';
  };

  return (
    <div className="w-64 bg-[#1a1f2e] h-screen fixed left-0 top-0 border-r border-gray-800 flex flex-col">
      {/* Logo */}
      <div className="p-6 border-b border-gray-800">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-cyan-500 rounded-lg flex items-center justify-center">
            <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
            </svg>
          </div>
          <div>
            <h1 className="text-white font-bold text-lg">ENTROPY</h1>
            <p className="text-gray-400 text-xs">GAMING</p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="p-4 space-y-1 flex-1">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href;
          return (
            <PermissionGate 
              key={item.name}
              user={user}
              requiredPermission={item.requiredPermission || 'public'}
            >
              <Link
                to={item.href}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors',
                  isActive 
                    ? 'bg-cyan-500/10 text-cyan-400' 
                    : 'text-gray-400 hover:bg-gray-800/50 hover:text-white'
                )}
              >
                <item.icon className="w-5 h-5" />
                <span className="font-medium">{item.name}</span>
              </Link>
            </PermissionGate>
          );
        })}
      </nav>

      {/* User Profile */}
      {user && (
        <div className="p-4 border-t border-gray-800">
          <Link
            to={`/members/${user.discord_id}`}
            className="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-800/50 transition-colors group"
          >
            <div className="relative">
              <img
                src={user.avatar_url || `https://cdn.discordapp.com/embed/avatars/${parseInt(user.discord_id) % 5}.png`}
                alt={user.username}
                className="w-10 h-10 rounded-full"
              />
              <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-[#1a1f2e]"></div>
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white font-medium text-sm truncate group-hover:text-cyan-400 transition-colors">
                {user.display_name || user.username}
              </p>
              <p className="text-gray-400 text-xs truncate">
                {permissions.isAdmin ? '👑 Admin' : permissions.isModerator ? '🛡️ Moderator' : 'Member'}
              </p>
            </div>
          </Link>
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-3 py-2 mt-2 rounded-lg text-gray-400 hover:bg-red-500/10 hover:text-red-400 transition-colors"
          >
            <LogOut className="w-5 h-5" />
            <span className="font-medium text-sm">Logout</span>
          </button>
        </div>
      )}
    </div>
  );
}
