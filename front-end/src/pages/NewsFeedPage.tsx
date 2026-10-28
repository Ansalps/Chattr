import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { searchUsersByUsername } from '../features/chat/chatService';
import { useFeedManagement } from '../features/feed/useFeedManagement';
import { FeedPostCard } from '../features/newsfeed/FeedPostCard';
import { PostComposerModal } from '../features/newsfeed/PostComposerModal';
import { fetchCurrentUserProfile, logoutCurrentUser } from '../features/profile/profileService';
import { formatCount } from '../shared/utils/formatting';
import { PROFILE_STORAGE_KEY, readStoredProfile } from '../shared/utils/normalizers';

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

function useUserSearch(currentUserId: string) {
  const [query, setQueryValue] = useState('');
  const [results, setResults] = useState<Array<{ id: string; name: string; username: string; profileImageUrl: string }>>([]);
  const [isSearching, setIsSearching] = useState(false);

  const setQuery = (nextQuery: string) => {
    setQueryValue(nextQuery);
    if (!nextQuery.trim()) {
      setResults([]);
    }
  };

  useEffect(() => {
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        setIsSearching(true);
        const normalizedResults = await searchUsersByUsername(trimmedQuery);
        setResults(normalizedResults.filter((user) => user.id !== currentUserId));
      } catch (searchError) {
        if (searchError instanceof DOMException && searchError.name === 'AbortError') {
          return;
        }

        console.error('search users failed', searchError);
        setResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [currentUserId, query]);

  return {
    query,
    setQuery,
    results,
    setResults,
    isSearching,
  };
}

export default function NewsFeedPage() {
  const navigate = useNavigate();
  const [currentUserProfile, setCurrentUserProfile] = useState({
    id: '',
    name: '',
    username: '',
    profileImageUrl: '',
    followerCount: 0,
    followingCount: 0,
    postCount: 0,
  });
  const storedProfile = readStoredProfile();
  const currentUserId = currentUserProfile.id || storedProfile.id || '';
  const storedDisplayName = (storedProfile.name || '').trim();
  const storedUserName = (storedProfile.username || '').replace(/^@/, '').trim();
  const liveDisplayName = (currentUserProfile.name || '').trim();
  const liveUserName = (currentUserProfile.username || '').replace(/^@/, '').trim();
  const currentUserName = (() => {
    const candidates = [
      liveDisplayName,
      storedDisplayName,
      liveUserName,
      storedUserName,
    ].filter((value): value is string => Boolean(value && value.trim() && value.toLowerCase() !== 'user profile'));

    return candidates[0] || 'User';
  })();
  const currentUserHandle = currentUserProfile.username || storedProfile.username || '@user';
  const currentProfileImageUrl =
    (currentUserProfile.profileImageUrl ? currentUserProfile.profileImageUrl : storedProfile.profileImageUrl) || '';

  const userSearchState = useUserSearch(currentUserId);
  const fallbackPosts = useMemo(() => [], []);
  const feed = useFeedManagement({
    fallbackPosts,
    currentUserId,
    currentUserHandle,
    currentUserName,
    currentProfileImageUrl,
  });
  const {
    posts,
    error,
    isLoading,
    commentDrafts,
    setCommentDrafts,
    replyDrafts,
    setReplyDrafts,
    replyingToComment,
    setReplyingToComment,
    expandedRepliesByComment,
    setExpandedRepliesByComment,
    openCommentFor,
    setOpenCommentFor,
    commentsByPost,
    commentLoading,
    postMenuOpenFor,
    setPostMenuOpenFor,
    isComposerOpen,
    postText,
    setPostText,
    selectedFiles,
    setSelectedFiles,
    isPosting,
    composerError,
    composerSuccess,
    editingPostId,
    handleToggleLike,
    handleOpenComments,
    handleAddComment,
    handleDeletePost,
    openComposerForNewPost,
    openComposerForEdit,
    closeComposer,
    handleCreatePost,
  } = feed;

  const {
    query: userSearch,
    setQuery: setUserSearch,
    results: userSearchResults,
    isSearching: isSearchingUsers,
    setResults: setUserSearchResults,
  } = userSearchState;

  useEffect(() => {
    const loadProfile = async () => {
      try {
        const nextProfile = await fetchCurrentUserProfile();
        setCurrentUserProfile(nextProfile);
      } catch (profileError) {
        console.error('fetch user profile info failed', profileError);
      }
    };

    void loadProfile();
  }, []);

  const quickStats = useMemo(
    () => [
      { label: 'Followers', value: formatCount(currentUserProfile.followerCount) },
      { label: 'Following', value: formatCount(currentUserProfile.followingCount) },
      { label: 'Posts', value: formatCount(currentUserProfile.postCount) },
    ],
    [currentUserProfile.followerCount, currentUserProfile.followingCount, currentUserProfile.postCount],
  );

  const goToProfile = (userId?: string) => {
    const safeUserId = String(userId ?? '').trim();
    navigate(safeUserId && safeUserId !== 'me' ? `/profile/${encodeURIComponent(safeUserId)}` : '/profile');
  };

  const handleUserSearchSelect = (userId: string) => {
    setUserSearch('');
    setUserSearchResults([]);
    navigate(`/profile/${encodeURIComponent(userId)}`);
  };

  const handleLogout = async () => {
    try {
      await logoutCurrentUser();
    } catch (error) {
      console.error('logout failed', error);
    } finally {
      localStorage.removeItem(PROFILE_STORAGE_KEY);
      window.location.href = '/login';
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
                onClick={openComposerForNewPost}
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

          </aside>

          <main className="space-y-5">
            <div className="rounded-[28px] border border-sky-100 bg-white p-4 shadow-sm">
              <button
                type="button"
                onClick={openComposerForNewPost}
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
            ) : posts.length === 0 ? (
              <div className="rounded-[28px] border border-sky-100 bg-white p-10 text-center shadow-sm">
                <p className="text-lg font-semibold text-slate-800">You and your followers have no posts.</p>
                <p className="mt-2 text-sm text-slate-500">Create the first post to start the conversation.</p>
              </div>
            ) : (
              posts.map((post) => (
                <FeedPostCard
                  key={post.id}
                  post={post}
                  currentUserId={currentUserId}
                  currentUserHandle={currentUserHandle}
                  currentUserName={currentUserName}
                  currentProfileImageUrl={currentProfileImageUrl}
                  postMenuOpenFor={postMenuOpenFor}
                  setPostMenuOpenFor={setPostMenuOpenFor}
                  openCommentFor={openCommentFor}
                  setOpenCommentFor={setOpenCommentFor}
                  onGoToProfile={goToProfile}
                  onToggleLike={handleToggleLike}
                  onOpenComments={handleOpenComments}
                  onEditPost={openComposerForEdit}
                  onDeletePost={handleDeletePost}
                  commentDrafts={commentDrafts}
                  setCommentDrafts={setCommentDrafts}
                  commentLoading={commentLoading}
                  commentsByPost={commentsByPost}
                  replyDrafts={replyDrafts}
                  setReplyDrafts={setReplyDrafts}
                  replyingToComment={replyingToComment}
                  setReplyingToComment={setReplyingToComment}
                  expandedRepliesByComment={expandedRepliesByComment}
                  setExpandedRepliesByComment={setExpandedRepliesByComment}
                  onAddComment={handleAddComment}
                />
              ))
            )}
          </main>

          <aside className="space-y-6" />
        </div>
      </div>

      <PostComposerModal
        isOpen={isComposerOpen}
        editingPostId={editingPostId}
        postText={postText}
        setPostText={setPostText}
        selectedFiles={selectedFiles}
        setSelectedFiles={setSelectedFiles}
        composerError={composerError}
        composerSuccess={composerSuccess}
        isPosting={isPosting}
        onClose={closeComposer}
        onPublish={handleCreatePost}
      />

    </div>
  );
}
