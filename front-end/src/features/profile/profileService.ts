import { normalizeUsername } from '../../shared/utils/formatting';
import { PROFILE_STORAGE_KEY, readStoredProfile } from '../../shared/utils/normalizers';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export type CurrentUserProfile = {
  id: string;
  name: string;
  username: string;
  profileImageUrl: string;
  followerCount: number;
  followingCount: number;
  postCount: number;
};

export type ProfilePost = {
  id: string;
  content: string;
  imageUrls: string[];
  likes: number;
  comments: number;
  time: string;
  isLiked?: boolean;
};

export type CustomLink = {
  id: string;
  label: string;
  url: string;
};

export type ProfileUser = {
  id: string;
  name: string;
  username: string;
  profileImageUrl?: string;
  bio?: string;
  website?: string;
  linkedin?: string;
  instagram?: string;
  customLinks?: CustomLink[];
  followerCount?: number;
  followingCount?: number;
};

export type ProfilePageData = {
  posts: ProfilePost[];
  user: ProfileUser;
  followerCount: number;
  followingCount: number;
  isFollowing: boolean;
};

export type SocialConnectionUser = {
  id: string;
  name: string;
  username: string;
  profileImageUrl?: string;
};

const getNestedUserList = (value: unknown): unknown[] => {
  if (Array.isArray(value)) {
    return value as unknown[];
  }

  if (!value || typeof value !== 'object') {
    return [];
  }

  const source = value as Record<string, unknown>;
  const keys = [
    'UserMetaData',
    'userMetaData',
    'UserMeataData',
    'userMeataData',
    'users',
    'Users',
    'followers',
    'following',
    'data',
    'response',
    'reponse',
  ];

  for (const key of keys) {
    const candidate = source[key];
    if (Array.isArray(candidate)) {
      return candidate;
    }

    if (candidate && typeof candidate === 'object') {
      const nested = getNestedUserList(candidate);
      if (nested.length > 0) {
        return nested;
      }
    }
  }

  return [];
};

