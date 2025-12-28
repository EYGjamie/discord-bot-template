import { Link, useLocation } from 'react-router-dom';
import { useState } from 'react';
import { 
  LayoutDashboard, 
  Users, 
  Calendar, 
  ListTodo, 
  Trophy,
  MessageSquare,
  Bell,
  Settings,
  LogOut,
  Menu,
  X as CloseIcon,
  ChevronLeft,
  ChevronRight
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

interface SidebarProps {
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
}

export default function Sidebar({ isCollapsed = false, onToggleCollapse }: SidebarProps) {
  const location = useLocation();
  const { user, logout } = useAuth();
  const permissions = usePermissions(user);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  const handleLogout = () => {
    logout();
    window.location.href = '/login';
  };

  // Mobile overlay für kleine Bildschirme
  const MobileOverlay = () => (
    <>
      {isMobileOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={() => setIsMobileOpen(false)}
        />
      )}
    </>
  );

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={() => setIsMobileOpen(!isMobileOpen)}
        className="lg:hidden fixed top-4 left-4 z-50 p-2 bg-[#1a1f2e] border border-gray-800 rounded-lg text-white hover:bg-gray-800 transition-colors"
      >
        {isMobileOpen ? <CloseIcon className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
      </button>

      <MobileOverlay />

      {/* Sidebar */}
      <div className={cn(
        "bg-[#1a1f2e] h-screen fixed left-0 top-0 border-r border-gray-800 flex flex-col transition-all duration-300 z-50",
        // Mobile: Slide in from left
        "lg:translate-x-0",
        isMobileOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
        // Tablet & Desktop: Toggle between collapsed (icons only) and expanded
        isCollapsed ? "lg:w-20" : "lg:w-64",
        // Mobile always full width when open
        "w-64"
      )}>
      {/* Logo */}
      <div className={cn(
        "p-6 border-b border-gray-800 transition-all",
        isCollapsed && "lg:p-4"
      )}>
        <div className={cn(
          "flex items-center gap-3",
          isCollapsed && "lg:justify-center"
        )}>
          <div className="w-10 h-10 bg-cyan-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg className="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
            </svg>
          </div>
          {!isCollapsed && (
            <div className="lg:block hidden">
              <h1 className="text-white font-bold text-lg">ENTROPY</h1>
              <p className="text-gray-400 text-xs">GAMING</p>
            </div>
          )}
          {/* Mobile always shows text */}
          <div className="lg:hidden">
            <h1 className="text-white font-bold text-lg">ENTROPY</h1>
            <p className="text-gray-400 text-xs">GAMING</p>
          </div>
        </div>
      </div>

      {/* Desktop Toggle Button */}
      <button
        onClick={onToggleCollapse}
        className="hidden lg:flex absolute -right-3 top-20 w-6 h-6 bg-[#1a1f2e] border border-gray-800 rounded-full items-center justify-center text-gray-400 hover:text-white hover:bg-gray-800 transition-colors z-10"
      >
        {isCollapsed ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
      </button>

      {/* Navigation */}
      <nav className={cn(
        "p-4 space-y-1 flex-1 overflow-y-auto",
        isCollapsed && "lg:p-2"
      )}>
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
                onClick={() => setIsMobileOpen(false)}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-lg transition-colors group relative',
                  isActive 
                    ? 'bg-cyan-500/10 text-cyan-400' 
                    : 'text-gray-400 hover:bg-gray-800/50 hover:text-white',
                  isCollapsed && 'lg:justify-center lg:px-2'
                )}
                title={isCollapsed ? item.name : undefined}
              >
                <item.icon className="w-5 h-5 flex-shrink-0" />
                {!isCollapsed && (
                  <span className="font-medium lg:block hidden">{item.name}</span>
                )}
                {/* Mobile always shows text */}
                <span className="font-medium lg:hidden">{item.name}</span>
                
                {/* Tooltip for collapsed state */}
                {isCollapsed && (
                  <div className="hidden lg:block absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity whitespace-nowrap z-50">
                    {item.name}
                  </div>
                )}
              </Link>
            </PermissionGate>
          );
        })}
      </nav>

      {/* User Profile */}
      {user && (
        <div className={cn(
          "p-4 border-t border-gray-800",
          isCollapsed && "lg:p-2"
        )}>
          <Link
            to={`/members/${user.discord_id}`}
            onClick={() => setIsMobileOpen(false)}
            className={cn(
              "flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-gray-800/50 transition-colors group relative",
              isCollapsed && "lg:justify-center lg:px-2"
            )}
          >
            <div className="relative flex-shrink-0">
              <img
                src={user.avatar_url || `https://cdn.discordapp.com/embed/avatars/${parseInt(user.discord_id) % 5}.png`}
                alt={user.username}
                className="w-10 h-10 rounded-full"
              />
              <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-[#1a1f2e]"></div>
            </div>
            {!isCollapsed && (
              <div className="flex-1 min-w-0 lg:block hidden">
                <p className="text-white font-medium text-sm truncate group-hover:text-cyan-400 transition-colors">
                  {user.display_name || user.username}
                </p>
                <p className="text-gray-400 text-xs truncate">
                  {permissions.isAdmin ? '👑 Admin' : permissions.isModerator ? '🛡️ Moderator' : 'Member'}
                </p>
              </div>
            )}
            {/* Mobile always shows text */}
            <div className="flex-1 min-w-0 lg:hidden">
              <p className="text-white font-medium text-sm truncate group-hover:text-cyan-400 transition-colors">
                {user.display_name || user.username}
              </p>
              <p className="text-gray-400 text-xs truncate">
                {permissions.isAdmin ? '👑 Admin' : permissions.isModerator ? '🛡️ Moderator' : 'Member'}
              </p>
            </div>

            {/* Tooltip for collapsed state */}
            {isCollapsed && (
              <div className="hidden lg:block absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity whitespace-nowrap z-50">
                {user.display_name || user.username}
              </div>
            )}
          </Link>
          <button
            onClick={handleLogout}
            className={cn(
              "w-full flex items-center gap-3 px-3 py-2 mt-2 rounded-lg text-gray-400 hover:bg-red-500/10 hover:text-red-400 transition-colors group relative",
              isCollapsed && "lg:justify-center lg:px-2"
            )}
            title={isCollapsed ? "Logout" : undefined}
          >
            <LogOut className="w-5 h-5 flex-shrink-0" />
            {!isCollapsed && (
              <span className="font-medium text-sm lg:block hidden">Logout</span>
            )}
            {/* Mobile always shows text */}
            <span className="font-medium text-sm lg:hidden">Logout</span>

            {/* Tooltip for collapsed state */}
            {isCollapsed && (
              <div className="hidden lg:block absolute left-full ml-2 px-2 py-1 bg-gray-900 text-white text-sm rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity whitespace-nowrap z-50">
                Logout
              </div>
            )}
          </button>
        </div>
      )}
    </div>
    </>
  );
}
