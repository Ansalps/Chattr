export const formatCount = (value: number) =>
  new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 }).format(value);

export const formatFriendlyTime = (value: unknown): string => {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    if (!trimmed) return 'just now';

    const lower = trimmed.toLowerCase();
    if (lower.includes('ago') || lower.includes('just now') || lower.includes('today') || lower.includes('yesterday')) {
      return trimmed;
    }

    const parsedDate = new Date(trimmed);
    if (!Number.isNaN(parsedDate.getTime())) {
      const diffMs = parsedDate.getTime() - Date.now();
      const diffMinutes = Math.round(diffMs / 60000);
      const diffHours = Math.round(diffMs / 3600000);
      const diffDays = Math.round(diffMs / 86400000);

      if (Math.abs(diffMinutes) < 60) {
        return `${Math.abs(diffMinutes) || 1}m ago`;
      }
      if (Math.abs(diffHours) < 24) {
        return `${Math.abs(diffHours) || 1}h ago`;
      }
      if (Math.abs(diffDays) < 7) {
        return `${Math.abs(diffDays) || 1}d ago`;
      }

      return parsedDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }

    return trimmed;
  }

  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return formatFriendlyTime(value.toISOString());
  }

  return 'just now';
};

export const parseMessageDate = (value: unknown): Date | null => {
  if (value === null || value === undefined || value === '') {
    return null;
  }

  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value;
  }

  if (typeof value === 'number') {
    const millis = value > 1_000_000_000_000 ? value : value * 1000;
    const parsed = new Date(millis);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  if (typeof value === 'string') {
    const trimmed = value.trim();
    if (!trimmed) return null;

    const numericValue = Number(trimmed);
    if (!Number.isNaN(numericValue)) {
      const timestamp = Math.abs(numericValue) > 1_000_000_000_000 ? numericValue : numericValue * 1000;
      const parsed = new Date(timestamp);
      if (!Number.isNaN(parsed.getTime())) return parsed;
    }

    const parsed = new Date(trimmed);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  return null;
};

export const formatMessageTimestamp = (value: unknown): string => {
  const parsed = parseMessageDate(value);
  if (!parsed) return 'just now';

  return new Intl.DateTimeFormat('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  }).format(parsed);
};

export const getInitials = (value: string) =>
  value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'A';

export const getUsernameInitial = (value: string) => {
  const trimmed = value.replace(/^@/, '').trim();
  return trimmed ? trimmed.charAt(0).toUpperCase() : 'A';
};

export const normalizeUsername = (value: string) => {
  const trimmed = value.trim();
  if (!trimmed) return '@user';
  return trimmed.startsWith('@') ? trimmed : `@${trimmed}`;
};
