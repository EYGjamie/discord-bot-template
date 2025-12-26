import { Link } from 'react-router-dom';
import type { Member } from '../../types';

interface UserCardProps {
  member: Member;
}

export default function UserCard({ member }: UserCardProps) {
  const getAvatarUrl = () => {
    if (member.avatar_url) return member.avatar_url;
    // Fallback to default Discord avatar
    return `https://cdn.discordapp.com/embed/avatars/${parseInt(member.id) % 5}.png`;
  };

  return (
    <Link
      to={`/members/${member.id}`}
      className="bg-[#1a1f2e] rounded-lg p-4 border border-gray-800 hover:border-cyan-500/50 transition-all hover:scale-105 flex flex-col items-center gap-3"
    >
      {/* Avatar */}
      <div className="relative">
        <img
          src={getAvatarUrl()}
          alt={member.display_name}
          className="w-16 h-16 rounded-full border-2 border-gray-700"
        />
        {member.bot && (
          <div className="absolute -bottom-1 -right-1 bg-cyan-500 text-white text-xs px-1.5 py-0.5 rounded font-bold">
            BOT
          </div>
        )}
      </div>

      {/* Display Name */}
      <div className="text-center">
        <h3 className="text-white font-medium truncate w-full">
          {member.display_name || member.name}
        </h3>
        {member.nick && member.nick !== member.display_name && (
          <p className="text-gray-400 text-xs truncate w-full">@{member.name}</p>
        )}
      </div>

      {/* Top Role */}
      {member.top_role_name && (
        <div
          className="px-3 py-1 rounded-full text-xs font-medium"
          style={{
            backgroundColor: member.top_role_color
              ? `${member.top_role_color}20`
              : '#4b556320',
            color: member.top_role_color || '#9ca3af',
            borderWidth: '1px',
            borderColor: member.top_role_color
              ? `${member.top_role_color}40`
              : '#4b556340',
          }}
        >
          {member.top_role_name}
        </div>
      )}
    </Link>
  );
}
