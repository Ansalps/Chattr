import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

type FeedComment = {
  id: string;
  author: string;
  text: string;
  time: string;
  profileImageUrl?: string;
  replyCount?: number;
  replies?: FeedComment[];
};

type FeedPost = {
  id: string;
  userId?: string;
  author: string;
  handle: string;
  time: string;
  content: string;
  imageUrls: string[];
  profileImageUrl?: string;
  isLiked: boolean;
  likes: number;
  comments: number;
  shares: number;
  accent: string;
  tag: string;
};

const fallbackPosts: FeedPost[] = [
  {
    id: 'welcome-1',
    author: 'Chattr Team',
    handle: '@chattr',
    time: '2h ago',
    content:
      'Welcome to your new home feed. Start following creators, share updates, and keep the conversation moving with your community.',
    imageUrls: [],
    isLiked: false,
    likes: 324,
    comments: 19,
    shares: 12,
    accent: 'from-sky-400 to-blue-600',
    tag: 'Community',
  },
  {
    id: 'welcome-2',
    author: 'Maya Johnson',
    handle: '@maya',
    time: '4h ago',
    content:
      'Designing clearer digital experiences for small teams. Small product choices really do build trust over time.',
    imageUrls: [],
    isLiked: true,
    likes: 188,
    comments: 24,
    shares: 8,
    accent: 'from-violet-400 to-fuchsia-600',
    tag: 'Product',
  },
  {
    id: 'welcome-3',
    author: 'Alex Rivers',
    handle: '@alex',
    time: 'Yesterday',
    content:
      'Shipping one thoughtful feature a day beats chasing every idea. Keep it simple, measurable, and useful.',
    imageUrls: [],
    isLiked: false,
    likes: 542,
    comments: 38,
    shares: 21,
    accent: 'from-emerald-400 to-teal-600',
    tag: 'Growth',
  },
];

const formatCount = (value: number) =>
  new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 }).format(value);

const formatFriendlyTime = (value: unknown): string => {
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

const formatChatMessageTime = (value: unknown): string => {
  if (typeof value !== 'string') {
    return 'now';
  }

  const trimmed = value.trim();
  if (!trimmed) return 'now';

  const parsedDate = new Date(trimmed);
  if (Number.isNaN(parsedDate.getTime())) {
    return trimmed;
  }

  return parsedDate.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  });
};

const PROFILE_STORAGE_KEY = 'chattr_profile_info';

const readStoredProfile = () => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) return { id: '', name: '', username: '', profileImageUrl: '' };

    const parsed = JSON.parse(raw) as Partial<{
      id?: string | number;
      userId?: string | number;
      name?: string;
      username?: string;
      profileImageUrl?: string;
    }>;
    return {
      id: parsed.id !== undefined && parsed.id !== null ? String(parsed.id) : parsed.userId !== undefined && parsed.userId !== null ? String(parsed.userId) : '',
      name: typeof parsed.name === 'string' ? parsed.name : '',
      username: typeof parsed.username === 'string' ? parsed.username : '',
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : '',
    };
  } catch (error) {
    return { id: '', name: '', username: '', profileImageUrl: '' };
  }
};

const getInitials = (value: string) =>
  value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'A';

const getUsernameInitial = (value: string) => {
  const trimmed = value.replace(/^@/, '').trim();
  return trimmed ? trimmed.charAt(0).toUpperCase() : 'A';
};

const normalizeUsername = (value: string) => {
  const trimmed = value.trim();
  if (!trimmed) return '@user';
  return trimmed.startsWith('@') ? trimmed : `@${trimmed}`;
};

