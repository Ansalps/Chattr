const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export type NotificationItem = {
  id: string;
  actorId: string;
  targetId: string;
  type: string;
  message: string;
  createdAt: string;
};

type NotificationPayload = Record<string, unknown>;

const asNotification = (value: unknown): NotificationItem | null => {
  if (!value || typeof value !== 'object') return null;
  const source = value as NotificationPayload;
  const message = String(source.message ?? 'You have a new notification.');
  const type = String(source.type ?? 'NOTIFICATION');

  const sourceId = String(source.id ?? '');
  return {
    id: sourceId && sourceId !== '0' ? sourceId : `live-${Date.now()}-${Math.random()}`,
    actorId: String(source.actor_id ?? source.actorId ?? ''),
    targetId: String(source.target_id ?? source.targetId ?? ''),
    type,
    message,
    createdAt: String(source.created_at ?? source.createdAt ?? new Date().toISOString()),
  };
};

export async function fetchNotifications(): Promise<NotificationItem[]> {
  const response = await fetch(`${API_BASE_URL}/user/notifications?page=1&limit=50`, {
    credentials: 'include',
  });
  const payload = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error((payload as NotificationPayload).message ? String((payload as NotificationPayload).message) : 'Unable to load notifications.');
  }

  const source = payload as NotificationPayload;
  const list = Array.isArray(source.notifications)
    ? source.notifications
    : Array.isArray((source.data as NotificationPayload | undefined)?.notifications)
      ? ((source.data as NotificationPayload).notifications as unknown[])
      : [];

  return list.map(asNotification).filter((item): item is NotificationItem => Boolean(item));
}

export function connectNotificationSocket(
  onNotification: (notification: NotificationItem) => void,
  onClose?: () => void,
): WebSocket {
  const socketUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:3000/user/notification/ws`;
  const socket = new WebSocket(socketUrl);

  socket.onmessage = (event) => {
    try {
      const parsed = JSON.parse(String(event.data));
      const notification = asNotification(parsed);
      if (notification) onNotification(notification);
    } catch {
      // Older services send plain text. Persisted notifications remain available on refresh.
    }
  };

  socket.onclose = () => onClose?.();

  return socket;
}
