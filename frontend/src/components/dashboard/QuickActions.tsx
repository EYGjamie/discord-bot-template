import { Plus, Calendar, RefreshCw, MessageSquare } from 'lucide-react';

const actions = [
  { 
    icon: Plus, 
    title: 'Create Task', 
    description: 'Add a new task for the team',
    color: 'bg-cyan-500'
  },
  { 
    icon: MessageSquare, 
    title: 'Add Member', 
    description: 'Invite new members to the org',
    color: 'bg-purple-500'
  },
  { 
    icon: Calendar, 
    title: 'Schedule Event', 
    description: 'Plan scrims, meetings, or tournaments',
    color: 'bg-green-500'
  },
  { 
    icon: RefreshCw, 
    title: 'Sync Discord', 
    description: 'Update member data from Discord',
    color: 'bg-blue-500'
  },
];

export default function QuickActions() {
  return (
    <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
      <h2 className="text-white text-lg font-semibold mb-4">Quick Actions</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {actions.map((action) => (
          <button
            key={action.title}
            className="flex items-start gap-3 p-4 rounded-lg border border-gray-700 hover:border-gray-600 hover:bg-gray-800/50 transition-colors text-left"
          >
            <div className={`p-2 rounded-lg ${action.color}`}>
              <action.icon className="w-5 h-5 text-white" />
            </div>
            <div>
              <h3 className="text-white font-medium mb-1">{action.title}</h3>
              <p className="text-gray-400 text-sm">{action.description}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
