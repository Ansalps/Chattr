const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export type SubscriptionPlan = {
  id: string;
  name: string;
  price: number;
  currency: string;
  period: string;
  interval: number;
  description: string;
  isActive: boolean;
};

type Payload = Record<string, unknown>;

const unwrap = (payload: unknown): Payload => {
  const root = (payload && typeof payload === 'object' ? payload : {}) as Payload;
  return (root.data && typeof root.data === 'object' ? root.data : root) as Payload;
};

const request = async (path: string, options?: RequestInit) => {
  const response = await fetch(`${API_BASE_URL}${path}`, { credentials: 'include', ...options });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok || Number((payload as Payload).status_code ?? 200) >= 400) {
    throw new Error(String((payload as Payload).message ?? 'Subscription request failed.'));
  }
  return payload;
};

const normalizePlan = (value: unknown): SubscriptionPlan | null => {
  if (!value || typeof value !== 'object') return null;
  const plan = value as Payload;
  const id = plan.id ?? plan.ID;
  if (id === undefined) return null;
  return {
    id: String(id), name: String(plan.name ?? plan.Name ?? 'Verification'),
    price: Number(plan.price ?? plan.Price ?? 0), currency: String(plan.currency ?? plan.Currency ?? 'INR'),
    period: String(plan.period ?? plan.Period ?? 'monthly'), interval: Number(plan.interval ?? plan.Interval ?? 1),
    description: String(plan.description ?? plan.Description ?? ''), isActive: Boolean(plan.is_active ?? plan.IsActive),
  };
};

export async function getPlans(admin = false): Promise<SubscriptionPlan[]> {
  const path = admin ? '/admin/subscription-plan/get-all-subscription-plans' : '/user/subscription-plan/get-all-active-subscription-plans';
  const payload = unwrap(await request(`${path}?page=1&limit=50`));
  const list = payload.subscriptionPlans ?? payload.SubscriptionPlans ?? [];
  return (Array.isArray(list) ? list : []).map(normalizePlan).filter((plan): plan is SubscriptionPlan => Boolean(plan));
}

export async function createPlan(input: Omit<SubscriptionPlan, 'id' | 'isActive'>) {
  await request('/admin/subscription-plan', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(input) });
}

export async function setPlanActive(id: string, active: boolean) {
  await request(`/admin/subscription-plan/${encodeURIComponent(id)}/${active ? 'activate' : 'deactivate'}`, { method: 'PATCH' });
}

export async function startSubscription(planId: string, totalCount: number) {
  const data = unwrap(await request(`/user/subscribe/${encodeURIComponent(planId)}`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ total_count: totalCount }) }));
  const subscription = (data.subscription ?? data.Subscription ?? data) as Payload;
  return { keyId: String(data.key_id ?? data.KeyID ?? ''), subscriptionId: String(subscription.razorpay_subscription_id ?? subscription.RazorpaySubscriptionId ?? ''), shortUrl: String(subscription.short_url ?? subscription.ShortUrl ?? '') };
}

export async function getSubscriptionDetails(): Promise<(Payload & { keyId: string }) | null> {
  const response = await request('/user/subscription');
  const root = response as Payload;
  const payload = root.data;
  if (!payload || typeof payload !== 'object') return null;

  const data = payload as Payload;
  const subscription = (data.subscription ?? data.Subscription ?? data) as Payload;
  const id = subscription.razorpay_subscription_id ?? subscription.RazorpaySubscriptionId;
  if (!id) return null;
  const plan = (subscription.subscription_plan ?? subscription.SubscriptionPlan ?? {}) as Payload;

  return {
    ...subscription,
    planName: String(plan.name ?? plan.Name ?? 'Verification plan'),
    price: Number(plan.price ?? plan.Price ?? 0),
    currency: String(plan.currency ?? plan.Currency ?? 'INR'),
    period: String(plan.period ?? plan.Period ?? 'period'),
    interval: Number(plan.interval ?? plan.Interval ?? 1),
    keyId: String(data.key_id ?? data.KeyID ?? ''),
  };
}

export async function cancelSubscription(reason: string) {
  await request('/user/unsubscribe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ cancel_reason: reason, cancel_at_cycle_end: false }),
  });
}

