import { useState, useEffect } from 'react';
import { Search, Filter, X } from 'lucide-react';
import UserCard from '../components/members/UserCard';
import { api } from '../services/api';
import type { Member } from '../types';

export default function MembersPage() {
  const [members, setMembers] = useState<Member[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [selectedRole, setSelectedRole] = useState<string>('all');
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [uniqueRoles, setUniqueRoles] = useState<string[]>([]);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);

  const ITEMS_PER_PAGE = 25;

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      setSearchQuery(searchInput);
      setCurrentPage(1);
    }, 500);

    return () => clearTimeout(timer);
  }, [searchInput]);

  // Fetch unique roles once on mount
  useEffect(() => {
    const fetchAllRoles = async () => {
      try {
        const data = await api.members.getMembers({
          page: 1,
          per_page: 100,
        });
        const roles = Array.from(new Set(data.members?.map((m: Member) => m.top_role_name).filter(Boolean)));
        setUniqueRoles(roles as string[]);
      } catch (error) {
        console.error('Failed to fetch roles:', error);
      }
    };
    fetchAllRoles();
  }, []);

  useEffect(() => {
    fetchMembers();
  }, [currentPage, searchQuery, selectedRole]);

  const fetchMembers = async () => {
    try {
      setLoading(true);
      const data = await api.members.getMembers({
        page: currentPage,
        per_page: ITEMS_PER_PAGE,
        search: searchQuery || undefined,
        role: selectedRole !== 'all' ? selectedRole : undefined,
      });

      setMembers(data.members || []);
      setTotal(data.total || 0);
      setTotalPages(data.total_pages || 1);
    } catch (error) {
      console.error('Failed to fetch members:', error);
      setMembers([]);
      setTotal(0);
      setTotalPages(1);
    } finally {
      setLoading(false);
    }
  };

  const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
  const endIndex = Math.min(startIndex + ITEMS_PER_PAGE, total);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading members...</div>
      </div>
    );
  }

  return (
    <div className="space-y-4 sm:space-y-6 p-4 sm:p-6 pt-16 lg:pt-4 sm:pt-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold text-white">Members</h1>
        <p className="text-gray-400 mt-1 text-sm sm:text-base">
          Showing {startIndex + 1}-{endIndex} of {total} members
        </p>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-3 sm:gap-4">
        {/* Search */}
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 sm:w-5 sm:h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search by name..."
            value={searchInput}
            onChange={e => setSearchInput(e.target.value)}
            className="w-full pl-9 sm:pl-10 pr-10 py-2 sm:py-2.5 bg-[#1a1f2e] border border-gray-700 rounded-lg text-white text-sm sm:text-base placeholder-gray-400 focus:outline-none focus:border-cyan-500"
          />
          {searchInput && (
            <button
              onClick={() => setSearchInput('')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        {/* Role Filter */}
        <div className="relative">
          <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 sm:w-5 sm:h-5 text-gray-400" />
          <select
            value={selectedRole}
            onChange={e => {
              setSelectedRole(e.target.value);
              setCurrentPage(1);
            }}
            className="w-full pl-9 sm:pl-10 pr-10 py-2 sm:py-2.5 bg-[#1a1f2e] border border-gray-700 rounded-lg text-white text-sm sm:text-base focus:outline-none focus:border-cyan-500 appearance-none cursor-pointer"
          >
            <option value="all">All Roles</option>
            {uniqueRoles.map(role => (
              <option key={role} value={role}>
                {role}
              </option>
            ))}
          </select>
          {selectedRole !== 'all' && (
            <button
              onClick={() => setSelectedRole('all')}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      {/* Members Grid */}
      {members.length === 0 ? (
        <div className="text-center py-12 text-gray-400 text-sm sm:text-base">
          No members found matching your filters.
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 2xl:grid-cols-7 gap-3 sm:gap-4">
          {members.map(member => (
            <UserCard key={member.id} member={member} />
          ))}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex flex-col sm:flex-row justify-center items-center gap-3 sm:gap-2">
          <button
            onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
            disabled={currentPage === 1}
            className="w-full sm:w-auto px-4 py-2 bg-[#1a1f2e] border border-gray-700 rounded-lg text-white text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:border-cyan-500 transition-colors"
          >
            Previous
          </button>

          <div className="flex gap-2 overflow-x-auto pb-2 sm:pb-0 w-full sm:w-auto justify-center">
            {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
              let pageNum;
              if (totalPages <= 5) {
                pageNum = i + 1;
              } else if (currentPage <= 3) {
                pageNum = i + 1;
              } else if (currentPage >= totalPages - 2) {
                pageNum = totalPages - 4 + i;
              } else {
                pageNum = currentPage - 2 + i;
              }

              return (
                <button
                  key={pageNum}
                  onClick={() => setCurrentPage(pageNum)}
                  className={`w-10 h-10 flex-shrink-0 rounded-lg transition-colors text-sm ${
                    currentPage === pageNum
                      ? 'bg-cyan-500 text-white'
                      : 'bg-[#1a1f2e] border border-gray-700 text-white hover:border-cyan-500'
                  }`}
                >
                  {pageNum}
                </button>
              );
            })}
          </div>

          <button
            onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
            disabled={currentPage === totalPages}
            className="w-full sm:w-auto px-4 py-2 bg-[#1a1f2e] border border-gray-700 rounded-lg text-white text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:border-cyan-500 transition-colors"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
}
