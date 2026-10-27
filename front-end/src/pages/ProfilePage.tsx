import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchProfilePageData } from '../features/profile/profileService';

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

type ProfilePost = {
  id: string;
  content: string;
  imageUrls: string[];
  likes: number;
  comments: number;
  time: string;
  isLiked?: boolean;
};

type CustomLink = {
  id: string;
  label: string;
  url: string;
};

type ProfileUser = {
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

const PROFILE_STORAGE_KEY = 'chattr_profile_info';

const readStoredProfile = () => {
  try {
    const raw = localStorage.getItem(PROFILE_STORAGE_KEY);
    if (!raw) {
      return {} as Partial<ProfileUser>;
    }

    const parsed = JSON.parse(raw) as Partial<ProfileUser> & { links?: Record<string, string>; customLinks?: Array<{ id?: string; label?: string; url?: string }> };
    const customLinks = Array.isArray(parsed.customLinks)
      ? (parsed.customLinks
          .filter((entry) => !!entry && typeof entry === 'object') as Array<Record<string, unknown>>)
          .map((entry, index) => ({
            id: typeof entry.id === 'string' ? entry.id : `link-${index}`,
            label: typeof entry.label === 'string' ? entry.label : 'Link',
            url: typeof entry.url === 'string' ? entry.url : '',
          }))
          .filter((entry) => entry.url.trim())
      : [];

    return {
      name: typeof parsed.name === 'string' ? parsed.name : undefined,
      username: typeof parsed.username === 'string' ? parsed.username : undefined,
      profileImageUrl: typeof parsed.profileImageUrl === 'string' ? parsed.profileImageUrl : undefined,
      bio: typeof parsed.bio === 'string' ? parsed.bio : undefined,
      website: typeof parsed.website === 'string' ? parsed.website : typeof parsed.links?.website === 'string' ? parsed.links.website : undefined,
      linkedin: typeof parsed.linkedin === 'string' ? parsed.linkedin : typeof parsed.links?.linkedin === 'string' ? parsed.links.linkedin : undefined,
      instagram: typeof parsed.instagram === 'string' ? parsed.instagram : typeof parsed.links?.instagram === 'string' ? parsed.links.instagram : undefined,
      customLinks: customLinks.length > 0 ? customLinks : undefined,
    };
  } catch (error) {
    return {} as Partial<ProfileUser>;
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

const normalizeLinkUrl = (value: string) => value.trim().replace(/\/+$/, '').toLowerCase();

export default function ProfilePage() {
  const navigate = useNavigate();
  const { userId } = useParams();
  const [posts, setPosts] = useState<ProfilePost[]>([]);
  const [user, setUser] = useState<ProfileUser>({
    id: userId ?? 'me',
    name: 'User Profile',
    username: '@user',
  });
  const [isLoading, setIsLoading] = useState(true);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [followerCount, setFollowerCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);
  const [isFollowing, setIsFollowing] = useState(false);
  const [isUpdatingFollowState, setIsUpdatingFollowState] = useState(false);
  const [error, setError] = useState('');
  const avatarInputRef = useRef<HTMLInputElement | null>(null);
  const isOwnProfile = !userId || userId === 'me';
  const [selectedPost, setSelectedPost] = useState<ProfilePost | null>(null);
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [showComments, setShowComments] = useState(false);
  const [commentDrafts, setCommentDrafts] = useState<Record<string, string>>({});
  const [replyDrafts, setReplyDrafts] = useState<Record<string, string>>({});
  const [replyingToComment, setReplyingToComment] = useState<Record<string, string | null>>({});
  const [expandedRepliesByComment, setExpandedRepliesByComment] = useState<Record<string, boolean>>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<string, FeedComment[]>>({});
  const [commentLoading, setCommentLoading] = useState<Record<string, boolean>>({});

  useEffect(() => {
    const targetUserId = userId;

    const fetchProfile = async () => {
      try {
        const nextProfileData = await fetchProfilePageData(targetUserId);
        setPosts(nextProfileData.posts);
        setFollowerCount(nextProfileData.followerCount);
        setFollowingCount(nextProfileData.followingCount);
        setIsFollowing(nextProfileData.isFollowing);
        setUser(nextProfileData.user);
      } catch (fetchError) {
        const message = fetchError instanceof Error ? fetchError.message : 'Unable to load this profile.';
        setError(message);
      } finally {
        setIsLoading(false);
      }
    };

    fetchProfile();
  }, [userId]);

  const handleFollowToggle = async () => {
    if (!userId || userId === 'me') return;

    setIsUpdatingFollowState(true);
    setError('');

    try {
      const method = isFollowing ? 'DELETE' : 'POST';
      const endpoint = `${API_BASE_URL}/user/relation/${isFollowing ? 'unfollow' : 'follow'}/${encodeURIComponent(userId)}`;
      const response = await fetch(endpoint, {
        method,
        credentials: 'include',
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || payload.error || `Unable to ${isFollowing ? 'unfollow' : 'follow'} this user.`);
      }

      setIsFollowing((current) => !current);
      setFollowerCount((current) => Math.max(0, current + (isFollowing ? -1 : 1)));
    } catch (followError) {
      const message = followError instanceof Error ? followError.message : 'Unable to update follow status.';
      setError(message);
    } finally {
      setIsUpdatingFollowState(false);
    }
  };

  const handleAvatarUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('image', file);

    setIsUploadingAvatar(true);
    setError('');

    try {
      const response = await fetch(`${API_BASE_URL}/user/set-profile-image`, {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok || payload.status_code >= 400) {
        throw new Error(payload.message || payload.error || 'Unable to upload profile image.');
      }

      const uploadedUrl =
        (typeof payload.ImageUrl === 'string' && payload.ImageUrl) ||
        (typeof payload.image_url === 'string' && payload.image_url) ||
        (typeof payload.imageUrl === 'string' && payload.imageUrl) ||
        '';

      if (!uploadedUrl) {
        throw new Error('The server did not return an uploaded image URL.');
      }

      const updatedProfile = {
        ...readStoredProfile(),
        name: user.name,
        username: user.username,
        profileImageUrl: uploadedUrl,
      };

      localStorage.setItem(PROFILE_STORAGE_KEY, JSON.stringify(updatedProfile));
      setUser((current) => ({ ...current, profileImageUrl: uploadedUrl }));
    } catch (uploadError) {
      const message = uploadError instanceof Error ? uploadError.message : 'Unable to upload profile image.';
      setError(message);
    } finally {
      setIsUploadingAvatar(false);
      event.target.value = '';
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

  const handleToggleLike = async (postId: string, currentLiked: boolean) => {
    try {
      const response = await fetch(`${API_BASE_URL}/user/post/${postId}/${currentLiked ? 'unlike' : 'like'}`, {
        method: currentLiked ? 'DELETE' : 'POST',
        credentials: 'include',
      });

      if (!response.ok) {
        const payload = await response.json().catch(() => ({}));
        throw new Error(payload.message || payload.error || 'Unable to update like status.');
      }

      setPosts((currentPosts) =>
        currentPosts.map((post) =>
          post.id === postId
            ? {
                ...post,
                isLiked: !currentLiked,
                likes: Math.max(0, post.likes + (currentLiked ? -1 : 1)),
              }
            : post,
        ),
      );

      if (selectedPost && selectedPost.id === postId) {
        setSelectedPost({
          ...selectedPost,
          isLiked: !currentLiked,
          likes: Math.max(0, selectedPost.likes + (currentLiked ? -1 : 1)),
        });
      }
    } catch (error) {
      console.error('toggle like failed', error);
    }
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
          post.id === postId ? { ...post, comments: post.comments + 1 } : post,
        ),
      );

      if (selectedPost && selectedPost.id === postId) {
        setSelectedPost({ ...selectedPost, comments: selectedPost.comments + 1 });
      }

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

  const summary = useMemo(
    () => [
      { label: 'Posts', value: String(posts.length) },
      { label: 'Followers', value: String(followerCount) },
      { label: 'Following', value: String(followingCount) },
    ],
    [posts.length, followerCount, followingCount],
  );

  const dedupedCustomLinks = useMemo(() => {
    const socialUrls = [user.website, user.linkedin, user.instagram]
      .filter((value): value is string => Boolean(value))
      .map(normalizeLinkUrl);

    return (user.customLinks ?? []).filter((link) => {
      if (!link.url) return false;
      return !socialUrls.includes(normalizeLinkUrl(link.url));
    });
  }, [user.customLinks, user.website, user.linkedin, user.instagram]);

  return (
    <div className="min-h-screen bg-[#f3f7fb] px-4 py-8 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <div className="mb-6 flex items-center justify-between">
          <button
            type="button"
            onClick={() => navigate('/newsfeed')}
            className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 shadow-sm"
          >
            ← Back to newsfeed
          </button>
          <div className="flex items-center gap-3">
            {!isOwnProfile && (
              <div className="flex items-center gap-2">
                {isFollowing ? (
                  <button
                    type="button"
                    disabled
                    className="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-2 text-sm font-semibold text-emerald-700 shadow-sm"
                  >
                    Following
                  </button>
                ) : null}
                <button
                  type="button"
                  disabled={isUpdatingFollowState}
                  onClick={handleFollowToggle}
                  className={`rounded-2xl px-4 py-2 text-sm font-semibold shadow-sm transition ${
                    isFollowing
                      ? 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
                      : 'bg-gradient-to-r from-sky-500 to-blue-600 text-white shadow-lg shadow-sky-300/40'
                  } disabled:cursor-not-allowed disabled:opacity-70`}
                >
                  {isUpdatingFollowState
                    ? isFollowing ? 'Unfollowing...' : 'Following...'
                    : isFollowing ? 'Unfollow' : 'Follow'}
                </button>
              </div>
            )}
            {isOwnProfile && (
              <button
                type="button"
                onClick={() => navigate('/profile/edit')}
                className="rounded-2xl border border-sky-200 bg-sky-50 px-4 py-2 text-sm font-semibold text-sky-700 shadow-sm"
              >
                Edit Profile
              </button>
            )}
            <button
              type="button"
              onClick={() => navigate('/newsfeed')}
              className="rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-sky-300/40"
            >
              Newsfeed
            </button>
          </div>
        </div>

        <div className="rounded-[30px] border border-sky-100 bg-white p-6 shadow-sm sm:p-8">
          <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
            <div className="flex items-center gap-4">
             <div className="relative">
                <div
                  onClick={() => {
                    if (isOwnProfile && !isUploadingAvatar) {
                      avatarInputRef.current?.click();
                    }
                  }}
                  className={`group relative flex h-20 w-20 items-center justify-center overflow-hidden rounded-[24px] bg-gradient-to-br from-sky-500 to-blue-600 text-2xl font-black text-white shadow-lg shadow-sky-300/40 transition ${
                    isOwnProfile ? 'cursor-pointer' : 'cursor-default'
                  } ${isUploadingAvatar ? 'opacity-80' : ''}`}
                  title={isOwnProfile ? "Change profile picture" : "Profile avatar"}
                >
                  {user.profileImageUrl ? (
                    <img src={user.profileImageUrl} alt={user.name} className="h-full w-full object-cover" />
                  ) : (
                    <span className="z-10">{getInitials(user.name)}</span>
                  )}

                  {isOwnProfile && (
                    <div className="absolute inset-0 z-20 flex items-center justify-center bg-slate-900/60 text-xs font-bold tracking-wide text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100">
                      {isUploadingAvatar ? 'Uploading...' : 'Change'}
                    </div>
                  )}
                </div>

                {isOwnProfile && (
                  <input
                    ref={avatarInputRef}
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={handleAvatarUpload}
                  />
                )}
              </div>

              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-600">Profile</p>
                <h1 className="mt-2 text-3xl font-black text-slate-900">{user.name}</h1>
                <p className="text-sm text-slate-500">{user.username}</p>
                {user.bio && <p className="mt-2 max-w-xl text-sm leading-6 text-slate-600">{user.bio}</p>}
                {(user.website || user.linkedin || user.instagram || (user.customLinks && user.customLinks.length > 0)) && (
                  <div className="mt-3 flex flex-wrap gap-2 text-xs">
                    {user.website && (
                      <a
                        href={user.website}
                        target="_blank"
                        rel="noreferrer"
                        className="rounded-full bg-sky-50 px-2.5 py-1 font-medium text-sky-700"
                      >
                        Website
                      </a>
                    )}
                    {user.linkedin && (
                      <a
                        href={user.linkedin}
                        target="_blank"
                        rel="noreferrer"
                        className="rounded-full bg-sky-50 px-2.5 py-1 font-medium text-sky-700"
                      >
                        LinkedIn
                      </a>
                    )}
                    {user.instagram && (
                      <a
                        href={user.instagram}
                        target="_blank"
                        rel="noreferrer"
                        className="rounded-full bg-sky-50 px-2.5 py-1 font-medium text-sky-700"
                      >
                        Instagram
                      </a>
                    )}
                    {dedupedCustomLinks.map((link) => (
                      <a
                        key={link.id}
                        href={link.url}
                        target="_blank"
                        rel="noreferrer"
                        className="rounded-full bg-sky-50 px-2.5 py-1 font-medium text-sky-700"
                      >
                        {link.label || link.url}
                      </a>
                    ))}
                  </div>
                )}
              </div>
            </div>

            <div className="grid grid-cols-3 gap-3 sm:min-w-[260px]">
              {summary.map((item) => (
                <div key={item.label} className="rounded-2xl bg-sky-50 px-3 py-3 text-center">
                  <div className="text-xl font-black text-slate-900">{item.value}</div>
                  <div className="text-[11px] uppercase tracking-[0.14em] text-slate-500">{item.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-6">
          {error && (
            <div className="mb-4 rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
              Loading profile posts...
            </div>
          ) : posts.length === 0 ? (
            <div className="rounded-[28px] border border-sky-100 bg-white p-6 text-sm text-slate-500 shadow-sm">
              No posts yet for this profile.
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-[2px] bg-slate-200">
              {posts.map((post) => (
                <article key={post.id} className="w-full bg-white">
                  {post.imageUrls.length > 0 && (
                    <div className="overflow-hidden">
                      <div className="grid grid-cols-1">
                        {post.imageUrls.slice(0, 4).map((imageUrl, index) => (
                          <button
                            key={`${post.id}-${index}`}
                            type="button"
                            onClick={() => {
                              setSelectedPost(post);
                              setSelectedImageIndex(index);
                              setShowComments(false);
                              fetchCommentsForPost(post.id);
                            }}
                            className="group relative aspect-square w-full overflow-hidden bg-slate-100"
                          >
                            <img
                              src={imageUrl}
                              alt={`Profile post ${index + 1}`}
                              className="h-full w-full object-cover transition duration-200 group-hover:scale-105"
                            />
                            <div className="absolute inset-0 flex items-center justify-center gap-4 bg-slate-900/45 opacity-0 transition duration-200 group-hover:opacity-100">
                              <span className={`text-sm font-semibold ${post.isLiked ? 'text-rose-500' : 'text-white'}`}>
                                ♥ {post.likes}
                              </span>
                              <span className="text-sm font-semibold text-white">💬 {post.comments}</span>
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  )}
                </article>
              ))}
            </div>
          )}
        </div>
      </div>

      {selectedPost && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 p-4">
          <div className="max-h-[90vh] w-full max-w-5xl overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-[0_25px_80px_rgba(15,23,42,0.25)]">
            <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-sky-600">Full image</p>
                <h3 className="mt-1 text-lg font-bold text-slate-900">{user.name}</h3>
              </div>
              <button
                type="button"
                onClick={() => setSelectedPost(null)}
                className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-500"
              >
                Close
              </button>
            </div>

            <div className="grid max-h-[calc(90vh-80px)] overflow-hidden lg:grid-cols-[1.4fr_0.8fr]">
              <div className="flex items-center justify-center bg-slate-100 p-4">
                <div className="relative w-full">
                  <img
                    src={selectedPost.imageUrls[selectedImageIndex]}
                    alt="Selected profile media"
                    className="max-h-[70vh] w-full rounded-2xl object-contain"
                  />

                  {selectedPost.imageUrls.length > 1 && (
                    <div className="mt-3 flex items-center justify-center gap-2">
                      <button
                        type="button"
                        onClick={() => setSelectedImageIndex((prev) => Math.max(0, prev - 1))}
                        disabled={selectedImageIndex === 0}
                        className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 disabled:cursor-not-allowed disabled:opacity-40"
                      >
                        Prev
                      </button>
                      <span className="text-sm text-slate-500">
                        {selectedImageIndex + 1} / {selectedPost.imageUrls.length}
                      </span>
                      <button
                        type="button"
                        onClick={() => setSelectedImageIndex((prev) => Math.min(selectedPost.imageUrls.length - 1, prev + 1))}
                        disabled={selectedImageIndex === selectedPost.imageUrls.length - 1}
                        className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-600 disabled:cursor-not-allowed disabled:opacity-40"
                      >
                        Next
                      </button>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex flex-col overflow-y-auto border-t border-slate-100 bg-white p-4 lg:border-l lg:border-t-0">
                <p className="text-sm leading-7 text-slate-700">{selectedPost.content}</p>

                <div className="mt-4 flex items-center gap-3 border-b border-slate-100 pb-4">
                  <button
                    type="button"
                    onClick={() => handleToggleLike(selectedPost.id, Boolean(selectedPost.isLiked))}
                    className={`flex cursor-pointer items-center gap-2 rounded-full px-3 py-1.5 text-sm font-medium ${selectedPost.isLiked ? 'bg-rose-50 text-rose-600' : 'bg-slate-100 text-slate-600'}`}
                  >
                    <span>{selectedPost.isLiked ? '♥' : '♡'}</span>
                    <span>{selectedPost.likes}</span>
                  </button>

                  <button
                    type="button"
                    onClick={async () => {
                      if (showComments) {
                        setShowComments(false);
                        return;
                      }

                      setShowComments(true);
                      await fetchCommentsForPost(selectedPost.id);
                    }}
                    className="flex cursor-pointer items-center gap-2 rounded-full bg-slate-100 px-3 py-1.5 text-sm font-medium text-slate-600"
                  >
                    <span>💬</span>
                    <span>{selectedPost.comments}</span>
                  </button>
                </div>

                <div className="mt-4 flex-1 space-y-3">
                  <div className="rounded-2xl border border-sky-100 bg-sky-50 p-3">
                    <textarea
                      value={commentDrafts[selectedPost.id] ?? ''}
                      onChange={(event) =>
                        setCommentDrafts((currentDrafts) => ({
                          ...currentDrafts,
                          [selectedPost.id]: event.target.value,
                        }))
                      }
                      rows={3}
                      placeholder="Write a comment..."
                      className="w-full resize-none border-0 bg-transparent text-sm text-slate-700 placeholder:text-slate-400 focus:outline-none"
                    />
                    <div className="mt-2 flex justify-end">
                      <button
                        type="button"
                        onClick={() => handleAddComment(selectedPost.id)}
                        className="cursor-pointer rounded-full bg-sky-600 px-3 py-1.5 text-sm font-semibold text-white"
                      >
                        Post comment
                      </button>
                    </div>
                  </div>

                  {showComments && (
                    <div className="space-y-3">
                      {commentLoading[selectedPost.id] ? (
                        <div className="text-sm text-slate-500">Loading comments...</div>
                      ) : (commentsByPost[selectedPost.id] ?? []).length > 0 ? (
                        (commentsByPost[selectedPost.id] ?? []).map((comment) => {
                          const replyDraftKey = `${selectedPost.id}:${comment.id}`;
                          const isReplying = replyingToComment[selectedPost.id] === comment.id;
                          const isExpanded = expandedRepliesByComment[`${selectedPost.id}:${comment.id}`] ?? false;

                          return (
                            <div key={comment.id} className="rounded-2xl bg-slate-50 px-3 py-2 ring-1 ring-slate-100">
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
                                      [`${selectedPost.id}:${comment.id}`]: !isExpanded,
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
                                      [selectedPost.id]: isReplying ? null : comment.id,
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
                                      onClick={() => setReplyingToComment((current) => ({ ...current, [selectedPost.id]: null }))}
                                      className="cursor-pointer rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-500"
                                    >
                                      Close
                                    </button>
                                    <button
                                      type="button"
                                      onClick={() => handleAddComment(selectedPost.id, comment.id)}
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
                                    <div key={reply.id} className="rounded-xl bg-white px-2.5 py-2 ring-1 ring-slate-100">
                                      <div className="mb-1 flex items-start gap-2">
                                        <div className="flex h-7 w-7 items-center justify-center overflow-hidden rounded-full bg-gradient-to-br from-sky-300 to-blue-500 text-[8px] font-bold text-white ring-1 ring-sky-100">
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
                        })
                      ) : (
                        <div className="rounded-2xl bg-slate-50 px-3 py-2 text-sm text-slate-500 ring-1 ring-slate-100">
                          No comments yet.
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
