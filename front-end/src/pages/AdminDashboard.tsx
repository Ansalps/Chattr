import { useEffect, useState, type FormEvent } from 'react';
import { createPlan, getPlans, setPlanActive, type SubscriptionPlan } from '../features/subscription/subscriptionService';

type AdminUser = {
  id?: number;
  email?: string;
};

export default function AdminDashboard() {
  const [adminUser, setAdminUser] = useState<AdminUser>({});
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [planError, setPlanError] = useState('');
  const [isSavingPlan, setIsSavingPlan] = useState(false);
  const [planForm, setPlanForm] = useState({ name: '', price: '', currency: 'INR', period: 'monthly', interval: '1', description: '' });

  const loadPlans = async () => {
    try { setPlans(await getPlans(true)); } catch (error) { setPlanError(error instanceof Error ? error.message : 'Unable to load subscription plans.'); }
  };

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

  useEffect(() => { void loadPlans(); }, []);

  const handleCreatePlan = async (event: FormEvent) => {
    event.preventDefault(); setPlanError(''); setIsSavingPlan(true);
    try {
      await createPlan({ name: planForm.name, price: Math.round(Number(planForm.price)), currency: planForm.currency, period: planForm.period, interval: Number(planForm.interval), description: planForm.description });
      setPlanForm({ name: '', price: '', currency: 'INR', period: 'monthly', interval: '1', description: '' });
      await loadPlans();
    } catch (error) { setPlanError(error instanceof Error ? error.message : 'Unable to create plan.'); }
    finally { setIsSavingPlan(false); }
  };

  const togglePlan = async (plan: SubscriptionPlan) => {
    try { await setPlanActive(plan.id, !plan.isActive); await loadPlans(); } catch (error) { setPlanError(error instanceof Error ? error.message : 'Unable to update plan.'); }
  };

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

        <section className="mb-8 grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
          <div className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm">
            <div className="mb-5 flex items-center justify-between"><div><h2 className="text-xl font-bold text-slate-900">Verification subscription plans</h2><p className="mt-1 text-sm text-slate-500">Plans shown to users for blue tick verification.</p></div><button onClick={() => void loadPlans()} className="text-sm font-semibold text-sky-700">Refresh</button></div>
            {planError && <p className="mb-3 rounded-xl bg-rose-50 p-3 text-sm text-rose-700">{planError}</p>}
            <div className="space-y-3">{plans.length === 0 ? <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">No plans created yet.</p> : plans.map((plan) => <div key={plan.id} className="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-slate-100 p-4"><div><p className="font-bold">{plan.name}</p><p className="text-sm text-slate-500">{plan.currency} {(plan.price / 100).toFixed(2)} · every {plan.interval} {plan.period}</p></div><button onClick={() => void togglePlan(plan)} className={`rounded-xl px-3 py-2 text-sm font-semibold ${plan.isActive ? 'bg-emerald-100 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}>{plan.isActive ? 'Active — deactivate' : 'Inactive — activate'}</button></div>)}</div>
          </div>
          <form onSubmit={handleCreatePlan} className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm"><h2 className="text-xl font-bold">Create verification plan</h2><div className="mt-4 space-y-3">{([['name', 'Plan name'], ['price', 'Price in INR'], ['description', 'Description']] as const).map(([field, label]) => <input key={field} required value={planForm[field]} onChange={(event) => setPlanForm((current) => ({ ...current, [field]: event.target.value }))} placeholder={label} type={field === 'price' ? 'number' : 'text'} className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm" />)}<div className="grid grid-cols-2 gap-3"><select value={planForm.period} onChange={(event) => setPlanForm((current) => ({ ...current, period: event.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2 text-sm"><option value="monthly">Monthly</option><option value="yearly">Yearly</option></select><input required min="1" type="number" value={planForm.interval} onChange={(event) => setPlanForm((current) => ({ ...current, interval: event.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2 text-sm" /></div><button disabled={isSavingPlan} className="w-full rounded-xl bg-sky-600 py-2.5 text-sm font-bold text-white disabled:opacity-60">{isSavingPlan ? 'Creating…' : 'Create plan'}</button></div></form>
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
