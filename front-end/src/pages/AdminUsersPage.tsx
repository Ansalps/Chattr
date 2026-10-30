import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getAllUsers, setUserBlocked, type ManagedUser } from '../features/admin/adminService';

const initials = (name: string) => name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join('').toUpperCase() || 'U';

export default function AdminUsersPage() {
  const navigate = useNavigate();
  const [users, setUsers] = useState<ManagedUser[]>([]);
  const [search, setSearch] = useState('');
  const [busyUserId, setBusyUserId] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');

  const loadUsers = useCallback(async () => {
    setError('');
    setIsLoading(true);
    try {
      setUsers(await getAllUsers());
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Unable to load users.');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => { void loadUsers(); }, [loadUsers]);

  const visibleUsers = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) return users;
    return users.filter((user) => `${user.name} ${user.username} ${user.email} ${user.status}`.toLowerCase().includes(query));
  }, [search, users]);

  const toggleBlock = async (user: ManagedUser) => {
    const blocked = user.status.toLowerCase() === 'blocked';
    setError(''); setBusyUserId(user.id);
    try {
      await setUserBlocked(user.id, !blocked);
      setUsers((current) => current.map((item) => item.id === user.id ? { ...item, status: blocked ? 'active' : 'blocked' } : item));
    } catch (actionError) {
      setError(actionError instanceof Error ? actionError.message : `Unable to ${blocked ? 'unblock' : 'block'} user.`);
    } finally {
      setBusyUserId('');
    }
  };

  return (
    <main className="min-h-screen bg-[#f3f7fb] p-4 text-slate-800 sm:p-8">
      <div className="mx-auto max-w-6xl">
        <header className="mb-6 flex flex-wrap items-center justify-between gap-4 rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
          <div>
            <button onClick={() => navigate('/admin/dashboard')} className="text-sm font-semibold text-sky-700">← Dashboard</button>
            <h1 className="mt-2 text-3xl font-black tracking-tight text-slate-900">Manage users</h1>
            <p className="mt-1 text-sm text-slate-500">{users.length} users</p>
          </div>
          <div className="flex gap-2">
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search users" className="w-56 rounded-xl border border-slate-200 px-3 py-2 text-sm" />
            <button onClick={() => void loadUsers()} disabled={isLoading} className="rounded-xl bg-sky-50 px-4 py-2 text-sm font-semibold text-sky-700 disabled:opacity-50">Refresh</button>
          </div>
        </header>

        {error && <p className="mb-4 rounded-xl bg-rose-50 p-3 text-sm text-rose-700">{error}</p>}
        <section className="overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-sm">
          {isLoading ? <p className="p-6 text-sm text-slate-500">Loading users…</p> : visibleUsers.length === 0 ? <p className="p-6 text-sm text-slate-500">{users.length ? 'No users match your search.' : 'No users found.'}</p> : (
            <div className="divide-y divide-slate-100">
              {visibleUsers.map((user) => {
                const blocked = user.status.toLowerCase() === 'blocked';
                return <article key={user.id} className="flex flex-wrap items-center justify-between gap-4 p-4 sm:p-5">
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex h-12 w-12 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-400 to-indigo-500 text-sm font-bold text-white">
                      {user.profileImageUrl ? <img src={user.profileImageUrl} alt={user.name} className="h-full w-full object-cover" /> : initials(user.name)}
                    </div>
                    <div className="min-w-0">
                      <p className="truncate font-bold text-slate-900">{user.name}</p>
                      <p className="truncate text-sm text-slate-500">{user.username ? `@${user.username.replace(/^@/, '')} · ` : ''}{user.email}</p>
                      <p className="mt-1 text-xs capitalize text-slate-400">{user.status}</p>
                    </div>
                  </div>
                  <button disabled={busyUserId === user.id} onClick={() => void toggleBlock(user)} className={`rounded-xl px-4 py-2 text-sm font-bold disabled:opacity-50 ${blocked ? 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100' : 'bg-rose-50 text-rose-700 hover:bg-rose-100'}`}>
                    {busyUserId === user.id ? 'Updating…' : blocked ? 'Unblock' : 'Block'}
                  </button>
                </article>;
              })}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
