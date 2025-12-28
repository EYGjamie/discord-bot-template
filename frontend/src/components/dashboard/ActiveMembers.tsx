import type { Member } from '../../types';

const mockMembers: Member[] = [
  { id: '1', name: 'alexstorm', display_name: 'Alex Storm', top_role_name: 'Team Captain', avatar_url: '', global_name: 'Alex Storm', bot: false, avatar: '', mention: '<@1>', created_at: '', nick: '', joined_at: '', top_role: '1', timed_out_until: undefined, premium_since: undefined, updated_at: '' },
  { id: '2', name: 'mayachen', display_name: 'Maya Chen', top_role_name: 'Support', avatar_url: '', global_name: 'Maya Chen', bot: false, avatar: '', mention: '<@2>', created_at: '', nick: '', joined_at: '', top_role: '2', timed_out_until: undefined, premium_since: undefined, updated_at: '' },
  { id: '3', name: 'jakewilson', display_name: 'Jake Wilson', top_role_name: 'Manager', avatar_url: '', global_name: 'Jake Wilson', bot: false, avatar: '', mention: '<@3>', created_at: '', nick: '', joined_at: '', top_role: '3', timed_out_until: undefined, premium_since: undefined, updated_at: '' },
  { id: '4', name: 'sarahkim', display_name: 'Sarah Kim', top_role_name: 'Coach', avatar_url: '', global_name: 'Sarah Kim', bot: false, avatar: '', mention: '<@4>', created_at: '', nick: '', joined_at: '', top_role: '4', timed_out_until: undefined, premium_since: undefined, updated_at: '' },
  { id: '5', name: 'nightowl', display_name: 'NightOwl', top_role_name: 'New Member', avatar_url: '', global_name: 'NightOwl', bot: false, avatar: '', mention: '<@5>', created_at: '', nick: '', joined_at: '', top_role: '5', timed_out_until: undefined, premium_since: undefined, updated_at: '' },
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
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4">Active Members</h2>
      <div className="space-y-3">
        {mockMembers.map((member) => (
          <div key={member.id} className="flex items-center gap-2 sm:gap-3">
            <div className="relative flex-shrink-0">
              <div className="w-9 h-9 sm:w-10 sm:h-10 rounded-full bg-gray-700 flex items-center justify-center text-white text-xs sm:text-sm font-medium">
                {member.display_name.split(' ').map(n => n[0]).join('')}
              </div>
              <div className="absolute bottom-0 right-0 w-2.5 h-2.5 sm:w-3 sm:h-3 bg-green-500 rounded-full border-2 border-[#1a1f2e]" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-white font-medium text-xs sm:text-sm truncate">{member.display_name}</p>
              <p className="text-gray-400 text-xs truncate">{member.top_role_name}</p>
            </div>
            <span className={`hidden sm:flex px-2 py-1 text-xs font-medium rounded border ${roleColors[member.top_role_name || ''] || 'bg-gray-500/10 text-gray-400 border-gray-500/20'} flex-shrink-0`}>
              {member.top_role_name?.split(' ')[0]}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
