import type { FeedPost } from '../types';
import { formatFriendlyTime, normalizeUsername } from './formatting';

export const PROFILE_STORAGE_KEY = 'chattr_profile_info';

export const readStoredProfile = () => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) return { id: '', name: '', username: '', profileImageUrl: '', blueTick: false };

    const parsed = JSON.parse(raw) as Partial<{
      id?: string | number;
      userId?: string | number;
      name?: string;
      username?: string;
      profileImageUrl?: string;
      blueTick?: boolean;
    }>;
    return {
      id:
        parsed.id !== undefined && parsed.id !== null
          ? String(parsed.id)
          : parsed.userId !== undefined && parsed.userId !== null
            ? String(parsed.userId)
            : '',
      name: typeof parsed.name === 'string' ? parsed.name : '',
      username: typeof parsed.username === 'string' ? parsed.username : '',
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : '',
      blueTick: parsed.blueTick === true,
    };
  } catch (error) {
    console.error(error)
    return { id: '', name: '', username: '', profileImageUrl: '', blueTick: false };
  }
};

export const normalizeStringArray = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value
      .map((entry) => {
        if (typeof entry === 'string') return entry;
        if (entry && typeof entry === 'object') {
          const item = entry as Record<string, unknown>;
          return typeof item.url === 'string'
            ? item.url
            : typeof item.media_url === 'string'
              ? item.media_url
              : '';
        }
        return '';
      })
      .filter((entry): entry is string => Boolean(entry));
  }

  if (typeof value === 'string' && value.trim()) {
    return [value];
  }

  return [];
};

export const getPostIdFromResponse = (value: unknown): string | null => {
  if (!value || typeof value !== 'object') return null;

  const source = value as Record<string, unknown>;
  const candidates = [source.res, source.data, source];

  for (const candidate of candidates) {
    if (!candidate || typeof candidate !== 'object') continue;

    const item = candidate as Record<string, unknown>;
    const rawId =
      item.post_id ??
      item.PostId ??
      item.postId ??
      item.PostID ??
      item.id ??
      item.ID;

    if (rawId === undefined || rawId === null || rawId === '') continue;

    const asNumber = Number(rawId);
    if (Number.isFinite(asNumber)) {
      return String(asNumber);
    }
  }

  return null;
};

