import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { cancelSubscription, getPlans, getSubscriptionDetails, startSubscription, type SubscriptionPlan } from '../features/subscription/subscriptionService';

declare global { interface Window { Razorpay?: new (options: Record<string, unknown>) => { open: () => void }; } }

const loadRazorpay = () => new Promise<boolean>((resolve) => {
  if (window.Razorpay) return resolve(true);
  const script = document.createElement('script');
  script.src = 'https://checkout.razorpay.com/v1/checkout.js';
  script.onload = () => resolve(Boolean(window.Razorpay));
  script.onerror = () => resolve(false);
  document.body.appendChild(script);
});

export default function SubscriptionPage() {
  const navigate = useNavigate();
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [subscription, setSubscription] = useState<Record<string, unknown> | null>(null);
  const [loadingPlan, setLoadingPlan] = useState('');
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const load = async () => {
    try {
      const [nextPlans, details] = await Promise.all([getPlans(), getSubscriptionDetails()]);
      setPlans(nextPlans);
      setSubscription(details);
    } catch (loadError) { setError(loadError instanceof Error ? loadError.message : 'Unable to load verification plans.'); }
  };
  useEffect(() => { void load(); }, []);

  const openPayment = async (keyId: string, subscriptionId: string, description: string) => {
    if (!keyId || !subscriptionId) throw new Error('Subscription checkout details are missing.');
    if (!window.Razorpay) throw new Error('Razorpay Checkout could not be loaded.');

    const checkout = new window.Razorpay({
      key: keyId,
      subscription_id: subscriptionId,
      name: 'Chattr',
      description,
      handler: async (response: Record<string, unknown>) => {
        const paidSubscriptionId = String(response.razorpay_subscription_id ?? subscriptionId);
        setNotice('Payment completed. Waiting for Razorpay’s subscription webhook to confirm activation…');
        for (let attempt = 0; attempt < 6; attempt += 1) {
          await new Promise((resolve) => window.setTimeout(resolve, 2000));
          try {
            const latest = await getSubscriptionDetails();
            setSubscription(latest);
            const status = String(latest?.status ?? latest?.Status ?? '').toLowerCase();
            if (status === 'active') {
              setNotice(`Subscription ${paidSubscriptionId} is active.`);
              return;
            }
          } catch {
            // Webhook delivery may still be processing; keep checking briefly.
          }
        }
      },
    });
    checkout.open();
  };

  const subscribe = async (plan: SubscriptionPlan) => {
    setError(''); setNotice(''); setLoadingPlan(plan.id);
    try {
      // Load checkout first so we don't create an unpaid subscription if the
      // Razorpay checkout script cannot be reached.
      const checkoutLoaded = await loadRazorpay();
      if (!checkoutLoaded || !window.Razorpay) throw new Error('Razorpay Checkout could not be loaded.');

      const order = await startSubscription(plan.id, 12);
      await openPayment(order.keyId, order.subscriptionId, `${plan.name} verification`);
    } catch (subscribeError) { setError(subscribeError instanceof Error ? subscribeError.message : 'Unable to start checkout.'); }
    finally { setLoadingPlan(''); }
  };

  const completePayment = async () => {
    if (!subscription) return;
    setError(''); setNotice(''); setLoadingPlan('existing');
    try {
      if (!window.Razorpay && !(await loadRazorpay())) throw new Error('Razorpay Checkout could not be loaded.');
      await openPayment(
        String(subscription.keyId ?? ''),
        String(subscription.razorpay_subscription_id ?? subscription.RazorpaySubscriptionId ?? ''),
        `${String(subscription.planName ?? subscription.Name ?? 'Verification')} verification`,
      );
    } catch (paymentError) {
      setError(paymentError instanceof Error ? paymentError.message : 'Unable to open payment checkout.');
    } finally { setLoadingPlan(''); }
  };

  const cancelCurrentSubscription = async () => {
    if (!subscription) return;
    if (!window.confirm('Cancel this subscription immediately?')) return;

    setError(''); setNotice(''); setLoadingPlan('cancel');
    try {
      await cancelSubscription('Cancelled by user');
      setNotice('Subscription cancellation requested.');
      await load();
    } catch (cancelError) {
      setError(cancelError instanceof Error ? cancelError.message : 'Unable to cancel subscription.');
    } finally { setLoadingPlan(''); }
  };

  return <div className="min-h-screen bg-[#f3f7fb] p-4 text-slate-800 sm:p-8"><div className="mx-auto max-w-5xl">
    <button onClick={() => navigate('/profile')} className="mb-6 rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm font-semibold">← Profile</button>
    <section className="rounded-[30px] bg-gradient-to-r from-sky-500 to-blue-700 p-7 text-white shadow-xl"><p className="text-xs font-bold uppercase tracking-[0.2em] text-sky-100">Chattr verification</p><h1 className="mt-2 text-3xl font-black">Get your blue tick</h1><p className="mt-2 max-w-2xl text-sky-100">Choose a subscription plan and pay securely with Razorpay. Your verified status is managed by Chattr’s subscription service.</p></section>
    {subscription && <section className="mt-5 rounded-2xl border border-sky-100 bg-white p-5 shadow-sm">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-xs font-bold uppercase tracking-wider text-sky-700">Current subscription</p>
          <h2 className="mt-1 text-xl font-black">{String(subscription.planName ?? subscription.Name ?? 'Verification plan')}</h2>
          <p className="mt-1 text-sm text-slate-600">Status: <strong className="capitalize">{String(subscription.status ?? subscription.Status ?? 'Unknown')}</strong></p>
          <p className="mt-1 text-sm text-slate-600">{String(subscription.currency ?? subscription.Currency ?? '')} {(Number(subscription.price ?? subscription.Price ?? 0) / 100).toFixed(2)} · every {String(subscription.interval ?? subscription.Interval ?? 1)} {String(subscription.period ?? subscription.Period ?? 'period')}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {String(subscription.status ?? subscription.Status ?? '').toLowerCase() === 'created' && <button disabled={Boolean(loadingPlan)} onClick={() => void completePayment()} className="rounded-xl bg-sky-600 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-60">{loadingPlan === 'existing' ? 'Opening checkout…' : 'Complete payment'}</button>}
          {!subscription.cancel_reason && !subscription.CancelReason && ['active', 'halted', 'created'].includes(String(subscription.status ?? subscription.Status ?? '').toLowerCase()) && <button disabled={Boolean(loadingPlan)} onClick={() => void cancelCurrentSubscription()} className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-2.5 text-sm font-bold text-rose-700 disabled:opacity-60">{loadingPlan === 'cancel' ? 'Cancelling…' : 'Cancel subscription'}</button>}
        </div>
      </div>
      <div className="mt-4 grid gap-3 text-sm text-slate-600 sm:grid-cols-3">
        <p>Started: {subscription.start_at || subscription.StartAt ? new Date(String(subscription.start_at ?? subscription.StartAt)).toLocaleDateString() : 'Pending payment'}</p>
        <p>Next charge: {subscription.next_charge_at || subscription.NextChargeAt ? new Date(String(subscription.next_charge_at ?? subscription.NextChargeAt)).toLocaleDateString() : 'Not scheduled'}</p>
        <p>Payments: {String(subscription.paid_count ?? subscription.PaidCount ?? 0)} of {String(subscription.total_count ?? subscription.TotalCount ?? 0)}</p>
      </div>
      {(subscription.cancel_reason || subscription.CancelReason) && <p className="mt-3 text-sm font-medium text-amber-700">Cancellation requested: {String(subscription.cancel_reason ?? subscription.CancelReason)}</p>}
    </section>}
    {error && <p className="mt-5 rounded-xl bg-rose-50 p-3 text-sm text-rose-700">{error}</p>}{notice && <p className="mt-5 rounded-xl bg-emerald-50 p-3 text-sm text-emerald-700">{notice}</p>}
    <div className="mt-7 grid gap-5 md:grid-cols-2 lg:grid-cols-3">{plans.map((plan) => <article key={plan.id} className="rounded-[28px] border border-sky-100 bg-white p-6 shadow-sm"><h2 className="text-xl font-black">{plan.name}</h2><p className="mt-2 min-h-12 text-sm text-slate-500">{plan.description}</p><p className="mt-6 text-3xl font-black">{plan.currency} {(plan.price / 100).toFixed(2)}<span className="text-sm font-medium text-slate-500"> / {plan.interval} {plan.period}</span></p><button disabled={Boolean(loadingPlan)} onClick={() => void subscribe(plan)} className="mt-6 w-full rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3 text-sm font-bold text-white disabled:opacity-60">{loadingPlan === plan.id ? 'Opening Razorpay…' : 'Get verified'}</button></article>)}</div>
  </div></div>;
}
