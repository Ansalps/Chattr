import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

type FeedComment = {
  id: string;
  author: string;
  text: string;
  time: string;
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

const PROFILE_STORAGE_KEY = 'chattr_profile_info';

const readStoredProfile = () => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) return { name: '', username: '', profileImageUrl: '' };

    const parsed = JSON.parse(raw) as Partial<{ name?: string; username?: string; profileImageUrl?: string }>;
    return {
      name: typeof parsed.name === 'string' ? parsed.name : '',
      username: typeof parsed.username === 'string' ? parsed.username : '',
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : '',
    };
  } catch (error) {
    return { name: '', username: '', profileImageUrl: '' };
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

      const author =
        (typeof source.author === 'string' && source.author) ||
        (typeof source.name === 'string' && source.name) ||
        (typeof source.UserName === 'string' && source.UserName) ||
        (typeof source.username === 'string' && source.username) ||
        (typeof userMeta.name === 'string' && userMeta.name) ||
        (typeof userMeta.UserName === 'string' && userMeta.UserName) ||
        (typeof userMeta.username === 'string' && userMeta.username) ||
        'Chattr User';

      const profileImageUrl =
        (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
        (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
        (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
        (typeof userMeta.profile_image_url === 'string' && userMeta.profile_image_url) ||
        (typeof userMeta.ProfileImgUrl === 'string' && userMeta.ProfileImgUrl) ||
        '';

      const handle =
        (typeof source.handle === 'string' && source.handle) ||
        (typeof source.Handle === 'string' && source.Handle) ||
        (typeof userMeta.username === 'string' && `@${userMeta.username}`) ||
        (typeof userMeta.UserName === 'string' && `@${userMeta.UserName}`) ||
        `@${author.toLowerCase().replace(/\s+/g, '') || 'user'}`;

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
          (userMeta.user_id as string | number | undefined) ??
          (userMeta.UserID as string | number | undefined) ??
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

  const storedProfile = readStoredProfile();
  const currentUserName = storedProfile.name || 'Jane Doe';
  const currentUserHandle = storedProfile.username || '@janedoe';
  const currentProfileImageUrl = storedProfile.profileImageUrl || '';

  useEffect(() => {
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

    fetchNewsFeed();
  }, []);

  const quickStats = useMemo(
    () => [
      { label: 'Followers', value: '18.4K' },
      { label: 'Following', value: '612' },
      { label: 'Posts', value: '1.2K' },
      { label: 'Reach', value: '92.6K' },
    ],
    [],
  );

  const sidebarPeople = [
    { name: 'Rina Shah', role: 'Product Lead', accent: 'from-pink-400 to-rose-500' },
    { name: 'Noah Kim', role: 'UX Strategist', accent: 'from-sky-400 to-cyan-500' },
    { name: 'Aisha Noor', role: 'Community Builder', accent: 'from-violet-400 to-purple-500' },
  ];

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

          const childTime =
            (typeof child.CommentAge === 'string' && child.CommentAge) ||
            (typeof child.comment_age === 'string' && child.comment_age) ||
            'just now';

          return {
            id: String(child.CommentID ?? child.comment_id ?? child.id ?? `${postId}-${index}-${childIndex}`),
            author: childAuthor,
            text: childText,
            time: childTime,
          };
        });

        return {
          id: String(comment.CommentID ?? comment.comment_id ?? comment.id ?? `${postId}-${index}`),
          author,
          text,
          time,
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

  const handleLogout = async () => {
    try {
      await fetch(`${API_BASE_URL}/user/logout`, {
        method: 'POST',
        credentials: 'include',
      });
    } catch (error) {
      console.error('logout failed', error);
    } finally {
      window.location.href = '/login';
    }
  };

  const handleCreatePost = async () => {
    if (!postText.trim() && selectedFiles.length === 0) {
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
      setIsComposerOpen(false);
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
                onClick={() => navigate('/profile')}
                className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 text-lg font-black text-white shadow-lg shadow-sky-300/40 transition hover:scale-[1.02]"
                aria-label="View profile"
              >
                {currentProfileImageUrl ? (
                  <img src={currentProfileImageUrl} alt={currentUserName} className="h-full w-full object-cover" />
                ) : (
                  <span className="flex h-full w-full items-center justify-center bg-gradient-to-br from-sky-500 to-blue-600">
                    {getUsernameInitial(currentUserHandle || currentUserName)}
                  </span>
                )}
              </button>
              <div>
                <p className="text-[10px] font-semibold uppercase tracking-[0.22em] text-sky-600">Chattr</p>
                <h1 className="text-2xl font-black tracking-tight text-slate-900">News feed</h1>
              </div>
            </div>

            <div className="flex flex-1 items-center justify-end gap-3">
              <div className="hidden w-full max-w-md rounded-2xl border border-sky-100 bg-sky-50 px-4 py-3 text-sm text-slate-500 md:block">
                Search creators, topics, and communities…
              </div>
              <button
                onClick={() => setIsComposerOpen(true)}
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
              <div className="mb-4 flex items-center gap-3">
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
              </div>

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
                onClick={() => setIsComposerOpen(true)}
                className="flex w-full items-center gap-3 text-left"
              >
                <div className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                  {currentProfileImageUrl ? (
                    <img src={currentProfileImageUrl} alt={currentUserName} className="h-full w-full object-cover" />
                  ) : (
                    <span className="flex h-full w-full items-center justify-center bg-gradient-to-br from-sky-400 to-blue-600">
                      {getUsernameInitial(currentUserHandle || currentUserName)}
                    </span>
                  )}
                </div>
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
                        onClick={() => navigate(`/profile/${encodeURIComponent(post.userId ?? 'me')}`)}
                        className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br text-sm font-bold text-white shadow-sm ring-1 ring-sky-100 transition hover:scale-[1.02]"
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
                          <p className="font-bold text-slate-900">{post.author}</p>
                          <span className="text-sm text-slate-500">{post.handle}</span>
                        </div>
                        <div className="mt-1 flex items-center gap-2 text-xs text-slate-500">
                          <span>{post.time}</span>
                          <span>•</span>
                          <span className="rounded-full bg-sky-50 px-2 py-0.5 font-medium text-sky-700">{post.tag}</span>
                        </div>
                      </div>
                    </div>

                    <button className="rounded-full border border-slate-200 px-2 py-1 text-xs font-medium text-slate-500">
                      •••
                    </button>
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
                                <div className="mb-1 flex items-center justify-between gap-3">
                                  <span className="text-sm font-semibold text-slate-800">{comment.author}</span>
                                  <span className="text-[11px] text-slate-400">{comment.time}</span>
                                </div>
                                <p className="text-sm leading-6 text-slate-600">{comment.text}</p>

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
                                        <div className="mb-1 flex items-center justify-between gap-3">
                                          <span className="text-xs font-semibold text-slate-800">{reply.author}</span>
                                          <span className="text-[10px] text-slate-400">{reply.time}</span>
                                        </div>
                                        <p className="text-xs leading-5 text-slate-600">{reply.text}</p>
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
                <p className="text-xs font-semibold uppercase tracking-[0.22em] text-sky-600">Create post</p>
                <h2 className="mt-2 text-2xl font-black text-slate-900">Share an update</h2>
              </div>
              <button
                type="button"
                onClick={() => setIsComposerOpen(false)}
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
                placeholder="What would you like to share with your community?"
                className="w-full rounded-2xl border border-sky-200 bg-sky-50/60 px-4 py-3 text-base text-slate-900 placeholder:text-slate-400 focus:border-sky-500 focus:outline-none focus:ring-2 focus:ring-sky-200"
              />

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
                {isPosting ? 'Posting...' : 'Create Post'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
