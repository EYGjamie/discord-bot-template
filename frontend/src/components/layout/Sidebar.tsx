import { Link, useLocation } from 'react-router-dom';
import { 
  LayoutDashboard, 
  Users, 
  Calendar, 
  ListTodo, 
  Trophy,
  MessageSquare,
  Bell,
  Settings 
} from 'lucide-react';
import { cn } from '../../utils/cn';

const navigation = [
  { name: 'Dashboard', icon: LayoutDashboard, href: '/dashboard' },
  { name: 'Members', icon: Users, href: '/members' },
  { name: 'Events', icon: Calendar, href: '/events' },
  { name: 'Tasks', icon: ListTodo, href: '/tasks' },
  { name: 'Matches', icon: Trophy, href: '/matches' },
  { name: 'Discord', icon: MessageSquare, href: '/discord' },
  { name: 'Notifications', icon: Bell, href: '/notifications' },
  { name: 'Settings', icon: Settings, href: '/settings' },
];

export default function Sidebar() {
  const location = useLocation();

  return (
    <div className="w-64 bg-[#1a1f2e] h-screen fixed left-0 top-0 border-r border-gray-800">
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
      <nav className="p-4 space-y-1">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href;
          return (
            <Link
              key={item.name}
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
          );
        })}
      </nav>
    </div>
  );
}
