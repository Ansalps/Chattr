import { useEffect, useState } from 'react';

type AdminUser = {
  id?: number;
  email?: string;
};

export default function AdminDashboard() {
  const [adminUser, setAdminUser] = useState<AdminUser>({});

  useEffect(() => {
    const storedUser = localStorage.getItem('chattr_admin_user');
    if (storedUser) {
      try {
        setAdminUser(JSON.parse(storedUser));
      } catch {
        setAdminUser({});
      }
    }
  }, []);

  const stats = [
    { label: 'Total users', value: '24.8K', tone: 'bg-sky-100 text-sky-700' },
    { label: 'Active today', value: '8.4K', tone: 'bg-emerald-100 text-emerald-700' },
    { label: 'Reports', value: '126', tone: 'bg-amber-100 text-amber-700' },
    { label: 'Subscriptions', value: '3.1K', tone: 'bg-violet-100 text-violet-700' },
  ];

  const activity = [
    { title: 'New creator onboarding', detail: '12 accounts approved this morning', accent: 'bg-sky-500' },
    { title: 'Moderation queue', detail: '24 posts flagged for review', accent: 'bg-amber-500' },
    { title: 'Revenue update', detail: 'Premium plan renewals are up 12%', accent: 'bg-emerald-500' },
  ];

  return (
    <div className="min-h-screen bg-[#f3f7fb] p-4 text-slate-800 sm:p-6 lg:p-10">
      <div className="mx-auto max-w-7xl">
        <header className="mb-8 flex flex-col gap-4 rounded-[28px] border border-sky-100 bg-white p-5 shadow-[0_20px_50px_rgba(14,116,144,0.08)] sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.22em] text-sky-600">Chattr admin</p>
            <h1 className="mt-2 text-3xl font-black tracking-tight text-slate-900">Dashboard</h1>
          </div>

          <div className="flex items-center gap-4">
            <div className="rounded-2xl border border-sky-100 bg-sky-50 px-3 py-2 text-right">
              <p className="text-[10px] uppercase tracking-[0.18em] text-slate-500">Signed in</p>
              <p className="text-sm font-semibold text-slate-800">{adminUser.email || 'admin@chattr.com'}</p>
            </div>
          </div>
        </header>

        <section className="mb-8 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {stats.map((stat) => (
            <div key={stat.label} className="rounded-[24px] border border-sky-100 bg-white p-5 shadow-sm">
              <div className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${stat.tone}`}>
                {stat.label}
              </div>
              <p className="mt-4 text-3xl font-black text-slate-900">{stat.value}</p>
            </div>
          ))}
        </section>

        <section className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
          <div className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm">
            <div className="mb-5 flex items-center justify-between">
              <h2 className="text-xl font-bold text-slate-900">Recent activity</h2>
              <button className="rounded-full border border-sky-200 bg-sky-50 px-3 py-1.5 text-sm font-medium text-sky-700">
                View all
              </button>
            </div>

            <div className="space-y-4">
              {activity.map((item) => (
                <div key={item.title} className="flex items-start gap-4 rounded-2xl border border-slate-100 bg-slate-50 p-4">
                  <div className={`mt-1 h-3 w-3 rounded-full ${item.accent}`} />
                  <div>
                    <p className="font-semibold text-slate-900">{item.title}</p>
                    <p className="mt-1 text-sm text-slate-500">{item.detail}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-6">
            <div className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm">
              <h3 className="text-lg font-bold text-slate-900">Quick actions</h3>
              <div className="mt-4 space-y-3">
                {['Manage users', 'Review reports', 'Subscription plans', 'Moderation queue'].map((action) => (
                  <button
                    key={action}
                    className="flex w-full items-center justify-between rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-left text-sm font-medium text-slate-700 transition hover:border-sky-200 hover:bg-sky-100"
                  >
                    <span>{action}</span>
                    <span aria-hidden="true">→</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-sky-100 bg-gradient-to-br from-sky-500 to-blue-600 p-6 text-white shadow-[0_20px_40px_rgba(59,130,246,0.25)]">
              <p className="text-sm uppercase tracking-[0.18em] text-sky-100">System health</p>
              <p className="mt-4 text-4xl font-black">99.9%</p>
              <p className="mt-2 text-sm text-sky-100">API gateway and admin services are stable.</p>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
