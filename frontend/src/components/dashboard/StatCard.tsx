import type { LucideIcon } from 'lucide-react';

interface StatCardProps {
  title: string;
  value: string | number;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon: LucideIcon;
  iconColor: string;
}

export default function StatCard({ title, value, change, changeType = 'neutral', icon: Icon, iconColor }: StatCardProps) {
  return (
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <p className="text-gray-400 text-xs sm:text-sm mb-1">{title}</p>
          <h3 className="text-white text-2xl sm:text-3xl font-bold mb-1 sm:mb-2">{value}</h3>
          {change && (
            <p className={`text-xs sm:text-sm ${
              changeType === 'positive' ? 'text-green-400' : 
              changeType === 'negative' ? 'text-red-400' : 
              'text-gray-400'
            }`}>
              {change}
            </p>
          )}
        </div>
        <div className={`p-2 sm:p-3 rounded-lg ${iconColor} flex-shrink-0`}>
          <Icon className="w-5 h-5 sm:w-6 sm:h-6 text-white" />
        </div>
      </div>
    </div>
  );
}
