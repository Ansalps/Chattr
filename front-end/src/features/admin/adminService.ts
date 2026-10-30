const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';
const PAGE_SIZE = 100;

export type ManagedUser = {
  id: string;
  name: string;
  username: string;
  email: string;
  status: string;
  profileImageUrl?: string;
};

type Payload = Record<string, unknown>;

async function request(path: string, options?: RequestInit): Promise<Payload> {
  const response = await fetch(`${API_BASE_URL}${path}`, { credentials: 'include', ...options });
  const payload = await response.json().catch(() => ({})) as Payload;
  if (!response.ok || Number(payload.status_code ?? 200) >= 400) {
    throw new Error(String(payload.message ?? 'Admin request failed.'));
  }
  return payload;
}

function normalizeUser(value: unknown): ManagedUser | null {
  if (!value || typeof value !== 'object') return null;
  const user = value as Payload;
  const id = user.id ?? user.ID;
  if (id === undefined || id === null) return null;
  return {
    id: String(id),
    name: String(user.name ?? user.Name ?? 'Chattr user'),
    username: String(user.user_name ?? user.UserName ?? user.username ?? ''),
    email: String(user.email ?? user.Email ?? ''),
    status: String(user.status ?? user.Status ?? 'unknown'),
    profileImageUrl: String(user.profile_img_url ?? user.ProfileImgUrl ?? ''),
  };
}

export async function getAllUsers(): Promise<ManagedUser[]> {
  const users: ManagedUser[] = [];
  for (let page = 1; ; page += 1) {
    const response = await request(`/admin/get-all-users?page=${page}&limit=${PAGE_SIZE}`);
    const data = (response.data ?? response.Data ?? response) as Payload;
    const pageUsers = data.users ?? data.Users;
    if (!Array.isArray(pageUsers)) break;
    users.push(...pageUsers.map(normalizeUser).filter((user): user is ManagedUser => Boolean(user)));
    if (pageUsers.length < PAGE_SIZE) break;
  }
  return users;
}

export async function setUserBlocked(userId: string, blocked: boolean): Promise<void> {
  await request(`/admin/${blocked ? 'block' : 'unblock'}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ user_id: Number(userId) }),
  });
}
