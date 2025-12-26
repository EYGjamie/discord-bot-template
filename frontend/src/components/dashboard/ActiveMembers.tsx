import type { Member } from '../../types';

const mockMembers: Member[] = [
  { id: '1', name: 'Alex Storm', role: 'Team Captain', status: 'online' },
  { id: '2', name: 'Maya Chen', role: 'Support', status: 'online' },
  { id: '3', name: 'Jake Wilson', role: 'Manager', status: 'online' },
  { id: '4', name: 'Sarah Kim', role: 'Coach', status: 'online' },
  { id: '5', name: 'NightOwl', role: 'New Member', status: 'online' },
];

const roleColors: Record<string, string> = {
  'Team Captain': 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
  'Support': 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  'Manager': 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  'Coach': 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
  'New Member': 'bg-blue-500/10 text-blue-400 border-blue-500/20',
};

export default function ActiveMembers() {
  return (
    <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
      <h2 className="text-white text-lg font-semibold mb-4">Active Members</h2>
      <div className="space-y-3">
        {mockMembers.map((member) => (
          <div key={member.id} className="flex items-center gap-3">
            <div className="relative">
              <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center text-white text-sm font-medium">
                {member.name.split(' ').map(n => n[0]).join('')}
              </div>
              {member.status === 'online' && (
                <div className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 rounded-full border-2 border-[#1a1f2e]" />
              )}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white font-medium text-sm">{member.name}</p>
              <p className="text-gray-400 text-xs">{member.role}</p>
            </div>
            <span className={`px-2 py-1 text-xs font-medium rounded border ${roleColors[member.role] || 'bg-gray-500/10 text-gray-400 border-gray-500/20'}`}>
              {member.role.split(' ')[0]}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