const normalizeStringArray = (value: unknown): string[] => {
  if (Array.isArray(value)) {
    return value
      .map((entry) => {
        if (typeof entry === 'string') return entry;
        if (entry && typeof entry === 'object') {
          const item = entry as Record<string, unknown>;
          return typeof item.url === 'string' ? item.url : typeof item.media_url === 'string' ? item.media_url : '';
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

const getPostIdFromResponse = (value: unknown): string | null => {
  if (!value || typeof value !== 'object') return null;

  const source = value as Record<string, unknown>;
  const candidates = [
    source.res,
    source.data,
    source,
  ];

  for (const candidate of candidates) {
    if (!candidate || typeof candidate !== 'object') continue;

    const item = candidate as Record<string, unknown>;
    const rawId =
      item.post_id ??
      item.PostId ??
      item.postId ??
      item.PostID ??
      item.post_id ??
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

const normalizeFeedPosts = (value: unknown): FeedPost[] => {
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

      const nestedUserData =
        (source.UserData as Record<string, unknown>) ||
        (source.userData as Record<string, unknown>) ||
        {};

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
          source.isLiked ??
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
        accent:
          typeof source.accent === 'string' && source.accent
            ? source.accent
            : 'from-sky-400 to-blue-600',
        tag: typeof source.tag === 'string' && source.tag ? source.tag : 'Update',
      };
    })
    .filter((post) => !!post.content || post.imageUrls.length > 0);
};

export default function NewsFeedPage() {
  const navigate = useNavigate();
  const [posts, setPosts] = useState<FeedPost[]>(fallbackPosts);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [isComposerOpen, setIsComposerOpen] = useState(false);
  const [postText, setPostText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isPosting, setIsPosting] = useState(false);
  const [composerError, setComposerError] = useState('');
  const [composerSuccess, setComposerSuccess] = useState('');
  const [commentDrafts, setCommentDrafts] = useState<Record<string, string>>({});
  const [replyDrafts, setReplyDrafts] = useState<Record<string, string>>({});
  const [replyingToComment, setReplyingToComment] = useState<Record<string, string | null>>({});
  const [expandedRepliesByComment, setExpandedRepliesByComment] = useState<Record<string, boolean>>({});
  const [openCommentFor, setOpenCommentFor] = useState<string | null>(null);
  const [commentsByPost, setCommentsByPost] = useState<Record<string, FeedComment[]>>({});
  const [commentLoading, setCommentLoading] = useState<Record<string, boolean>>({});
  const [postMenuOpenFor, setPostMenuOpenFor] = useState<string | null>(null);
  const [editingPostId, setEditingPostId] = useState<string | null>(null);
  const [currentUserProfile, setCurrentUserProfile] = useState({
    id: '',
    name: '',
    username: '',
    profileImageUrl: '',
    followerCount: 0,
    followingCount: 0,
    postCount: 0,
  });
  const [userSearch, setUserSearch] = useState('');
  const [userSearchResults, setUserSearchResults] = useState<
    Array<{ id: string; name: string; username: string; profileImageUrl: string }>
  >([]);
  const [isSearchingUsers, setIsSearchingUsers] = useState(false);
  const [isChatOpen, setIsChatOpen] = useState(false);
  const [chatSearch, setChatSearch] = useState('');
  const [chatSearchResults, setChatSearchResults] = useState<
    Array<{ id: string; name: string; username: string; profileImageUrl: string }>
  >([]);
  const [isSearchingChatUsers, setIsSearchingChatUsers] = useState(false);
  const [chatSessions, setChatSessions] = useState<
    Array<{
      id: string;
      conversationId: string;
      name: string;
      username: string;
      profileImageUrl: string;
      lastMessage: string;
      lastMessageTime: string;
    }>
  >([]);
  const [chatSelectedUser, setChatSelectedUser] = useState<{
    id: string;
    conversationId?: string;
    name: string;
    username: string;
    profileImageUrl: string;
  } | null>(null);
  const [chatMessages, setChatMessages] = useState<
    Array<{ id: string; senderId: string; content: string; createdAt: string; mine: boolean }>
  >([]);
  const [chatDraft, setChatDraft] = useState('');
  const [chatSocket, setChatSocket] = useState<WebSocket | null>(null);
  const [chatError, setChatError] = useState('');
  const [isLoadingChatHistory, setIsLoadingChatHistory] = useState(false);

  const storedProfile = readStoredProfile();
  const currentUserId = currentUserProfile.id || storedProfile.id || '';
  const storedDisplayName = (storedProfile.name || '').trim();
  const storedUserName = (storedProfile.username || '').replace(/^@/, '').trim();
  const currentUserName = (() => {
    const liveDisplayName = (currentUserProfile.name || '').trim();
    const liveUserName = (currentUserProfile.username || '').replace(/^@/, '').trim();

    if (liveDisplayName && liveDisplayName.toLowerCase() !== liveUserName.toLowerCase()) {
      return liveDisplayName;
    }

    if (storedDisplayName && storedDisplayName.toLowerCase() !== storedUserName.toLowerCase()) {
      return storedDisplayName;
    }

    return 'User Profile';
  })();
  const currentUserHandle =
    currentUserProfile.username || storedProfile.username || '@user';
  const currentProfileImageUrl =
    (currentUserProfile.profileImageUrl ? currentUserProfile.profileImageUrl : storedProfile.profileImageUrl) || '';

  useEffect(() => {
    const fetchCurrentUserProfile = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/user/get-profile-information`, {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        const payload = await response.json().catch(() => ({}));

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || 'Unable to load your profile details.');
        }

        const rawProfile =
          (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
          (payload && typeof payload === 'object' && 'reponse' in payload ? (payload as Record<string, unknown>).reponse : undefined) ||
          (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
          payload;

        const profileInfo = (rawProfile && typeof rawProfile === 'object' ? rawProfile as Record<string, unknown> : {}) || {};
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

        const nextProfile = {
          id: resolvedId,
          name: resolvedName,
          username: resolvedUsername,
          profileImageUrl: resolvedProfileImageUrl,
          followerCount: resolvedFollowerCount,
          followingCount: resolvedFollowingCount,
          postCount: resolvedPostCount,
        };

        setCurrentUserProfile(nextProfile);
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
      } catch (profileError) {
        console.error('fetch user profile info failed', profileError);
      }
    };

    const fetchNewsFeed = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/user/newsfeed`, {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
        });

        const payload = await response.json();

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || 'Unable to load your news feed right now.');
        }

        const feedData =
          Array.isArray(payload)
            ? payload
            : Array.isArray(payload.PostUserData)
              ? payload.PostUserData
              : Array.isArray(payload.postUserData)
                ? payload.postUserData
                : Array.isArray(payload.data)
                  ? payload.data
                  : Array.isArray(payload.data?.posts)
                    ? payload.data.posts
                    : Array.isArray(payload.data?.PostUserData)
                      ? payload.data.PostUserData
                      : Array.isArray(payload.data?.newsfeed)
                        ? payload.data.newsfeed
                        : Array.isArray(payload.data?.items)
                          ? payload.data.items
                          : [];

        setPosts(normalizeFeedPosts(feedData));
      } catch (feedError) {
        const message =
          feedError instanceof Error ? feedError.message : 'Unable to load your news feed right now.';

        setError(message);
        setPosts(fallbackPosts);
      } finally {
        setIsLoading(false);
      }
    };

    fetchCurrentUserProfile();
    fetchNewsFeed();
  }, []);

  const quickStats = useMemo(
    () => [
      { label: 'Followers', value: formatCount(currentUserProfile.followerCount) },
      { label: 'Following', value: formatCount(currentUserProfile.followingCount) },
      { label: 'Posts', value: formatCount(currentUserProfile.postCount) },
    ],
    [currentUserProfile.followerCount, currentUserProfile.followingCount, currentUserProfile.postCount],
  );

  const sidebarPeople = [
    { name: 'Rina Shah', role: 'Product Lead', accent: 'from-pink-400 to-rose-500' },
    { name: 'Noah Kim', role: 'UX Strategist', accent: 'from-sky-400 to-cyan-500' },
    { name: 'Aisha Noor', role: 'Community Builder', accent: 'from-violet-400 to-purple-500' },
  ];

  const goToProfile = (userId?: string) => {
    const safeUserId = String(userId ?? '').trim();
    navigate(safeUserId && safeUserId !== 'me' ? `/profile/${encodeURIComponent(safeUserId)}` : '/profile');
  };

  const handleUserSearchSelect = (userId: string) => {
    setUserSearch('');
    setUserSearchResults([]);
    navigate(`/profile/${encodeURIComponent(userId)}`);
  };

  const fetchRecentChatProfiles = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/user/get-recent-chat-profiles?page=1&limit=20`, {
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });

      if (!response.ok) {
        throw new Error('Unable to load recent chats.');
      }

      const payload = await response.json().catch(() => ({}));
      const data =
        (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
        (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
        payload;
      const chatProfiles = Array.isArray((data as Record<string, unknown>)?.ChatProfiles)
        ? ((data as Record<string, unknown>).ChatProfiles as unknown[])
        : Array.isArray((data as Record<string, unknown>)?.chatProfiles)
          ? ((data as Record<string, unknown>).chatProfiles as unknown[])
          : Array.isArray((data as Record<string, unknown>)?.data)
            ? ((data as Record<string, unknown>).data as unknown[])
            : [];

      const normalizedChats = chatProfiles
        .map((conversation) => {
          const source = (conversation ?? {}) as Record<string, unknown>;
          const conversationId = String(
            source.chat_id ?? source.ChatID ?? source.chatId ?? source.ConversationID ?? source.conversation_id ?? '',
          );
          const name = String(
            source.chat_name ?? source.ChatName ?? source.chatName ?? source.name ?? source.Name ?? 'User',
          );
          const username = normalizeUsername(
            String(source.username ?? source.UserName ?? source.userName ?? source.name ?? source.Name ?? '@user'),
          );
          const profileImageUrl =
            (typeof source.chat_image === 'string' && source.chat_image) ||
            (typeof source.ChatImage === 'string' && source.ChatImage) ||
            (typeof source.chatImage === 'string' && source.chatImage) ||
            (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
            (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
            '';
          const lastMessage = String(source.last_message ?? source.LastMessage ?? '');
          const lastMessageTime = formatFriendlyTime(
            source.last_message_time ?? source.LastMessageTime ?? source.lastMessageTime ?? new Date().toISOString(),
          );

          if (!conversationId && !name) return null;

          return {
            id: String(source.user_id ?? source.UserID ?? source.userId ?? source.UserId ?? name ?? conversationId),
            conversationId,
            name,
            username,
            profileImageUrl,
            lastMessage,
            lastMessageTime,
          };
        })
        .filter(Boolean) as Array<{
          id: string;
          conversationId: string;
          name: string;
          username: string;
          profileImageUrl: string;
          lastMessage: string;
          lastMessageTime: string;
        }>;

      setChatSessions(normalizedChats);
    } catch (chatError) {
      console.error('loading recent chats failed', chatError);
      setChatSessions([]);
    }
  };

  const openChatSession = async (user: {
    id: string;
    conversationId?: string;
    name: string;
    username: string;
    profileImageUrl: string;
  }) => {
    setChatSelectedUser(user);
    setChatSearch('');
    setChatSearchResults([]);
    setChatDraft('');
    setChatError('');
    setChatMessages([]);
    setIsLoadingChatHistory(true);

    try {
      if (!user.conversationId) {
        return;
      }

      const historyResponse = await fetch(`${API_BASE_URL}/user/chat/${encodeURIComponent(user.conversationId)}?page=1&limit=50`, {
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });

      if (!historyResponse.ok) {
        throw new Error('Unable to load chat history.');
      }

      const historyPayload = await historyResponse.json().catch(() => ({}));
      const historyData =
        (historyPayload && typeof historyPayload === 'object' && 'data' in historyPayload
          ? (historyPayload as Record<string, unknown>).data
          : historyPayload) ?? {};
      const historyMessages = Array.isArray((historyData as Record<string, unknown>)?.messages)
        ? ((historyData as Record<string, unknown>).messages as unknown[])
        : Array.isArray((historyData as Record<string, unknown>)?.Messages)
          ? ((historyData as Record<string, unknown>).Messages as unknown[])
          : [];

      const normalizedMessages = historyMessages.map((message, index) => {
        const item = (message ?? {}) as Record<string, unknown>;
        const senderId = String(item.sender_id ?? item.SenderID ?? item.senderId ?? '');
        const content = String(item.content ?? item.Content ?? '');
        const createdAt = String(item.created_at ?? item.CreatedAt ?? item.createdAt ?? new Date().toISOString());

        return {
          id: String(item.message_id ?? item.MessageID ?? item.id ?? `${senderId}-${createdAt}-${index}`),
          senderId,
          content,
          createdAt,
          mine: senderId === (currentUserId || storedProfile.id || ''),
        };
      });

      setChatMessages(normalizedMessages.filter((message) => message.content.trim()));
    } catch (historyError) {
      console.error('loading chat history failed', historyError);
      setChatMessages([]);
    } finally {
      setIsLoadingChatHistory(false);
    }
  };

  useEffect(() => {
    const trimmedQuery = userSearch.trim();
    if (!trimmedQuery) {
      setUserSearchResults([]);
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        setIsSearchingUsers(true);
        const response = await fetch(`${API_BASE_URL}/user/search?username=${encodeURIComponent(trimmedQuery)}&limit=5&offset=0`, {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          signal: controller.signal,
        });

        const payload = await response.json().catch(() => ({}));

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || 'Unable to search users.');
        }

        const dataPayload =
          (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
          (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
          payload;

        const userList: unknown[] =
          Array.isArray((dataPayload as Record<string, unknown>)?.UserMeataData)
            ? ((dataPayload as Record<string, unknown>).UserMeataData as unknown[])
            : Array.isArray((dataPayload as Record<string, unknown>)?.userMeataData)
              ? ((dataPayload as Record<string, unknown>).userMeataData as unknown[])
              : Array.isArray((dataPayload as Record<string, unknown>)?.data)
                ? ((dataPayload as Record<string, unknown>).data as unknown[])
                : Array.isArray((dataPayload as Record<string, unknown>)?.Data)
                  ? ((dataPayload as Record<string, unknown>).Data as unknown[])
                  : [];

        const normalizedResults = userList
          .map((user) => {
            const source = (user ?? {}) as Record<string, unknown>;
            const userId = String(
              source.UserId ?? source.userId ?? source.user_id ?? source.UserID ?? source.userID ?? '',
            );
            const name = String(
              source.Name ?? source.name ?? source.FullName ?? source.full_name ?? source.UserName ?? source.userName ?? 'User',
            );
            const username = String(
              source.UserName ?? source.userName ?? source.username ?? '@user',
            );
            const profileImageUrl =
              (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
              (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
              (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
              '';

            if (!userId && !name && !username) return null;

            return {
              id: userId || name || username,
              name,
              username: normalizeUsername(username),
              profileImageUrl,
            };
          })
          .filter(Boolean) as Array<{ id: string; name: string; username: string; profileImageUrl: string }>;

        setUserSearchResults(normalizedResults);
      } catch (searchError) {
        if (searchError instanceof DOMException && searchError.name === 'AbortError') {
          return;
        }

        console.error('search users failed', searchError);
        setUserSearchResults([]);
      } finally {
        setIsSearchingUsers(false);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [userSearch]);

  useEffect(() => {
    if (isChatOpen) {
      fetchRecentChatProfiles();
    }
  }, [isChatOpen]);

  useEffect(() => {
    const trimmedQuery = chatSearch.trim();
    if (!trimmedQuery) {
      setChatSearchResults([]);
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        setIsSearchingChatUsers(true);
        const response = await fetch(`${API_BASE_URL}/user/search?username=${encodeURIComponent(trimmedQuery)}&limit=5&offset=0`, {
          method: 'GET',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          signal: controller.signal,
        });

        const payload = await response.json().catch(() => ({}));

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || 'Unable to search chat users.');
        }

        const dataPayload =
          (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
          (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
          payload;

        const userList: unknown[] =
          Array.isArray((dataPayload as Record<string, unknown>)?.UserMeataData)
            ? ((dataPayload as Record<string, unknown>).UserMeataData as unknown[])
            : Array.isArray((dataPayload as Record<string, unknown>)?.userMeataData)
              ? ((dataPayload as Record<string, unknown>).userMeataData as unknown[])
              : Array.isArray((dataPayload as Record<string, unknown>)?.data)
                ? ((dataPayload as Record<string, unknown>).data as unknown[])
                : Array.isArray((dataPayload as Record<string, unknown>)?.Data)
                  ? ((dataPayload as Record<string, unknown>).Data as unknown[])
                  : [];

        const normalizedResults = userList
          .map((user) => {
            const source = (user ?? {}) as Record<string, unknown>;
            const userId = String(
              source.UserId ?? source.userId ?? source.user_id ?? source.UserID ?? source.userID ?? '',
            );
            const name = String(
              source.Name ?? source.name ?? source.FullName ?? source.full_name ?? source.UserName ?? source.userName ?? 'User',
            );
            const username = String(
              source.UserName ?? source.userName ?? source.username ?? '@user',
            );
            const profileImageUrl =
              (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
              (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
              (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
              '';

            if (!userId && !name && !username) return null;

            return {
              id: userId || name || username,
              name,
              username: normalizeUsername(username),
              profileImageUrl,
            };
          })
          .filter(Boolean) as Array<{ id: string; name: string; username: string; profileImageUrl: string }>;

        setChatSearchResults(normalizedResults.filter((user) => user.id !== (currentUserId || storedProfile.id || '')));
      } catch (searchError) {
        if (searchError instanceof DOMException && searchError.name === 'AbortError') {
          return;
        }

        console.error('search chat users failed', searchError);
        setChatSearchResults([]);
      } finally {
        setIsSearchingChatUsers(false);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [chatSearch, currentUserId, storedProfile.id]);

  useEffect(() => {
    if (!chatSelectedUser) {
      return;
    }

    const socketUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:3000/user/ws`;
    const socket = new WebSocket(socketUrl);

    socket.onopen = () => {
      setChatError('');
    };

    socket.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data as string);

        if (payload && payload.error) {
          setChatError(String(payload.error));
          return;
        }

        const senderId = String(payload.sender_id ?? payload.SenderID ?? payload.senderId ?? '');
        const content = String(payload.content ?? payload.Content ?? '');
        if (!senderId && !content) {
          return;
        }

        setChatMessages((currentMessages) => {
          const nextMessage = {
            id: String(payload.message_id ?? payload.MessageID ?? payload.id ?? `${senderId}-${content}-${Date.now()}`),
            senderId,
            content,
            createdAt: String(payload.created_at ?? payload.CreatedAt ?? payload.createdAt ?? new Date().toISOString()),
            mine: senderId === (currentUserId || storedProfile.id || ''),
          };

          const alreadyExists = currentMessages.some((message) => message.id === nextMessage.id);
          return alreadyExists ? currentMessages : [...currentMessages, nextMessage];
        });
      } catch (error) {
        console.error('failed to parse chat websocket payload', error);
      }
    };

    socket.onerror = () => {
      setChatError('Unable to connect to chat right now.');
    };

    setChatSocket(socket);

    return () => {
      socket.close();
      setChatSocket(null);
    };
  }, [chatSelectedUser, currentUserId, storedProfile.id]);

  const handleSendChatMessage = () => {
    if (!chatSelectedUser) {
      return;
    }

    const trimmedMessage = chatDraft.trim();
    if (!trimmedMessage) {
      return;
    }

    if (!chatSocket || chatSocket.readyState !== WebSocket.OPEN) {
      setChatError('Chat is not connected yet. Please try again.');
      return;
    }

    const messagePayload = {
      type: 'individual',
      recipient_id: Number(chatSelectedUser.id),
      content: trimmedMessage,
      created_at: new Date().toISOString(),
    };

    chatSocket.send(JSON.stringify(messagePayload));
    setChatMessages((currentMessages) => [
      ...currentMessages,
      {
        id: `local-${Date.now()}`,
        senderId: currentUserId || storedProfile.id || 'me',
        content: trimmedMessage,
        createdAt: new Date().toISOString(),
        mine: true,
      },
    ]);
    setChatDraft('');
    setChatError('');
  };

  const handleToggleLike = async (postId: string, currentLiked: boolean) => {
    try {
      const endpoint = `${API_BASE_URL}/user/post/${postId}/${currentLiked ? 'unlike' : 'like'}`;
      console.log(`post Id: ${postId}`)
      const response = await fetch(endpoint, {
        method: currentLiked ? 'DELETE' : 'POST',
        credentials: 'include',
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload.message || payload.error || 'Unable to update like status.');
      }

      setPosts((currentPosts) =>
        currentPosts.map((post) => {
          if (post.id !== postId) return post;
          const nextLiked = !currentLiked;
          return {
            ...post,
            isLiked: nextLiked,
            likes: Math.max(0, post.likes + (nextLiked ? 1 : -1)),
          };
        }),
      );
    } catch (error) {
      console.error('toggle like failed', error);
    }
  };

  const fetchCommentsForPost = async (postId: string) => {
    if (commentLoading[postId]) return;

    setCommentLoading((current) => ({ ...current, [postId]: true }));

    try {
      const response = await fetch(`${API_BASE_URL}/user/post/${postId}/comment?limit=20&offset=0`, {
        method: 'GET',
        credentials: 'include',
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || payload.error || 'Unable to load comments.');
      }

      const data = payload.data ?? payload;
      const commentList = Array.isArray(data?.Comments)
        ? data.Comments
        : Array.isArray(data?.comments)
          ? data.comments
          : Array.isArray(payload.Comments)
            ? payload.Comments
            : Array.isArray(payload.comments)
              ? payload.comments
              : [];

      const normalizedComments: FeedComment[] = commentList.map((comment: Record<string, unknown>, index: number) => {
        const userDetails =
          (comment.UserDetails as Record<string, unknown>) ||
          (comment.userDetails as Record<string, unknown>) ||
          (comment.user_details as Record<string, unknown>) ||
          {};

        const author =
          (typeof comment.Name === 'string' && comment.Name) ||
          (typeof comment.name === 'string' && comment.name) ||
          (typeof userDetails.Name === 'string' && userDetails.Name) ||
          (typeof userDetails.name === 'string' && userDetails.name) ||
          (typeof userDetails.UserName === 'string' && userDetails.UserName) ||
          (typeof userDetails.userName === 'string' && userDetails.userName) ||
          'User';

        const text =
          (typeof comment.CommentText === 'string' && comment.CommentText) ||
          (typeof comment.comment_text === 'string' && comment.comment_text) ||
          (typeof comment.commentText === 'string' && comment.commentText) ||
          'Comment';

        const profileImageUrl =
          (typeof comment.ProfileImgUrl === 'string' && comment.ProfileImgUrl) ||
          (typeof comment.profile_image_url === 'string' && comment.profile_image_url) ||
          (typeof comment.profileImageUrl === 'string' && comment.profileImageUrl) ||
          (typeof userDetails.ProfileImgUrl === 'string' && userDetails.ProfileImgUrl) ||
          (typeof userDetails.profile_image_url === 'string' && userDetails.profile_image_url) ||
          (typeof userDetails.profileImageUrl === 'string' && userDetails.profileImageUrl) ||
          '';

        const time =
          (typeof comment.CommentAge === 'string' && comment.CommentAge) ||
          (typeof comment.comment_age === 'string' && comment.comment_age) ||
          'just now';

        const rawChildComments = Array.isArray(comment.ChildComment)
          ? comment.ChildComment
          : Array.isArray(comment.childComment)
            ? comment.childComment
            : [];

        const replies = rawChildComments.map((child: Record<string, unknown>, childIndex: number) => {
          const childUserDetails =
            (child.UserDetails as Record<string, unknown>) ||
            (child.userDetails as Record<string, unknown>) ||
            (child.user_details as Record<string, unknown>) ||
            {};

          const childAuthor =
            (typeof child.Name === 'string' && child.Name) ||
            (typeof child.name === 'string' && child.name) ||
            (typeof childUserDetails.Name === 'string' && childUserDetails.Name) ||
            (typeof childUserDetails.name === 'string' && childUserDetails.name) ||
            (typeof childUserDetails.UserName === 'string' && childUserDetails.UserName) ||
            (typeof childUserDetails.userName === 'string' && childUserDetails.userName) ||
            'User';

          const childText =
            (typeof child.CommentText === 'string' && child.CommentText) ||
            (typeof child.comment_text === 'string' && child.comment_text) ||
            (typeof child.commentText === 'string' && child.commentText) ||
            'Comment';

          const childProfileImageUrl =
            (typeof child.ProfileImgUrl === 'string' && child.ProfileImgUrl) ||
            (typeof child.profile_image_url === 'string' && child.profile_image_url) ||
            (typeof child.profileImageUrl === 'string' && child.profileImageUrl) ||
            (typeof childUserDetails.ProfileImgUrl === 'string' && childUserDetails.ProfileImgUrl) ||
            (typeof childUserDetails.profile_image_url === 'string' && childUserDetails.profile_image_url) ||
            (typeof childUserDetails.profileImageUrl === 'string' && childUserDetails.profileImageUrl) ||
            '';

          const childTime =
            (typeof child.CommentAge === 'string' && child.CommentAge) ||
            (typeof child.comment_age === 'string' && child.comment_age) ||
            'just now';

          return {
            id: String(child.CommentID ?? child.comment_id ?? child.id ?? `${postId}-${index}-${childIndex}`),
            author: childAuthor,
            text: childText,
            time: childTime,
            profileImageUrl: childProfileImageUrl,
          };
        });

        return {
          id: String(comment.CommentID ?? comment.comment_id ?? comment.id ?? `${postId}-${index}`),
          author,
          text,
          time,
          profileImageUrl,
          replyCount: Number(comment.ChildCommentCount ?? comment.child_comment_count ?? replies.length) || replies.length,
          replies,
        };
      });

      setCommentsByPost((current) => ({ ...current, [postId]: normalizedComments }));
    } catch (error) {
      console.error('fetch comments failed', error);
      setCommentsByPost((current) => ({ ...current, [postId]: [] }));
    } finally {
      setCommentLoading((current) => ({ ...current, [postId]: false }));
    }
  };

  const handleOpenComments = async (postId: string) => {
    if (openCommentFor === postId) {
      setOpenCommentFor(null);
      return;
    }

    await fetchCommentsForPost(postId);
    setOpenCommentFor(postId);
  };

  const handleAddComment = async (postId: string, parentCommentId?: string) => {
    const draftKey = parentCommentId ? `${postId}:${parentCommentId}` : postId;
    const commentText = (parentCommentId ? (replyDrafts[draftKey] ?? '') : (commentDrafts[postId] ?? '')).trim();
    if (!commentText) return;

    try {
      const response = await fetch(`${API_BASE_URL}/user/post/${postId}/comment`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          comment_text: commentText,
          ...(parentCommentId ? { parent_comment_id: Number(parentCommentId) } : {}),
        }),
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload.message || payload.error || 'Unable to add your comment.');
      }

      setPosts((currentPosts) =>
        currentPosts.map((post) =>
          post.id === postId
            ? { ...post, comments: post.comments + 1 }
            : post,
        ),
      );

      if (parentCommentId) {
        setReplyDrafts((currentDrafts) => ({ ...currentDrafts, [draftKey]: '' }));
        setReplyingToComment((current) => ({ ...current, [postId]: null }));
      } else {
        setCommentDrafts((currentDrafts) => ({ ...currentDrafts, [postId]: '' }));
      }

      await fetchCommentsForPost(postId);
    } catch (error) {
      console.error('add comment failed', error);
    }
  };

  const handleDeletePost = async (postId: string) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user/post/${postId}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      const payload = await response.json().catch(() => ({}));
      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || payload.error || 'Unable to delete this post.');
      }

      setPosts((currentPosts) => currentPosts.filter((post) => post.id !== postId));
      setPostMenuOpenFor(null);
    } catch (error) {
      console.error('delete post failed', error);
    }
  };

  const handleEditPost = (post: FeedPost) => {
    setEditingPostId(post.id);
    setPostText(post.content || '');
    setSelectedFiles([]);
    setComposerError('');
    setComposerSuccess('');
    setIsComposerOpen(true);
    setPostMenuOpenFor(null);
  };

  const closeComposer = () => {
    setIsComposerOpen(false);
    setEditingPostId(null);
    setPostText('');
    setSelectedFiles([]);
    setComposerError('');
  };

  const handleLogout = async () => {
    try {
      await fetch(`${API_BASE_URL}/user/logout`, {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      console.error('logout failed', error);
    } finally {
      localStorage.removeItem(PROFILE_STORAGE_KEY);
      window.location.href = '/login';
    }
  };

  const handleCreatePost = async () => {
    if (editingPostId) {
      if (!postText.trim()) {
        setComposerError('Please add a caption to update your post.');
        return;
      }
    } else if (!postText.trim() && selectedFiles.length === 0) {
      setComposerError('Please add a caption or at least one media file.');
      return;
    }

    if (selectedFiles.length > 5) {
      setComposerError('You can upload up to 5 files per post.');
      return;
    }

    setIsPosting(true);
    setComposerError('');
    setComposerSuccess('');

    try {
      if (editingPostId) {
        const response = await fetch(`${API_BASE_URL}/user/post/${editingPostId}`, {
          method: 'PATCH',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            caption: postText.trim(),
          }),
        });

        const payload = await response.json().catch(() => ({}));

        if (!response.ok || payload.status_code >= 400) {
          throw new Error(payload.message || payload.error || 'Unable to update your post.');
        }

        setPosts((currentPosts) =>
          currentPosts.map((post) =>
            post.id === editingPostId ? { ...post, content: postText.trim() || post.content } : post,
          ),
        );
        setComposerSuccess('Post updated successfully.');
        closeComposer();
        return;
      }

      const formData = new FormData();
      if (postText.trim()) {
        formData.append('caption', postText.trim());
      }
      selectedFiles.forEach((file) => {
        formData.append('media', file);
      });

      const response = await fetch(`${API_BASE_URL}/user/post`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      const payload = await response.json();

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || payload.error || 'Unable to create your post.');
      }

      const createdPostId = getPostIdFromResponse(payload) ?? `local-${Date.now()}`;

      const createdPost: FeedPost = {
        id: createdPostId,
        author: currentUserName,
        handle: currentUserHandle || '@user',
        time: formatFriendlyTime(new Date().toISOString()),
        content: postText.trim(),
        imageUrls: selectedFiles.length > 0 ? selectedFiles.map((file) => URL.createObjectURL(file)) : [],
        profileImageUrl: currentProfileImageUrl || undefined,
        isLiked: false,
        likes: 0,
        comments: 0,
        shares: 0,
        accent: 'from-sky-400 to-blue-600',
        tag: 'New',
      };

      setPosts((currentPosts) => [createdPost, ...currentPosts]);
      setPostText('');
      setSelectedFiles([]);
      setComposerSuccess('Post created successfully.');
      closeComposer();
    } catch (postError) {
      const message =
        postError instanceof Error ? postError.message : 'Unable to create your post right now.';
      setComposerError(message);
    } finally {
      setIsPosting(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#f3f7fb] text-slate-800">
      <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <header className="mb-6 rounded-[28px] border border-sky-100 bg-white/90 p-4 shadow-[0_20px_50px_rgba(14,116,144,0.08)] backdrop-blur-xl">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={() => navigate('/newsfeed')}
                className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-500 via-cyan-500 to-blue-600 shadow-lg shadow-sky-300/40 transition hover:scale-[1.02]"
                aria-label="Go to newsfeed"
              >
                <svg viewBox="0 0 64 64" className="h-8 w-8 drop-shadow-sm" aria-hidden="true">
                  <defs>
                    <linearGradient id="chattr-logo-gradient" x1="0%" x2="100%" y1="0%" y2="100%">
                      <stop offset="0%" stopColor="#ffffff" />
                      <stop offset="100%" stopColor="#dff4ff" />
                    </linearGradient>
                  </defs>
                  <path
                    d="M18 18.5c0-5.2 4.3-9.5 9.5-9.5h8.8c5.2 0 9.5 4.3 9.5 9.5v15.4c0 5.2-4.3 9.5-9.5 9.5h-7.1L20.6 51c-1.1 1.3-3.2.4-3.2-1.5V34c0-5.2 4.3-9.5 9.5-9.5h-8.9Z"
                    fill="url(#chattr-logo-gradient)"
                    opacity="0.96"
                  />
                  <circle cx="25.5" cy="27.3" r="2.2" fill="#0ea5e9" />
                  <circle cx="32.1" cy="27.3" r="2.2" fill="#0ea5e9" />
                  <circle cx="38.7" cy="27.3" r="2.2" fill="#0ea5e9" />
                  <path d="M22.5 35.5c2.2 2 5 3.1 9.3 3.1 4.7 0 7.7-1.2 9.8-3.2" stroke="#0ea5e9" strokeWidth="2.2" strokeLinecap="round" fill="none"/>
                </svg>
              </button>
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-[0.22em] text-sky-600">Chattr</p>
                <h1 className="text-2xl font-black tracking-tight text-slate-900">News feed</h1>
              </div>
            </div>

            <div className="flex w-full flex-1 items-center justify-end gap-3">
              <div className="relative hidden flex-1 md:block">
                <input
                  type="text"
                  value={userSearch}
                  onChange={(event) => setUserSearch(event.target.value)}
                  placeholder="Search creators, topics, and communities…"
                  className="w-full rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-300 focus:outline-none focus:ring-2 focus:ring-sky-200"
                />
                {userSearch.trim() && (
                  <div className="absolute left-0 right-0 top-[calc(100%+8px)] z-20 max-h-72 overflow-y-auto rounded-2xl border border-sky-100 bg-white shadow-[0_18px_50px_rgba(14,116,144,0.12)]">
                    {isSearchingUsers ? (
                      <div className="px-3 py-2 text-sm text-slate-500">Searching users…</div>
                    ) : userSearchResults.length > 0 ? (
                      userSearchResults.map((user) => (
                        <button
                          key={user.id}
                          type="button"
                          onClick={() => handleUserSearchSelect(user.id)}
                          className="flex w-full cursor-pointer items-center gap-3 border-b border-slate-100 px-3 py-2.5 text-left transition hover:bg-sky-50 last:border-b-0"
                        >
                          <div className="flex h-9 w-9 items-center justify-center overflow-hidden rounded-xl bg-gradient-to-br from-sky-400 to-indigo-500 text-[10px] font-bold text-white">
                            {user.profileImageUrl ? (
                              <img src={user.profileImageUrl} alt={user.name} className="h-full w-full object-cover" />
                            ) : (
                              <span>{getInitials(user.name || user.username)}</span>
                            )}
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="truncate text-sm font-semibold text-slate-800">{user.name}</div>
                            <div className="truncate text-xs text-slate-500">{user.username}</div>
                          </div>
                        </button>
                      ))
                    ) : (
                      <div className="px-3 py-2 text-sm text-slate-500">No users found.</div>
                    )}
                  </div>
                )}
              </div>
              <button
                onClick={() => {
                  setEditingPostId(null);
                  setPostText('');
                  setSelectedFiles([]);
                  setComposerError('');
                  setComposerSuccess('');
                  setIsComposerOpen(true);
                }}
                className="rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-5 py-3 text-sm font-semibold text-white shadow-lg shadow-sky-300/40"
              >
                Create Post
              </button>
              <button
                type="button"
                onClick={handleLogout}
                className="rounded-2xl border border-slate-200 bg-white px-5 py-3 text-sm font-semibold text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
              >
                Logout
              </button>
            </div>
          </div>
        </header>

        <div className="grid gap-6 lg:grid-cols-[260px_minmax(0,1fr)_320px]">
          <aside className="space-y-6">
            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <button
                type="button"
                onClick={() => navigate('/profile')}
                className="mb-4 flex w-full cursor-pointer items-center gap-3 rounded-2xl text-left transition hover:bg-sky-50/70"
              >
                <div className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-400 to-indigo-500 text-sm font-bold text-white">
                  {currentProfileImageUrl ? (
                    <img src={currentProfileImageUrl} alt={currentUserName} className="h-full w-full object-cover" />
                  ) : (
                    <span className="flex h-full w-full items-center justify-center bg-gradient-to-br from-sky-400 to-indigo-500">
                      {getUsernameInitial(currentUserHandle || currentUserName)}
                    </span>
                  )}
                </div>
                <div>
                  <p className="font-bold text-slate-900">{currentUserName}</p>
                  <p className="text-sm text-slate-500">{currentUserHandle}</p>
                </div>
              </button>

              <div className="space-y-3">
                {quickStats.map((stat) => (
                  <div key={stat.label} className="flex items-center justify-between rounded-2xl bg-sky-50 px-3 py-2 text-sm">
                    <span className="text-slate-500">{stat.label}</span>
                    <span className="font-bold text-slate-900">{stat.value}</span>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <p className="mb-4 text-sm font-bold uppercase tracking-[0.18em] text-slate-500">Trending</p>
              <div className="space-y-3 text-sm text-slate-600">
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#productDesign</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#remoteWork</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#creatorEconomy</div>
                <div className="rounded-2xl bg-slate-50 px-3 py-2">#community</div>
              </div>
            </div>
          </aside>

          <main className="space-y-5">
            <div className="rounded-[28px] border border-sky-100 bg-white p-4 shadow-sm">
              <button
                type="button"
                onClick={() => {
                  setEditingPostId(null);
                  setPostText('');
                  setSelectedFiles([]);
                  setComposerError('');
                  setComposerSuccess('');
                  setIsComposerOpen(true);
                }}
                className="flex w-full items-center gap-3 text-left"
              >
                <div className="flex-1 rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-slate-500 transition hover:border-sky-200 hover:bg-sky-100">
                  Share something with your community...
                </div>
              </button>
            </div>

            {error && (
              <div className="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
                {error}
              </div>
            )}

            {isLoading ? (
              <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
                Loading feed...
              </div>
            ) : (
              posts.map((post) => (
                <article key={post.id} className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
                  <div className="mb-4 flex items-start justify-between gap-3">
                    <div className="flex items-start gap-3">
                      <button
                        type="button"
                        onClick={() => goToProfile(post.userId)}
                        className="flex h-12 w-12 cursor-pointer items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br text-sm font-bold text-white shadow-sm ring-1 ring-sky-100 transition hover:scale-[1.02]"
                        aria-label={`View ${post.author}'s profile`}
                        style={{ backgroundImage: post.profileImageUrl ? 'none' : undefined }}
                      >
                        {post.profileImageUrl ? (
                          <img src={post.profileImageUrl} alt={post.author} className="h-full w-full object-cover" />
                        ) : (
                          <span className={`flex h-full w-full items-center justify-center bg-gradient-to-br ${post.accent}`}>
                            {getInitials(post.author)}
                          </span>
                        )}
                      </button>

                      <div>
                        <div className="flex items-center gap-2">
                          <button
                            type="button"
                            onClick={() => goToProfile(post.userId)}
                            className="cursor-pointer text-left font-bold text-slate-900 transition hover:text-sky-700"
                            aria-label={`View ${post.author}'s profile`}
                          >
                            {post.author}
                          </button>
                          <button
                            type="button"
                            onClick={() => goToProfile(post.userId)}
                            className="cursor-pointer text-left text-sm text-slate-500 transition hover:text-sky-700"
                            aria-label={`View ${post.author}'s profile`}
                          >
                            {post.handle}
                          </button>
                        </div>
                        <div className="mt-1 flex items-center gap-2 text-xs text-slate-500">
                          <span>{post.time}</span>
                          <span>•</span>
                          <span className="rounded-full bg-sky-50 px-2 py-0.5 font-medium text-sky-700">{post.tag}</span>
                        </div>
                      </div>
                    </div>

                    {(() => {
                      const normalizedPostUserId = String(post.userId ?? '').trim();
                      const normalizedCurrentHandle = currentUserHandle.toLowerCase();
                      const normalizedPostHandle = post.handle.toLowerCase();
                      const nestedUserIdFromPayload = String((post as FeedPost & { UserData?: { UserID?: string | number } }).UserData?.UserID ?? '').trim();

                      const isOwnPost =
                        !!currentUserId &&
                        (normalizedPostUserId === currentUserId ||
                          nestedUserIdFromPayload === currentUserId ||
                          (!normalizedPostUserId &&
                            (post.author === currentUserName ||
                              normalizedPostHandle === normalizedCurrentHandle ||
                              post.profileImageUrl === currentProfileImageUrl)));

                      if (!isOwnPost) return null;

                      return (
                        <div className="relative">
                          <button
                            type="button"
                            onClick={() => setPostMenuOpenFor((current) => (current === post.id ? null : post.id))}
                            className="rounded-full border border-slate-200 px-2 py-1 text-xs font-medium text-slate-500 transition hover:border-sky-200 hover:text-sky-700"
                            aria-label={`More options for ${post.author}'s post`}
                          >
                            •••
                          </button>

                          {postMenuOpenFor === post.id && (
                            <div className="absolute right-0 top-full z-20 mt-2 min-w-[120px] overflow-hidden rounded-2xl border border-sky-100 bg-white shadow-lg">
                              <button
                                type="button"
                                onClick={() => handleEditPost(post)}
                                className="block w-full px-3 py-2 text-left text-sm text-slate-700 transition hover:bg-sky-50 hover:text-sky-700"
                              >
                                Edit
                              </button>
                              <button
                                type="button"
                                onClick={() => handleDeletePost(post.id)}
                                className="block w-full px-3 py-2 text-left text-sm text-red-600 transition hover:bg-red-50"
                              >
                                Delete
                              </button>
                            </div>
                          )}
                        </div>
                      );
                    })()}
                  </div>

                  <p className="text-[15px] leading-7 text-slate-700">{post.content}</p>

                  {post.imageUrls.length > 0 && (
                    <div className={`mt-4 grid gap-2 ${post.imageUrls.length > 1 ? 'grid-cols-2' : 'grid-cols-1'}`}>
                      {post.imageUrls.slice(0, 4).map((imageUrl, index) => (
                        <img
                          key={`${post.id}-image-${index}`}
                          src={imageUrl}
                          alt={`${post.author} post ${index + 1}`}
                          className={`w-full rounded-2xl border border-sky-100 object-cover ${post.imageUrls.length === 1 ? 'max-h-[420px]' : 'h-52'} `}
                          onError={(event) => {
                            const target = event.currentTarget as HTMLImageElement;
                            target.style.display = 'none';
                          }}
                        />
                      ))}
                    </div>
                  )}

                  <div className="mt-5 flex items-center justify-between border-t border-slate-100 pt-4 text-sm text-slate-500">
                    <div className="flex items-center gap-2 sm:gap-5">
                      <button
                        type="button"
                        onClick={() => handleToggleLike(post.id, post.isLiked)}
                        className={`flex cursor-pointer items-center gap-2 rounded-full px-3 py-1.5 transition ${post.isLiked ? 'bg-rose-50 text-rose-600' : 'bg-slate-50 text-slate-600'} `}
                      >
                        <span>{post.isLiked ? '♥' : '♡'}</span>
                        <span>{formatCount(post.likes)}</span>
                      </button>
                      <button
                        type="button"
                        onClick={() => handleOpenComments(post.id)}
                        className="flex cursor-pointer items-center gap-2 rounded-full bg-slate-50 px-3 py-1.5 text-slate-600 transition hover:bg-sky-50 hover:text-sky-700"
                      >
                        <span>💬</span>
                        <span>{formatCount(post.comments)}</span>
                      </button>
                      <span className="flex items-center gap-2 rounded-full bg-slate-50 px-3 py-1.5">
                        <span>↻</span>
                        <span>{formatCount(post.shares)}</span>
                      </span>
                    </div>
                    <button className="rounded-full bg-sky-50 px-3 py-1.5 font-medium text-sky-700">View post</button>
                  </div>

                  {openCommentFor === post.id && (
                    <div className="mt-4 rounded-2xl border border-sky-100 bg-sky-50/60 p-3">
                      <div className="mb-3 rounded-2xl border border-sky-200 bg-white p-3">
                        <textarea
                          value={commentDrafts[post.id] ?? ''}
                          onChange={(event) =>
                            setCommentDrafts((currentDrafts) => ({
                              ...currentDrafts,
                              [post.id]: event.target.value,
                            }))
                          }
                          rows={3}
                          placeholder="Write a comment..."
                          className="w-full resize-none border-0 bg-transparent text-sm text-slate-700 placeholder:text-slate-400 focus:outline-none"
                        />
                        <div className="mt-2 flex items-center justify-end gap-2">
                          <button
                            type="button"
                            onClick={() => setOpenCommentFor(null)}
                            className="cursor-pointer rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-500"
                          >
                            Close
                          </button>
                          <button
                            type="button"
                            onClick={() => handleAddComment(post.id)}
                            className="cursor-pointer rounded-full bg-sky-600 px-3 py-1.5 text-sm font-semibold text-white"
                          >
                            Post comment
                          </button>
                        </div>
                      </div>

                      {commentLoading[post.id] ? (
                        <div className="py-2 text-sm text-slate-500">Loading comments...</div>
                      ) : (commentsByPost[post.id] ?? []).length > 0 ? (
                        <div className="space-y-3">
                          {(commentsByPost[post.id] ?? []).map((comment) => {
                            const replyDraftKey = `${post.id}:${comment.id}`;
                            const isReplying = replyingToComment[post.id] === comment.id;
                            const isExpanded = expandedRepliesByComment[`${post.id}:${comment.id}`] ?? false;

                            return (
                              <div key={comment.id} className="rounded-2xl bg-white px-3 py-2 shadow-sm ring-1 ring-slate-100">
                                <div className="mb-2 flex items-start gap-3">
                                  <div className="flex h-9 w-9 items-center justify-center overflow-hidden rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[10px] font-bold text-white ring-1 ring-sky-100">
                                    {comment.profileImageUrl ? (
                                      <img src={comment.profileImageUrl} alt={comment.author} className="h-full w-full object-cover" />
                                    ) : (
                                      <span>{getInitials(comment.author)}</span>
                                    )}
                                  </div>
                                  <div className="min-w-0 flex-1">
                                    <div className="mb-1 flex items-center justify-between gap-3">
                                      <span className="text-sm font-semibold text-slate-800">{comment.author}</span>
                                      <span className="text-[11px] text-slate-400">{comment.time}</span>
                                    </div>
                                    <p className="text-sm leading-6 text-slate-600">{comment.text}</p>
                                  </div>
                                </div>

                                {comment.replyCount && comment.replyCount > 0 ? (
                                  <button
                                    type="button"
                                    onClick={() =>
                                      setExpandedRepliesByComment((current) => ({
                                        ...current,
                                        [`${post.id}:${comment.id}`]: !isExpanded,
                                      }))
                                    }
                                    className="mt-2 cursor-pointer text-xs font-medium text-sky-700 hover:text-sky-800"
                                  >
                                    {isExpanded ? 'Hide ' : ''}
                                    {comment.replyCount} {comment.replyCount === 1 ? 'reply' : 'replies'}
                                  </button>
                                ) : null}

                                <div className="mt-2 flex items-center gap-2">
                                  <button
                                    type="button"
                                    onClick={() =>
                                      setReplyingToComment((current) => ({
                                        ...current,
                                        [post.id]: isReplying ? null : comment.id,
                                      }))
                                    }
                                    className="cursor-pointer text-xs font-medium text-sky-700 hover:text-sky-800"
                                  >
                                    {isReplying ? 'Cancel' : 'Reply'}
                                  </button>
                                </div>

                                {isReplying && (
                                  <div className="mt-3 rounded-2xl border border-sky-200 bg-sky-50 p-2">
                                    <textarea
                                      value={replyDrafts[replyDraftKey] ?? ''}
                                      onChange={(event) =>
                                        setReplyDrafts((currentDrafts) => ({
                                          ...currentDrafts,
                                          [replyDraftKey]: event.target.value,
                                        }))
                                      }
                                      rows={2}
                                      placeholder={`Reply to ${comment.author}...`}
                                      className="w-full resize-none border-0 bg-transparent text-sm text-slate-700 placeholder:text-slate-400 focus:outline-none"
                                    />
                                    <div className="mt-2 flex items-center justify-end gap-2">
                                      <button
                                        type="button"
                                        onClick={() => setReplyingToComment((current) => ({ ...current, [post.id]: null }))}
                                        className="cursor-pointer rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-500"
                                      >
                                        Close
                                      </button>
                                      <button
                                        type="button"
                                        onClick={() => handleAddComment(post.id, comment.id)}
                                        className="cursor-pointer rounded-full bg-sky-600 px-3 py-1.5 text-xs font-semibold text-white"
                                      >
                                        Post reply
                                      </button>
                                    </div>
                                  </div>
                                )}

                                {isExpanded && (comment.replies ?? []).length > 0 && (
                                  <div className="mt-3 space-y-2 border-l border-sky-200 pl-3">
                                    {(comment.replies ?? []).map((reply) => (
                                      <div key={reply.id} className="rounded-xl bg-sky-50 px-2.5 py-2">
                                        <div className="flex items-start gap-2">
                                          <div className="flex h-7 w-7 items-center justify-center overflow-hidden rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[9px] font-bold text-white ring-1 ring-sky-100">
                                            {reply.profileImageUrl ? (
                                              <img src={reply.profileImageUrl} alt={reply.author} className="h-full w-full object-cover" />
                                            ) : (
                                              <span>{getInitials(reply.author)}</span>
                                            )}
                                          </div>
                                          <div className="min-w-0 flex-1">
                                            <div className="mb-1 flex items-center justify-between gap-3">
                                              <span className="text-xs font-semibold text-slate-800">{reply.author}</span>
                                              <span className="text-[10px] text-slate-400">{reply.time}</span>
                                            </div>
                                            <p className="text-xs leading-5 text-slate-600">{reply.text}</p>
                                          </div>
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                )}
                              </div>
                            );
                          })}
                        </div>
                      ) : (
                        <div className="rounded-2xl bg-white px-3 py-2 text-sm text-slate-500 ring-1 ring-slate-100">
                          No comments yet. Be the first to share your thought.
                        </div>
                      )}
                    </div>
                  )}
                </article>
              ))
            )}
          </main>

          <aside className="space-y-6">
            <div className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
              <p className="mb-4 text-sm font-bold uppercase tracking-[0.18em] text-slate-500">Suggested</p>
              <div className="space-y-4">
                {sidebarPeople.map((person) => (
                  <div key={person.name} className="flex items-center justify-between gap-3 rounded-2xl bg-slate-50 p-3">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${person.accent} text-xs font-bold text-white`}>
                        {person.name
                          .split(' ')
                          .map((part) => part[0])
                          .join('')
                          .slice(0, 2)}
                      </div>
                      <div>
                        <p className="font-semibold text-slate-800">{person.name}</p>
                        <p className="text-xs text-slate-500">{person.role}</p>
                      </div>
                    </div>
                    <button className="rounded-full bg-sky-50 px-3 py-1.5 text-xs font-medium text-sky-700">
                      Follow
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-[28px] border border-sky-100 bg-gradient-to-br from-sky-500 to-blue-600 p-5 text-white shadow-[0_20px_40px_rgba(59,130,246,0.25)]">
              <p className="text-sm uppercase tracking-[0.18em] text-sky-100">Your pulse</p>
              <p className="mt-4 text-4xl font-black">96%</p>
              <p className="mt-2 text-sm text-sky-100">Engagement is trending up after your recent posts.</p>
            </div>
          </aside>
        </div>
      </div>

      {isComposerOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="w-full max-w-xl rounded-[30px] border border-sky-100 bg-white p-5 shadow-[0_25px_80px_rgba(15,23,42,0.2)]">
            <div className="mb-4 flex items-center justify-between">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.22em] text-sky-600">
                  {editingPostId ? 'Edit post' : 'Create post'}
                </p>
                <h2 className="mt-2 text-2xl font-black text-slate-900">
                  {editingPostId ? 'Update your post' : 'Share an update'}
                </h2>
              </div>
              <button
                type="button"
                onClick={closeComposer}
                className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-500"
              >
                Close
              </button>
            </div>

            <div className="space-y-4">
              <textarea
                value={postText}
                onChange={(e) => setPostText(e.target.value)}
                rows={5}
                placeholder={editingPostId ? 'Edit your caption...' : 'What would you like to share with your community?'}
                className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
              />

              {!editingPostId && (
                <div>
                  <label className="mb-2 block text-sm font-semibold text-slate-700">Add media</label>
                  <input
                    type="file"
                    accept="image/*,video/*,audio/*"
                    multiple
                    onChange={(event) => setSelectedFiles(Array.from(event.target.files ?? []))}
                    className="block w-full rounded-2xl border border-dashed border-sky-200 bg-sky-50 p-3 text-sm text-slate-600 file:mr-3 file:rounded-full file:border-0 file:bg-sky-100 file:px-3 file:py-2 file:text-sm file:font-semibold file:text-sky-700"
                  />
                  {selectedFiles.length > 0 && (
                    <div className="mt-2 text-xs text-slate-500">
                      {selectedFiles.length} file(s) selected
                    </div>
                  )}
                </div>
              )}

              {composerError && (
                <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
                  {composerError}
                </div>
              )}

              {composerSuccess && (
                <div className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-700">
                  {composerSuccess}
                </div>
              )}

              <button
                type="button"
                disabled={isPosting}
                onClick={handleCreatePost}
                className="flex w-full items-center justify-center rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3.5 text-base font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105 disabled:cursor-not-allowed disabled:opacity-70"
              >
                {isPosting ? (editingPostId ? 'Updating...' : 'Posting...') : editingPostId ? 'Update Post' : 'Create Post'}
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="fixed bottom-6 right-6 z-40 flex flex-col items-end gap-3">
        {isChatOpen && (
          <div className="h-[480px] w-[360px] overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
            <div className="flex items-center justify-between border-b border-slate-100 bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3 text-white">
              <div className="flex items-center gap-3">
                {chatSelectedUser ? (
                  <button
                    type="button"
                    onClick={() => {
                      setChatSelectedUser(null);
                      setChatMessages([]);
                      setChatDraft('');
                      setChatError('');
                      setChatSearch('');
                      setChatSearchResults([]);
                    }}
                    className="rounded-full border border-white/30 px-2 py-1 text-xs font-medium text-white/90"
                  >
                    Back
                  </button>
                ) : (
                  <div className="flex h-9 w-9 items-center justify-center rounded-full border border-white/60 bg-white/15 text-lg">
                    💬
                  </div>
                )}

                {chatSelectedUser ? (
                  <>
                    {chatSelectedUser.profileImageUrl ? (
                      <img
                        src={chatSelectedUser.profileImageUrl}
                        alt={chatSelectedUser.name}
                        className="h-9 w-9 rounded-full border border-white/60 object-cover"
                      />
                    ) : (
                      <div className="flex h-9 w-9 items-center justify-center rounded-full border border-white/60 bg-white/15 text-xs font-bold">
                        {getInitials(chatSelectedUser.name)}
                      </div>
                    )}
                    <div>
                      <p className="text-sm font-bold">{chatSelectedUser.name}</p>
                      <p className="text-[10px] text-sky-100">{chatSelectedUser.username}</p>
                    </div>
                  </>
                ) : (
                  <div>
                    <p className="text-sm font-bold">Messages</p>
                    <p className="text-[10px] text-sky-100">Recent chats</p>
                  </div>
                )}
              </div>

              <button
                type="button"
                onClick={() => setIsChatOpen(false)}
                className="rounded-full border border-white/30 px-2 py-1 text-xs font-medium text-white/90"
              >
                Close
              </button>
            </div>

            {!chatSelectedUser ? (
              <div className="flex h-[calc(480px-64px)] flex-col bg-slate-50">
                <div className="relative border-b border-slate-100 bg-slate-50 p-3">
                  <div className="relative">
                    <input
                      value={chatSearch}
                      onChange={(event) => setChatSearch(event.target.value)}
                      placeholder="Search messages or people"
                      className="w-full rounded-full border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
                    />
                    {isSearchingChatUsers && (
                      <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-medium uppercase tracking-[0.18em] text-sky-600">
                        Searching
                      </span>
                    )}
                  </div>

                  {chatSearch && chatSearchResults.length > 0 && (
                    <div className="absolute left-3 right-3 top-[calc(100%+6px)] z-20 max-h-52 overflow-y-auto rounded-2xl border border-slate-200 bg-white p-2 shadow-lg">
                      {chatSearchResults.map((user) => (
                        <button
                          key={user.id}
                          type="button"
                          onClick={() => {
                            setChatSearch('');
                            setChatSearchResults([]);
                            void openChatSession({
                              id: user.id,
                              name: user.name,
                              username: user.username,
                              profileImageUrl: user.profileImageUrl,
                              conversationId: '',
                            });
                          }}
                          className="flex w-full items-center gap-3 rounded-xl px-2 py-2 text-left transition hover:bg-slate-50"
                        >
                          {user.profileImageUrl ? (
                            <img src={user.profileImageUrl} alt={user.name} className="h-9 w-9 rounded-full object-cover" />
                          ) : (
                            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-sky-100 text-xs font-bold text-sky-700">
                              {getInitials(user.name)}
                            </div>
                          )}
                          <div className="min-w-0 flex-1">
                            <p className="truncate text-sm font-semibold text-slate-800">{user.name}</p>
                            <p className="truncate text-xs text-slate-500">{user.username}</p>
                          </div>
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                <div className="flex-1 overflow-y-auto px-2 py-2">
                  {chatSessions.length === 0 ? (
                    <div className="flex h-full items-center justify-center px-4 text-center text-sm text-slate-500">
                      No recent chats yet. Search for someone to start one.
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {chatSessions
                        .filter((session) => {
                          const query = chatSearch.trim().toLowerCase();
                          if (!query) return true;
                          return (
                            session.name.toLowerCase().includes(query) ||
                            session.username.toLowerCase().includes(query.replace(/^@/, '')) ||
                            session.lastMessage.toLowerCase().includes(query)
                          );
                        })
                        .map((session) => (
                          <button
                            key={session.conversationId || session.id}
                            type="button"
                            onClick={() => void openChatSession(session)}
                            className="flex w-full items-center gap-3 rounded-2xl bg-white p-3 text-left shadow-sm ring-1 ring-slate-100 transition hover:bg-slate-50"
                          >
                            {session.profileImageUrl ? (
                              <img src={session.profileImageUrl} alt={session.name} className="h-12 w-12 rounded-full object-cover" />
                            ) : (
                              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                                {getInitials(session.name)}
                              </div>
                            )}
                            <div className="min-w-0 flex-1">
                              <div className="mb-1 flex items-center justify-between gap-2">
                                <p className="truncate text-sm font-semibold text-slate-800">{session.name}</p>
                                <span className="text-[10px] text-slate-400">{session.lastMessageTime}</span>
                              </div>
                              <p className="truncate text-xs text-slate-500">{session.username}</p>
                              <p className="mt-1 truncate text-xs text-slate-600">{session.lastMessage || 'Start a conversation'}</p>
                            </div>
                          </button>
                        ))}
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="flex h-[calc(480px-64px)] flex-col bg-slate-50">
                <div className="flex-1 space-y-2 overflow-y-auto px-3 py-3">
                  {isLoadingChatHistory ? (
                    <div className="flex h-full items-center justify-center text-sm text-slate-500">Loading conversation...</div>
                  ) : chatMessages.length > 0 ? (
                    chatMessages.map((message) => (
                      <div
                        key={message.id}
                        className={`flex ${message.mine ? 'justify-end' : 'justify-start'}`}
                      >
                        <div
                          className={`max-w-[80%] rounded-2xl px-3 py-2 text-sm ${
                            message.mine
                              ? 'bg-gradient-to-r from-sky-500 to-blue-600 text-white'
                              : 'bg-white text-slate-700 ring-1 ring-slate-200'
                          }`}
                        >
                          <div>{message.content}</div>
                          <div
                            className={`mt-1 text-[10px] ${
                              message.mine ? 'text-sky-100' : 'text-slate-400'
                            }`}
                          >
                            {formatChatMessageTime(message.createdAt)}
                          </div>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="flex h-full items-center justify-center text-center text-sm text-slate-500">
                      Start the conversation with {chatSelectedUser.name}
                    </div>
                  )}
                </div>

                <div className="border-t border-slate-200 bg-white p-3">
                  {chatError && (
                    <div className="mb-2 rounded-xl border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-600">
                      {chatError}
                    </div>
                  )}
                  <div className="flex items-center gap-2">
                    <input
                      value={chatDraft}
                      onChange={(event) => setChatDraft(event.target.value)}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter' && !event.shiftKey) {
                          event.preventDefault();
                          handleSendChatMessage();
                        }
                      }}
                      placeholder="Type a message"
                      className="flex-1 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
                    />
                    <button
                      type="button"
                      onClick={handleSendChatMessage}
                      className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white"
                    >
                      Send
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        <button
          type="button"
          onClick={() => setIsChatOpen((current) => !current)}
          className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-r from-sky-500 to-blue-600 text-2xl text-white shadow-[0_18px_40px_rgba(59,130,246,0.45)] transition hover:scale-[1.02]"
          aria-label="Open messages"
        >
          💬
        </button>
      </div>
    </div>
  );
}
