import { Calendar, Clock, Users } from 'lucide-react';
import type { Event } from '../../types';

const mockEvents: Event[] = [
  {
    id: '1',
    title: 'Weekly Scrim vs Team Alpha',
    description: 'Valorant',
    date: 'Dec 24',
    time: '6:00 PM',
    platform: 'Valorant',
    participants: 10,
    status: 'upcoming'
  },
  {
    id: '2',
    title: 'Team Meeting',
    description: 'Discord',
    date: 'Dec 25',
    time: '6:00 PM',
    platform: 'Discord',
    participants: 15,
    status: 'upcoming'
  },
  {
    id: '3',
    title: 'Tournament Qualifier',
    description: 'Valorant',
    date: 'Dec 28',
    time: '3:00 PM',
    platform: 'Valorant',
    participants: 5,
    status: 'confirmed'
  },
];

const statusColors = {
  upcoming: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  confirmed: 'bg-green-500/10 text-green-400 border-green-500/20',
  completed: 'bg-gray-500/10 text-gray-400 border-gray-500/20',
};

export default function UpcomingEvents() {
  return (
    <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
      <h2 className="text-white text-lg font-semibold mb-4">Upcoming Events</h2>
      <div className="space-y-3">
        {mockEvents.map((event) => (
          <div
            key={event.id}
            className="p-4 rounded-lg border border-gray-700 hover:border-gray-600 transition-colors"
          >
            <div className="flex items-start justify-between mb-2">
              <h3 className="text-white font-medium">{event.title}</h3>
              <span className={`px-2 py-1 text-xs font-medium rounded border ${statusColors[event.status]}`}>
                {event.status}
              </span>
            </div>
            <p className="text-gray-400 text-sm mb-3">{event.description}</p>
            <div className="flex items-center gap-4 text-gray-500 text-sm">
              <div className="flex items-center gap-1">
                <Calendar className="w-4 h-4" />
                <span>{event.date}</span>
              </div>
              <div className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                <span>{event.time}</span>
              </div>
              <div className="flex items-center gap-1">
                <Users className="w-4 h-4" />
                <span>{event.participants}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
