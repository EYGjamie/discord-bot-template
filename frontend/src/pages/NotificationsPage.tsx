import { useState, useEffect } from 'react';
import { Bell, ChevronDown, ChevronUp } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

interface NotificationSettings {
  global: {
    id: number;
    user_id: string;
    guild_id: string;
    task_notifications_enabled: boolean;
    notify_on_assignment: boolean;
    notify_on_task_update: boolean;
    notify_on_comment: boolean;
    notify_on_due_date_change: boolean;
    notify_on_unassignment: boolean;
    notify_on_checklist_item: boolean;
    created_at: string;
    updated_at: string;
  };
  boards: BoardNotificationSetting[];
}

interface BoardNotificationSetting {
  id: number;
  user_id: string;
  board_id: number;
  notifications_enabled: boolean;
  notify_on_assignment: boolean;
  notify_on_task_update: boolean;
  notify_on_comment: boolean;
  notify_on_due_date_change: boolean;
  notify_on_unassignment: boolean;
  notify_on_checklist_item: boolean;
  created_at: string;
  updated_at: string;
}

interface Board {
  id: number;
  name: string;
  description: string;
  color: string;
}

export default function NotificationsPage() {
  const { user } = useAuth();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<NotificationSettings | null>(null);
  const [boards, setBoards] = useState<Board[]>([]);
  const [expandedBoards, setExpandedBoards] = useState<Set<number>>(new Set());
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (user) {
      fetchData();
    }
  }, [user]);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);

      const token = localStorage.getItem('token');
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      // Fetch notification settings
      const settingsRes = await fetch('/api/notification-settings', {
        credentials: 'include',
        headers,
      });
      if (!settingsRes.ok) {
        const errorText = await settingsRes.text();
        console.error('Settings API error:', settingsRes.status, errorText);
        throw new Error(`Failed to fetch notification settings: ${settingsRes.status}`);
      }
      const settingsData = await settingsRes.json();

      // Fetch boards
      const boardsRes = await fetch('/api/boards', {
        credentials: 'include',
        headers,
      });
      if (!boardsRes.ok) throw new Error('Failed to fetch boards');
      const boardsData = await boardsRes.json();

      // Set both states together to avoid race conditions
      setSettings({
        ...settingsData,
        boards: Array.isArray(settingsData.boards) ? settingsData.boards : []
      });
      setBoards(boardsData);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError('Fehler beim Laden der Benachrichtigungseinstellungen');
    } finally {
      setLoading(false);
    }
  };

  const updateGlobalSettings = async (updates: Partial<NotificationSettings['global']>) => {
    if (!settings) return;

    try {
      setSaving(true);
      setError(null);

      const updatedSettings = {
        task_notifications_enabled: updates.task_notifications_enabled ?? settings.global.task_notifications_enabled,
        notify_on_assignment: updates.notify_on_assignment ?? settings.global.notify_on_assignment,
        notify_on_task_update: updates.notify_on_task_update ?? settings.global.notify_on_task_update,
        notify_on_comment: updates.notify_on_comment ?? settings.global.notify_on_comment,
        notify_on_due_date_change: updates.notify_on_due_date_change ?? settings.global.notify_on_due_date_change,
        notify_on_unassignment: updates.notify_on_unassignment ?? settings.global.notify_on_unassignment,
        notify_on_checklist_item: updates.notify_on_checklist_item ?? settings.global.notify_on_checklist_item,
      };

      const token = localStorage.getItem('token');
      const headers: HeadersInit = { 'Content-Type': 'application/json' };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const res = await fetch('/api/notification-settings', {
        method: 'PUT',
        headers,
        credentials: 'include',
        body: JSON.stringify(updatedSettings),
      });

      if (!res.ok) throw new Error('Failed to update settings');
      const data = await res.json();

      setSettings({
        ...settings,
        global: data,
      });
    } catch (err) {
      console.error('Error updating global settings:', err);
      setError('Fehler beim Speichern der Einstellungen');
    } finally {
      setSaving(false);
    }
  };

  const updateBoardSettings = async (boardId: number, updates: Partial<BoardNotificationSetting>) => {
    if (!settings) return;

    try {
      setSaving(true);
      setError(null);

      // Get current board settings or use defaults
      const currentBoardSettings = settings.boards.find(b => b.board_id === boardId);
      const updatedBoardSettings = {
        notifications_enabled: updates.notifications_enabled ?? currentBoardSettings?.notifications_enabled ?? true,
        notify_on_assignment: updates.notify_on_assignment ?? currentBoardSettings?.notify_on_assignment ?? true,
        notify_on_task_update: updates.notify_on_task_update ?? currentBoardSettings?.notify_on_task_update ?? true,
        notify_on_comment: updates.notify_on_comment ?? currentBoardSettings?.notify_on_comment ?? true,
        notify_on_due_date_change: updates.notify_on_due_date_change ?? currentBoardSettings?.notify_on_due_date_change ?? true,
        notify_on_unassignment: updates.notify_on_unassignment ?? currentBoardSettings?.notify_on_unassignment ?? true,
        notify_on_checklist_item: updates.notify_on_checklist_item ?? currentBoardSettings?.notify_on_checklist_item ?? true,
      };

      const token = localStorage.getItem('token');
      const headers: HeadersInit = { 'Content-Type': 'application/json' };
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const res = await fetch(`/api/notification-settings/boards/${boardId}`, {
        method: 'PUT',
        headers,
        credentials: 'include',
        body: JSON.stringify(updatedBoardSettings),
      });

      if (!res.ok) throw new Error('Failed to update board settings');
      const data = await res.json();

      // Update or add board settings
      const existingIndex = settings.boards.findIndex(b => b.board_id === boardId);
      const newBoards = [...settings.boards];
      if (existingIndex >= 0) {
        newBoards[existingIndex] = data;
      } else {
        newBoards.push(data);
      }

      setSettings({
        ...settings,
        boards: newBoards,
      });
    } catch (err) {
      console.error('Error updating board settings:', err);
      setError('Fehler beim Speichern der Board-Einstellungen');
    } finally {
      setSaving(false);
    }
  };

  const toggleBoardExpanded = (boardId: number) => {
    const newExpanded = new Set(expandedBoards);
    if (newExpanded.has(boardId)) {
      newExpanded.delete(boardId);
    } else {
      newExpanded.add(boardId);
    }
    setExpandedBoards(newExpanded);
  };

  const getBoardSettings = (boardId: number): BoardNotificationSetting | undefined => {
    if (!settings || !settings.boards) return undefined;
    return settings.boards.find(b => b.board_id === boardId);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!settings || !settings.boards) {
    return (
      <div className="p-8 text-center">
        <p className="text-red-500">Fehler beim Laden der Benachrichtigungseinstellungen</p>
        {error && <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">{error}</p>}
        <button
          onClick={fetchData}
          className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
        >
          Erneut versuchen
        </button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-4xl mx-auto p-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <div className="flex items-center gap-3 mb-6">
            <Bell className="w-8 h-8 text-blue-500" />
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
              Benachrichtigungen
            </h1>
          </div>

          {error && (
            <div className="mb-4 p-4 bg-red-100 dark:bg-red-900/20 text-red-700 dark:text-red-400 rounded-lg">
              {error}
            </div>
          )}

          {/* Global Settings */}
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Globale Einstellungen
            </h2>
            <div className="space-y-4 bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium text-gray-900 dark:text-white">
                    Task-Benachrichtigungen
                  </h3>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Alle Benachrichtigungen für Tasks aktivieren/deaktivieren
                  </p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.global.task_notifications_enabled}
                    onChange={(e) => updateGlobalSettings({ task_notifications_enabled: e.target.checked })}
                    disabled={saving}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className={`space-y-4 ${!settings.global.task_notifications_enabled ? 'opacity-50' : ''}`}>
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Zuweisung benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn dir eine Aufgabe zugewiesen wird
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_assignment}
                      onChange={(e) => updateGlobalSettings({ notify_on_assignment: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Bearbeitung benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn jemand anderes deine zugewiesenen Tasks bearbeitet
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_task_update}
                      onChange={(e) => updateGlobalSettings({ notify_on_task_update: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Kommentaren benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn jemand einen Kommentar zu deiner Task hinzufügt
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_comment}
                      onChange={(e) => updateGlobalSettings({ notify_on_comment: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Fälligkeitsdatum-Änderung benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn das Fälligkeitsdatum deiner Task geändert wird
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_due_date_change}
                      onChange={(e) => updateGlobalSettings({ notify_on_due_date_change: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Zuweisung entfernt benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn du von einer Task entfernt wirst
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_unassignment}
                      onChange={(e) => updateGlobalSettings({ notify_on_unassignment: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>

                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium text-gray-900 dark:text-white">
                      Bei Checklisten-Element benachrichtigen
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Erhalte eine DM, wenn ein Checklisten-Element zu deiner Task hinzugefügt wird
                    </p>
                  </div>
                  <label className="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.global.notify_on_checklist_item}
                      onChange={(e) => updateGlobalSettings({ notify_on_checklist_item: e.target.checked })}
                      disabled={saving || !settings.global.task_notifications_enabled}
                      className="sr-only peer"
                    />
                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                  </label>
                </div>
              </div>
            </div>
          </div>

          {/* Board-Specific Settings */}
          <div>
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
              Board-spezifische Einstellungen
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              Überschreibe die globalen Einstellungen für einzelne Boards
            </p>
            <div className="space-y-3">
              {boards.map((board) => {
                const boardSettings = getBoardSettings(board.id);
                const isExpanded = expandedBoards.has(board.id);
                const isEnabled = boardSettings?.notifications_enabled ?? true;

                return (
                  <div
                    key={board.id}
                    className="bg-gray-50 dark:bg-gray-700/50 rounded-lg overflow-hidden"
                  >
                    <div
                      className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                      onClick={() => toggleBoardExpanded(board.id)}
                    >
                      <div className="flex items-center gap-3">
                        <div
                          className="w-4 h-4 rounded"
                          style={{ backgroundColor: board.color }}
                        ></div>
                        <span className="font-medium text-gray-900 dark:text-white">
                          {board.name}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className={`text-sm ${isEnabled ? 'text-green-600 dark:text-green-400' : 'text-gray-500'}`}>
                          {isEnabled ? 'Aktiviert' : 'Deaktiviert'}
                        </span>
                        {isExpanded ? (
                          <ChevronUp className="w-5 h-5 text-gray-500" />
                        ) : (
                          <ChevronDown className="w-5 h-5 text-gray-500" />
                        )}
                      </div>
                    </div>

                    {isExpanded && (
                      <div className="p-4 border-t border-gray-200 dark:border-gray-600 space-y-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h3 className="font-medium text-gray-900 dark:text-white">
                              Board-Benachrichtigungen
                            </h3>
                            <p className="text-sm text-gray-600 dark:text-gray-400">
                              Benachrichtigungen für dieses Board aktivieren
                            </p>
                          </div>
                          <label className="relative inline-flex items-center cursor-pointer">
                            <input
                              type="checkbox"
                              checked={isEnabled}
                              onChange={(e) =>
                                updateBoardSettings(board.id, { notifications_enabled: e.target.checked })
                              }
                              disabled={saving}
                              className="sr-only peer"
                            />
                            <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                          </label>
                        </div>

                        <div className={`space-y-4 ${!isEnabled ? 'opacity-50' : ''}`}>
                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Zuweisung
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen für Zuweisungen in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_assignment ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_assignment: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Bearbeitung
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen für Bearbeitungen in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_task_update ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_task_update: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Kommentaren
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen für Kommentare in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_comment ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_comment: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Fälligkeitsdatum-Änderung
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen für Fälligkeitsdatum-Änderungen in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_due_date_change ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_due_date_change: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Zuweisung entfernt
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen wenn Zuweisung entfernt wird in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_unassignment ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_unassignment: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>

                          <div className="flex items-center justify-between">
                            <div>
                              <h3 className="font-medium text-gray-900 dark:text-white">
                                Bei Checklisten-Element
                              </h3>
                              <p className="text-sm text-gray-600 dark:text-gray-400">
                                Benachrichtigungen für Checklisten-Elemente in diesem Board
                              </p>
                            </div>
                            <label className="relative inline-flex items-center cursor-pointer">
                              <input
                                type="checkbox"
                                checked={boardSettings?.notify_on_checklist_item ?? true}
                                onChange={(e) =>
                                  updateBoardSettings(board.id, { notify_on_checklist_item: e.target.checked })
                                }
                                disabled={saving || !isEnabled}
                                className="sr-only peer"
                              />
                              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                            </label>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
