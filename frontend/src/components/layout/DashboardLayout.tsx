import { useState } from 'react';
import type { ReactNode } from 'react';
import Sidebar from './Sidebar';

interface DashboardLayoutProps {
  children: ReactNode;
}

export default function DashboardLayout({ children }: DashboardLayoutProps) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <div className="flex min-h-screen bg-[#0f1419]">
      <Sidebar 
        isCollapsed={isSidebarCollapsed} 
        onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)} 
      />
      <main className={`
        flex-1 
        transition-all 
        duration-300
        lg:ml-64
        ${isSidebarCollapsed ? 'lg:ml-20' : 'lg:ml-64'}
        min-h-screen
      `}>
        {children}
      </main>
    </div>
  );
}
