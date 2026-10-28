import type { CustomLink, FeedComment, ProfileUser } from './profile.types';

type StoredCustomLink = {
  id?: string;
  label?: string;
  url?: string;
};

export const PROFILE_STORAGE_KEY = 'chattr_profile_info';

export const readStoredProfile = () => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) {
      return {} as Partial<ProfileUser>;
    }

    const parsed = JSON.parse(raw) as Partial<ProfileUser> & {
      links?: Record<string, string>;
      customLinks?: StoredCustomLink[];
    };

    const customLinks = Array.isArray(parsed.customLinks)
      ? parsed.customLinks
          .filter((entry) => !!entry && typeof entry === 'object')
          .map((entry, index) => {
            const candidate = entry as Partial<StoredCustomLink>;
            return {
              id: typeof candidate.id === 'string' ? candidate.id : `link-${index}`,
              label: typeof candidate.label === 'string' ? candidate.label : 'Link',
              url: typeof candidate.url === 'string' ? candidate.url : '',
            };
          })
          .filter((entry) => entry.url.trim())
      : [];

    return {
      id: typeof parsed.id === 'string' ? parsed.id : undefined,
      name: typeof parsed.name === 'string' ? parsed.name : undefined,
      username: typeof parsed.username === 'string' ? parsed.username : undefined,
      profileImageUrl:
        typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : undefined,
      bio: typeof parsed.bio === 'string' ? parsed.bio : undefined,
      website:
        typeof parsed.website === 'string'
          ? parsed.website
          : typeof parsed.links?.website === 'string'
            ? parsed.links.website
            : undefined,
      linkedin:
        typeof parsed.linkedin === 'string'
          ? parsed.linkedin
          : typeof parsed.links?.linkedin === 'string'
            ? parsed.links.linkedin
            : undefined,
      instagram:
        typeof parsed.instagram === 'string'
          ? parsed.instagram
          : typeof parsed.links?.instagram === 'string'
            ? parsed.links.instagram
            : undefined,
      customLinks: customLinks.length > 0 ? customLinks : undefined,
    };
  } catch (_error) {
    return {} as Partial<ProfileUser>;
  }
};

export const getInitials = (value: string) =>
  value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'A';

export const normalizeLinkUrl = (value: string) => value.trim().replace(/\/+$/, '').toLowerCase();

const getStringValue = (record: Record<string, unknown>, keys: string[]) => {
  for (const key of keys) {
    const value = record[key];
    if (typeof value === 'string' && value.trim()) {
      return value;
    }
  }

  return undefined;
};

const normalizeCommentNode = (
  rawComment: Record<string, unknown>,
  parentId: string,
  index: number,
): FeedComment => {
  const userDetails =
    (rawComment.UserDetails as Record<string, unknown>) ||
    (rawComment.userDetails as Record<string, unknown>) ||
    (rawComment.user_details as Record<string, unknown>) ||
    {};

  const author =
    getStringValue(rawComment, ['Name', 'name']) ||
    getStringValue(userDetails, ['Name', 'name', 'UserName', 'userName']) ||
    'User';

  const text =
    getStringValue(rawComment, ['CommentText', 'comment_text', 'commentText']) || 'Comment';

  const profileImageUrl =
    getStringValue(rawComment, ['ProfileImgUrl', 'profile_image_url', 'profileImageUrl']) ||
    getStringValue(userDetails, ['ProfileImgUrl', 'profile_image_url', 'profileImageUrl']) ||
    '';

  const time =
    getStringValue(rawComment, ['CommentAge', 'comment_age']) || 'just now';

  const rawChildComments = Array.isArray(rawComment.ChildComment)
    ? rawComment.ChildComment
    : Array.isArray(rawComment.childComment)
      ? rawComment.childComment
      : [];

  const replies = rawChildComments.map((child, childIndex) =>
    normalizeCommentNode(
      child as Record<string, unknown>,
      `${parentId}-${index}`,
      childIndex,
    ),
  );

  return {
    id: String(
      rawComment.CommentID ??
        rawComment.comment_id ??
        rawComment.id ??
        `${parentId}-${index}`,
    ),
    author,
    text,
    time,
    profileImageUrl,
    replyCount:
      Number(rawComment.ChildCommentCount ?? rawComment.child_comment_count ?? replies.length) ||
      replies.length,
    replies,
  };
};

export const normalizeCommentTree = (commentList: unknown[], postId: string): FeedComment[] =>
  (Array.isArray(commentList) ? commentList : []).map((comment, index) =>
    normalizeCommentNode(comment as Record<string, unknown>, postId, index),
  );

export const normalizeCustomLinks = (links?: CustomLink[]) =>
  (links ?? []).filter((link) => link && typeof link.url === 'string' && link.url.trim());

export const dedupeCustomLinks = (
  user: Pick<ProfileUser, 'customLinks' | 'website' | 'linkedin' | 'instagram'>,
) => {
  const socialUrls = [user.website, user.linkedin, user.instagram]
    .filter((value): value is string => Boolean(value))
    .map(normalizeLinkUrl);

  return (user.customLinks ?? []).filter((link) => {
    if (!link.url) {
      return false;
    }

    return !socialUrls.includes(normalizeLinkUrl(link.url));
  });
};