const normalizeSocialConnectionUser = (value: unknown): SocialConnectionUser | null => {
  if (!value || typeof value !== 'object') {
    return null;
  }

  const source = value as Record<string, unknown>;
  const id = String(
    source.UserID ??
      source.userID ??
      source.user_id ??
      source.UserId ??
      source.id ??
      source.User ??
      '',
  );
  const username = normalizeUsername(
    String(
      source.UserName ??
        source.userName ??
        source.username ??
        source.user_name ??
        source.Username ??
        source.User ??
        '@user',
    ),
  );
  const name = String(
    source.Name ??
      source.name ??
      source.FullName ??
      source.full_name ??
      (username && username !== '@user' ? username.replace(/^@/, '') : 'User'),
  );
  const profileImageUrl =
    (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
    (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
    (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
    '';

  if (!id && !name && !username) {
    return null;
  }

  return {
    id: id || name || username,
    name: name || username.replace(/^@/, '') || 'User',
    username: username || '@user',
    profileImageUrl: profileImageUrl || undefined,
  };
};

const normalizeProfilePosts = (value: unknown): ProfilePost[] => {
  const items = Array.isArray(value)
    ? value
    : value && typeof value === 'object' && Array.isArray((value as Record<string, unknown>).data)
      ? ((value as Record<string, unknown>).data as unknown[])
      : [];

  return items.map((item) => {
    const source = item as Record<string, unknown>;

    const content =
      (typeof source.caption === 'string' && source.caption) ||
      (typeof source.Caption === 'string' && source.Caption) ||
      (typeof source.content === 'string' && source.content) ||
      (typeof source.Content === 'string' && source.Content) ||
      'Shared a post.';

    const imageUrls = Array.isArray(source.media_urls)
      ? source.media_urls.filter((entry): entry is string => typeof entry === 'string')
      : Array.isArray(source.mediaUrls)
        ? source.mediaUrls.filter((entry): entry is string => typeof entry === 'string')
        : Array.isArray(source.MediaUrls)
          ? source.MediaUrls.filter((entry): entry is string => typeof entry === 'string')
          : [];

    const isLiked = Boolean(
      source.isLiked ??
        source.IsLiked ??
        source.is_liked ??
        source.Isliked ??
        false,
    );

    return {
      id: String(source.post_id ?? source.PostID ?? source.id ?? `${Date.now()}-${Math.random()}`),
      content,
      imageUrls,
      likes: Number(
        source.likes ??
          source.Likes ??
          source.LikeCount ??
          source.likeCount ??
          source.like_count ??
          source.likes_count ??
          source.LikesCount ??
          0,
      ) || 0,
      comments: Number(
        source.comments ??
          source.Comments ??
          source.comments_count ??
          source.CommentsCount ??
          source.comment_count ??
          0,
      ) || 0,
      time: String(source.post_age ?? source.PostAge ?? source.created_at ?? source.CreatedAt ?? 'just now'),
      isLiked,
    };
  });
};

export async function fetchCurrentUserProfile() {
  const storedProfile = readStoredProfile();

  const response = await fetch(`${API_BASE_URL}/user/get-profile-information`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load your profile details.');
  }

  const rawProfile =
    (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
    (payload && typeof payload === 'object' && 'reponse' in payload ? (payload as Record<string, unknown>).reponse : undefined) ||
    (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
    payload;

  const profileInfo = (rawProfile && typeof rawProfile === 'object' ? (rawProfile as Record<string, unknown>) : {}) || {};
  const storedName = (storedProfile.name || '').trim();
  const storedUsername = (storedProfile.username || '').replace(/^@/, '').trim();
  const resolvedId = String(profileInfo.UserID ?? profileInfo.userID ?? profileInfo.user_id ?? storedProfile.id ?? '');
  const resolvedName = String(
    profileInfo.Name ??
      profileInfo.name ??
      (storedName && storedName.toLowerCase() !== storedUsername.toLowerCase() ? storedName : '') ??
      'User Profile',
  );
  const resolvedUsername = normalizeUsername(
    String(
      profileInfo.UserName ??
        profileInfo.userName ??
        profileInfo.username ??
        storedProfile.username ??
        'user',
    ),
  );
  const resolvedProfileImageUrl =
    (typeof profileInfo.ProfileImgUrl === 'string' && profileInfo.ProfileImgUrl) ||
    (typeof profileInfo.profile_image_url === 'string' && profileInfo.profile_image_url) ||
    (typeof profileInfo.profileImageUrl === 'string' && profileInfo.profileImageUrl) ||
    storedProfile.profileImageUrl ||
    '';
  const resolvedFollowerCount = Number(profileInfo.FollowerCount ?? profileInfo.follower_count ?? 0) || 0;
  const resolvedFollowingCount = Number(profileInfo.FollowingCount ?? profileInfo.following_count ?? 0) || 0;
  const resolvedPostCount = Number(profileInfo.PostCount ?? profileInfo.post_count ?? 0) || 0;

  const nextProfile: CurrentUserProfile = {
    id: resolvedId,
    name: resolvedName,
    username: resolvedUsername,
    profileImageUrl: resolvedProfileImageUrl,
    followerCount: resolvedFollowerCount,
    followingCount: resolvedFollowingCount,
    postCount: resolvedPostCount,
  };

  localStorage.setItem(
    PROFILE_STORAGE_KEY,
    JSON.stringify({
      ...storedProfile,
      id: resolvedId,
      name: resolvedName,
      username: resolvedUsername.startsWith('@') ? resolvedUsername.slice(1) : resolvedUsername,
      profileImageUrl: resolvedProfileImageUrl,
    }),
  );

  return nextProfile;
}

export async function fetchFollowersList(targetUserId?: string): Promise<SocialConnectionUser[]> {
  const userId = targetUserId && targetUserId !== 'me' ? targetUserId : undefined;
  const endpoint = userId
    ? `${API_BASE_URL}/user/${encodeURIComponent(userId)}/relation/followers?limit=50&offset=0`
    : `${API_BASE_URL}/user/relation/followers?limit=50&offset=0`;

  const response = await fetch(endpoint, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load followers.');
  }

  const followerList = getNestedUserList(payload);
  return followerList
    .map((entry) => normalizeSocialConnectionUser(entry))
    .filter((entry): entry is SocialConnectionUser => Boolean(entry));
}

export async function fetchFollowingList(targetUserId?: string): Promise<SocialConnectionUser[]> {
  const userId = targetUserId && targetUserId !== 'me' ? targetUserId : undefined;
  const endpoint = userId
    ? `${API_BASE_URL}/user/${encodeURIComponent(userId)}/relation/following?limit=50&offset=0`
    : `${API_BASE_URL}/user/relation/following?limit=50&offset=0`;

  const response = await fetch(endpoint, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load following.');
  }

  const followingList = getNestedUserList(payload);
  return followingList
    .map((entry) => normalizeSocialConnectionUser(entry))
    .filter((entry): entry is SocialConnectionUser => Boolean(entry));
}

export async function logoutCurrentUser(): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/user/logout`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to log out right now.');
  }
}

export async function fetchProfilePageData(targetUserId?: string): Promise<ProfilePageData> {
  const isOwnProfileRequest = !targetUserId || targetUserId === 'me';
  const profileEndpoint = isOwnProfileRequest ? `${API_BASE_URL}/user/posts` : `${API_BASE_URL}/user/${targetUserId}/posts`;
  const response = await fetch(profileEndpoint, {
    method: 'GET',
    credentials: 'include',
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load this profile.');
  }

  const data = Array.isArray((payload as Record<string, unknown>).data)
    ? ((payload as Record<string, unknown>).data as unknown[])
    : Array.isArray(payload)
      ? (payload as unknown[])
      : [];

  const posts = normalizeProfilePosts(data);

  if (!isOwnProfileRequest) {
    const publicProfileResponse = await fetch(`${API_BASE_URL}/user/${targetUserId}/get-public-profile`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const publicProfilePayload = await publicProfileResponse.json().catch(() => ({}));

    if (!publicProfileResponse.ok || (publicProfilePayload && typeof publicProfilePayload === 'object' && 'status_code' in publicProfilePayload && Number((publicProfilePayload as Record<string, unknown>).status_code) >= 400)) {
      throw new Error((publicProfilePayload as Record<string, unknown>).message ? String((publicProfilePayload as Record<string, unknown>).message) : 'Unable to load this profile.');
    }

    const userInfo =
      (publicProfilePayload && typeof publicProfilePayload === 'object' && 'user_info' in publicProfilePayload
        ? (publicProfilePayload as Record<string, unknown>).user_info
        : publicProfilePayload) as Record<string, unknown>;
    const socialStats =
      (publicProfilePayload && typeof publicProfilePayload === 'object' && 'social_stats' in publicProfilePayload
        ? (publicProfilePayload as Record<string, unknown>).social_stats
        : {}) as Record<string, unknown>;

    const user: ProfileUser = {
      id: String(targetUserId),
      name: String(
        (typeof userInfo.Name === 'string' && userInfo.Name) ||
          (typeof userInfo.name === 'string' && userInfo.name) ||
          'User Profile',
      ),
      username: normalizeUsername(
        String(
          (typeof userInfo.UserName === 'string' && userInfo.UserName) ||
            (typeof userInfo.userName === 'string' && userInfo.userName) ||
            (typeof userInfo.username === 'string' && userInfo.username) ||
            '@user',
        ),
      ),
      profileImageUrl:
        (typeof userInfo.ProfileImgUrl === 'string' && userInfo.ProfileImgUrl) ||
        (typeof userInfo.profile_image_url === 'string' && userInfo.profile_image_url) ||
        (typeof userInfo.profileImageUrl === 'string' && userInfo.profileImageUrl) ||
        undefined,
      bio:
        (typeof userInfo.Bio === 'string' && userInfo.Bio) ||
        (typeof userInfo.bio === 'string' && userInfo.bio) ||
        undefined,
      website:
        (typeof userInfo.Website === 'string' && userInfo.Website) ||
        (typeof userInfo.website === 'string' && userInfo.website) ||
        undefined,
      linkedin:
        (typeof userInfo.LinkedIn === 'string' && userInfo.LinkedIn) ||
        (typeof userInfo.linkedin === 'string' && userInfo.linkedin) ||
        undefined,
      instagram:
        (typeof userInfo.Instagram === 'string' && userInfo.Instagram) ||
        (typeof userInfo.instagram === 'string' && userInfo.instagram) ||
        undefined,
      followerCount: Number(socialStats.followers ?? socialStats.FollowerCount ?? 0) || 0,
      followingCount: Number(socialStats.following ?? socialStats.FollowingCount ?? 0) || 0,
    };

    return {
      posts,
      user,
      followerCount: user.followerCount ?? 0,
      followingCount: user.followingCount ?? 0,
      isFollowing: Boolean(
        socialStats.is_already_following ??
          socialStats.isAlreadyFollowing ??
          socialStats.IsAlreadyFollowing ??
          false,
      ),
    };
  }

  let profileInfo: Record<string, unknown> = {};
  try {
    const profileInfoResponse = await fetch(`${API_BASE_URL}/user/get-profile-information`, {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (profileInfoResponse.ok) {
      const profilePayload = await profileInfoResponse.json().catch(() => ({}));
      const nestedProfileData =
        (profilePayload && typeof profilePayload === 'object' && 'reponse' in profilePayload
          ? (profilePayload as Record<string, unknown>).reponse
          : undefined) ||
        (profilePayload && typeof profilePayload === 'object' && 'response' in profilePayload
          ? (profilePayload as Record<string, unknown>).response
          : undefined) ||
        (profilePayload && typeof profilePayload === 'object' && 'data' in profilePayload
          ? (profilePayload as Record<string, unknown>).data
          : undefined) ||
        profilePayload;

      profileInfo = nestedProfileData && typeof nestedProfileData === 'object' ? (nestedProfileData as Record<string, unknown>) : {};
    }
  } catch (profileInfoError) {
    console.error('fetch profile info failed', profileInfoError);
  }

  const firstItem = data[0] as Record<string, unknown> | undefined;
  const dynamicUserMeta =
    (firstItem?.user_data as Record<string, unknown>) ||
    (firstItem?.userData as Record<string, unknown>) ||
    (firstItem?.UserData as Record<string, unknown>) ||
    {};

  const storedProfile = readStoredProfile() as Partial<ProfileUser> & {
    bio?: string;
    website?: string;
    linkedin?: string;
    instagram?: string;
    customLinks?: CustomLink[];
  };
  const resolvedName =
    (typeof profileInfo.Name === 'string' && profileInfo.Name) ||
    (typeof profileInfo.name === 'string' && profileInfo.name) ||
    storedProfile.name ||
    ((typeof dynamicUserMeta.name === 'string' && dynamicUserMeta.name) ||
      (typeof dynamicUserMeta.Name === 'string' && dynamicUserMeta.Name) ||
      'User Profile');

  const resolvedUsername = normalizeUsername(
    String(
      (typeof profileInfo.UserName === 'string' && profileInfo.UserName) ||
        (typeof profileInfo.userName === 'string' && profileInfo.userName) ||
        (typeof profileInfo.username === 'string' && profileInfo.username) ||
        storedProfile.username ||
        (typeof dynamicUserMeta.username === 'string' && dynamicUserMeta.username) ||
        (typeof dynamicUserMeta.UserName === 'string' && dynamicUserMeta.UserName) ||
        '@user',
    ),
  );

  const resolvedBio =
    (typeof profileInfo.Bio === 'string' && profileInfo.Bio) ||
    (typeof profileInfo.bio === 'string' && profileInfo.bio) ||
    storedProfile.bio ||
    undefined;

  const resolvedWebsite =
    (typeof profileInfo.Website === 'string' && profileInfo.Website) ||
    (typeof profileInfo.website === 'string' && profileInfo.website) ||
    storedProfile.website ||
    undefined;

  const resolvedLinkedin =
    (typeof profileInfo.LinkedIn === 'string' && profileInfo.LinkedIn) ||
    (typeof profileInfo.linkedin === 'string' && profileInfo.linkedin) ||
    storedProfile.linkedin ||
    undefined;

  const resolvedInstagram =
    (typeof profileInfo.Instagram === 'string' && profileInfo.Instagram) ||
    (typeof profileInfo.instagram === 'string' && profileInfo.instagram) ||
    storedProfile.instagram ||
    undefined;

  const resolvedCustomLinks = Array.isArray(profileInfo.customLinks)
    ? ((profileInfo.customLinks as Array<unknown>)
        .filter((entry) => !!entry && typeof entry === 'object') as Array<Record<string, unknown>>)
        .map((entry, index) => ({
          id: typeof entry.id === 'string' ? entry.id : `link-${index}`,
          label: typeof entry.label === 'string' ? entry.label : 'Link',
          url: typeof entry.url === 'string' ? entry.url : '',
        }))
        .filter((entry) => entry.url.trim())
    : Array.isArray(storedProfile.customLinks)
      ? storedProfile.customLinks
      : [];

  const resolvedProfileImageUrl =
    (typeof profileInfo.ProfileImgUrl === 'string' && profileInfo.ProfileImgUrl) ||
    (typeof profileInfo.profile_image_url === 'string' && profileInfo.profile_image_url) ||
    (typeof profileInfo.profileImageUrl === 'string' && profileInfo.profileImageUrl) ||
    storedProfile.profileImageUrl ||
    ((typeof dynamicUserMeta.profile_image_url === 'string' && dynamicUserMeta.profile_image_url) ||
      (typeof dynamicUserMeta.ProfileImgUrl === 'string' && dynamicUserMeta.ProfileImgUrl) ||
      undefined);

  const resolvedFollowerCount = Number(profileInfo.FollowerCount ?? profileInfo.follower_count ?? 0) || 0;
  const resolvedFollowingCount = Number(profileInfo.FollowingCount ?? profileInfo.following_count ?? 0) || 0;

  return {
    posts,
    user: {
      id: String(targetUserId ?? 'me'),
      name: resolvedName,
      username: resolvedUsername,
      profileImageUrl: resolvedProfileImageUrl,
      bio: resolvedBio,
      website: resolvedWebsite,
      linkedin: resolvedLinkedin,
      instagram: resolvedInstagram,
      customLinks: resolvedCustomLinks,
      followerCount: resolvedFollowerCount,
      followingCount: resolvedFollowingCount,
    },
    followerCount: resolvedFollowerCount,
    followingCount: resolvedFollowingCount,
    isFollowing: false,
  };
}