export const normalizeFeedPosts = (value: unknown): FeedPost[] => {
  const list = Array.isArray(value)
    ? value
    : value && typeof value === 'object'
      ? (value as Record<string, unknown>).PostUserData && Array.isArray((value as Record<string, unknown>).PostUserData)
        ? ((value as Record<string, unknown>).PostUserData as unknown[])
        : (value as Record<string, unknown>).data && Array.isArray((value as Record<string, unknown>).data)
          ? ((value as Record<string, unknown>).data as unknown[])
          : []
      : [];

  return list
    .map((item) => {
      const source = item as Record<string, unknown>;
      const userMeta =
        (source.user_meta_data as Record<string, unknown>) ||
        (source.userMetaData as Record<string, unknown>) ||
        (source.UserMetaData as Record<string, unknown>) ||
        (source.user as Record<string, unknown>) ||
        (source.UserData as Record<string, unknown>) ||
        (source.userData as Record<string, unknown>) ||
        {};

      const nestedUserData = (source.UserData as Record<string, unknown>) || (source.userData as Record<string, unknown>) || {};
      const blueTickValue =
        source.blue_tick ??
        source.blueTick ??
        source.BlueTick ??
        userMeta.blue_tick ??
        userMeta.blueTick ??
        userMeta.BlueTick ??
        nestedUserData.blue_tick ??
        nestedUserData.blueTick ??
        nestedUserData.BlueTick;
      const blueTick = blueTickValue === true || blueTickValue === 'true';

      const rawName =
        (typeof source.name === 'string' && source.name) ||
        (typeof source.Name === 'string' && source.Name) ||
        (typeof source.full_name === 'string' && source.full_name) ||
        (typeof source.FullName === 'string' && source.FullName) ||
        (typeof userMeta.name === 'string' && userMeta.name) ||
        (typeof userMeta.Name === 'string' && userMeta.Name) ||
        (typeof userMeta.full_name === 'string' && userMeta.full_name) ||
        (typeof userMeta.FullName === 'string' && userMeta.FullName) ||
        '';

      const rawUsername =
        (typeof source.UserName === 'string' && source.UserName) ||
        (typeof source.userName === 'string' && source.userName) ||
        (typeof source.username === 'string' && source.username) ||
        (typeof source.handle === 'string' && source.handle) ||
        (typeof source.Handle === 'string' && source.Handle) ||
        (typeof userMeta.UserName === 'string' && userMeta.UserName) ||
        (typeof userMeta.userName === 'string' && userMeta.userName) ||
        (typeof userMeta.username === 'string' && userMeta.username) ||
        (typeof userMeta.handle === 'string' && userMeta.handle) ||
        (typeof userMeta.Handle === 'string' && userMeta.Handle) ||
        '';

      const author = rawName || rawUsername.replace(/^@/, '') || 'Chattr User';

      const profileImageUrl =
        (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
        (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
        (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
        (typeof userMeta.profile_image_url === 'string' && userMeta.profile_image_url) ||
        (typeof userMeta.ProfileImgUrl === 'string' && userMeta.ProfileImgUrl) ||
        '';

      const handle =
        (typeof source.handle === 'string' && normalizeUsername(source.handle)) ||
        (typeof source.Handle === 'string' && normalizeUsername(source.Handle)) ||
        (typeof userMeta.username === 'string' && normalizeUsername(userMeta.username)) ||
        (typeof userMeta.UserName === 'string' && normalizeUsername(userMeta.UserName)) ||
        (typeof userMeta.userName === 'string' && normalizeUsername(userMeta.userName)) ||
        (rawUsername ? normalizeUsername(rawUsername) : `@${author.toLowerCase().replace(/\s+/g, '') || 'user'}`);

      const content =
        (typeof source.content === 'string' && source.content) ||
        (typeof source.Content === 'string' && source.Content) ||
        (typeof source.caption === 'string' && source.caption) ||
        (typeof source.Caption === 'string' && source.Caption) ||
        (typeof source.description === 'string' && source.description) ||
        '';

      const imageUrls = normalizeStringArray(
        source.media_urls ??
          source.mediaUrls ??
          source.MediaUrls ??
          source.Media ??
          source.media ??
          source.images ??
          source.image_urls ??
          source.imageUrls ??
          source.Images,
      );

      const isLiked = Boolean(
        source.isLiked ??
          source.IsLiked ??
          source.is_liked ??
          source.Isliked ??
          false,
      );

      const likes = Number(
        source.likes ??
          source.Likes ??
          source.LikeCount ??
          source.likeCount ??
          source.like_count ??
          source.likes_count ??
          source.LikesCount ??
          source.likesCount ??
          0,
      ) || 0;
      const comments = Number(
        source.comments ??
          source.Comments ??
          source.CommentsCount ??
          source.commentsCount ??
          source.comment_count ??
          source.comments_count ??
          0,
      ) || 0;
      const shares = Number(
        source.shares ??
          source.share_count ??
          source.shareCount ??
          source.Shares ??
          0,
      ) || 0;

      const userId = String(
        source.user_id ??
          source.UserID ??
          source.userId ??
          source.userID ??
          (nestedUserData.UserID as string | number | undefined) ??
          (nestedUserData.userID as string | number | undefined) ??
          (userMeta.user_id as string | number | undefined) ??
          (userMeta.UserID as string | number | undefined) ??
          (userMeta.userID as string | number | undefined) ??
          'me',
      );

      return {
        id: String(
          source.id ??
            source.post_id ??
            source.PostID ??
            source.PostId ??
            `${author}-${Math.random()}`,
        ),
        userId,
        author,
        blueTick,
        handle,
        time: formatFriendlyTime(
          source.post_age ??
            source.PostAge ??
            source.time ??
            source.created_at ??
            source.createdAt ??
            source.CreatedAt ??
            new Date().toISOString(),
        ),
        content,
        imageUrls,
        profileImageUrl,
        isLiked,
        likes,
        comments,
        shares,
        accent: typeof source.accent === 'string' && source.accent ? source.accent : 'from-sky-400 to-blue-600',
        tag: typeof source.tag === 'string' && source.tag ? source.tag : 'Update',
      };
    })
    .filter((post) => !!post.content || post.imageUrls.length > 0);
};
