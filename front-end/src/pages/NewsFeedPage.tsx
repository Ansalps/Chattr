import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  buildChatMessagePayload,
  connectChatSocket,
  fetchRecentChatProfiles,
  loadChatSessionData,
  searchUsersByUsername,
  searchUsersForSelection,
} from '../features/chat/chatService';
import {
  addMemberToGroup,
  createGroup,
  fetchGroupMembers,
  removeMemberFromGroup,
  updateGroupProfileImage,
} from '../features/groups/groupService';
import {
  addCommentToPost,
  createPostWithMedia,
  deletePostById,
  fetchCommentsForPost,
  fetchNewsFeed,
  toggleLikePost,
  updatePostContent,
} from '../features/feed/feedService';
import { ChatPanel } from '../features/chat/ChatPanel';
import { PostComposerModal } from '../features/newsfeed/PostComposerModal';
import { fetchCurrentUserProfile, logoutCurrentUser } from '../features/profile/profileService';
import type { ChatSession, ChatUserProfile, FeedComment, FeedPost } from '../shared/types';
import { formatCount, formatFriendlyTime } from '../shared/utils/formatting';
import { PROFILE_STORAGE_KEY, readStoredProfile } from '../shared/utils/normalizers';

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
  const [chatSessions, setChatSessions] = useState<ChatSession[]>([]);
  const [isCreatingGroup, setIsCreatingGroup] = useState(false);
  const [groupNameDraft, setGroupNameDraft] = useState('');
  const [groupMemberSearch, setGroupMemberSearch] = useState('');
  const [groupMemberResults, setGroupMemberResults] = useState<ChatUserProfile[]>([]);
  const [selectedGroupMembers, setSelectedGroupMembers] = useState<ChatUserProfile[]>([]);
  const [chatSelectedUser, setChatSelectedUser] = useState<
    (ChatUserProfile & { conversationId?: string; isGroup?: boolean; groupId?: string }) | null
  >(null);
  const [chatMessages, setChatMessages] = useState<
    Array<{
      id: string;
      senderId: string;
      senderName: string;
      senderProfileImageUrl: string;
      content: string;
      createdAt: string;
      mine: boolean;
    }>
  >([]);
  const [chatDraft, setChatDraft] = useState('');
  const [chatSocket, setChatSocket] = useState<WebSocket | null>(null);
  const [chatError, setChatError] = useState('');
  const [isLoadingChatHistory, setIsLoadingChatHistory] = useState(false);
  const [isGroupDetailOpen, setIsGroupDetailOpen] = useState(false);
  const [groupMembers, setGroupMembers] = useState<ChatUserProfile[]>([]);
  const [groupCreatorId, setGroupCreatorId] = useState('');

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
    const loadProfile = async () => {
      try {
        const nextProfile = await fetchCurrentUserProfile();
        setCurrentUserProfile(nextProfile);
      } catch (profileError) {
        console.error('fetch user profile info failed', profileError);
      }
    };

    const loadNewsFeed = async () => {
      try {
        const nextPosts = await fetchNewsFeed();
        setPosts(nextPosts);
      } catch (feedError) {
        const message =
          feedError instanceof Error ? feedError.message : 'Unable to load your news feed right now.';

        setError(message);
        setPosts(fallbackPosts);
      } finally {
        setIsLoading(false);
      }
    };

    loadProfile();
    loadNewsFeed();
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

  const loadRecentChats = async () => {
    try {
      const normalizedChats = await fetchRecentChatProfiles();
      setChatSessions(normalizedChats);
    } catch (chatError) {
      console.error('loading recent chats failed', chatError);
      setChatSessions([]);
    }
  };

  const openChatSession = async (user: ChatUserProfile & { conversationId?: string; isGroup?: boolean; groupId?: string }) => {
    setChatSelectedUser(user);
    setChatSearch('');
    setChatSearchResults([]);
    setChatDraft('');
    setChatError('');
    setChatMessages([]);
    setIsLoadingChatHistory(true);
    setIsGroupDetailOpen(false);

    try {
      const { messages, groupMembers: nextGroupMembers } = await loadChatSessionData({
        conversationId: user.conversationId,
        currentUserId: currentUserId || storedProfile.id || '',
        groupId: user.groupId,
        isGroup: user.isGroup,
      });

      setChatMessages(messages);
      setGroupMembers(nextGroupMembers.filter((member) => member.id));
    } catch (historyError) {
      console.error('loading chat history failed', historyError);
      setChatMessages([]);
    } finally {
      setIsLoadingChatHistory(false);
    }
  };

  const refreshGroupMembers = async (groupId: string) => {
    if (!groupId) {
      setGroupMembers([]);
      return;
    }

    try {
      const { members, creatorId } = await fetchGroupMembers(groupId);
      setGroupCreatorId(creatorId || currentUserId || '');
      setGroupMembers(members.filter((member) => member.id));
    } catch (error) {
      console.error('refresh group members failed', error);
      setGroupMembers([]);
    }
  };

  const handleAddMemberToGroup = async (member: ChatUserProfile) => {
    if (!chatSelectedUser?.groupId) {
      return;
    }

    try {
      await addMemberToGroup(chatSelectedUser.groupId, member.id);
      setGroupMemberSearch('');
      setGroupMemberResults([]);
      await refreshGroupMembers(chatSelectedUser.groupId);
      setChatError('');
    } catch (error) {
      console.error('add member to group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to add member to group.');
    }
  };

  const handleRemoveMemberFromGroup = async (member: ChatUserProfile) => {
    if (!chatSelectedUser?.groupId) {
      return;
    }

    try {
      await removeMemberFromGroup(chatSelectedUser.groupId, member.id);
      setGroupMembers((current) => current.filter((candidate) => candidate.id !== member.id));
      setChatError('');
    } catch (error) {
      console.error('remove member from group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to remove member from group.');
    }
  };

  const handleGroupProfileImageChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !chatSelectedUser?.groupId) {
      return;
    }

    try {
      const finalProfileImageUrl = await updateGroupProfileImage(chatSelectedUser.groupId, file);

      setChatSelectedUser((current) =>
        current
          ? { ...current, profileImageUrl: finalProfileImageUrl }
          : current,
      );
      setChatSessions((current) =>
        current.map((session) =>
          session.groupId === chatSelectedUser.groupId || session.conversationId === chatSelectedUser.groupId
            ? { ...session, profileImageUrl: finalProfileImageUrl }
            : session,
        ),
      );
      setChatError('');
    } catch (error) {
      console.error('group profile image update failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to update group profile image.');
    } finally {
      event.target.value = '';
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
        const normalizedResults = await searchUsersByUsername(trimmedQuery);
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
      loadRecentChats();
    }
  }, [isChatOpen]);

  useEffect(() => {
    const trimmedQuery = groupMemberSearch.trim();
    if (!trimmedQuery) {
      setGroupMemberResults([]);
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        const normalizedResults = await searchUsersForSelection(trimmedQuery, currentUserId || storedProfile.id || '');
        setGroupMemberResults(normalizedResults);
      } catch (searchError) {
        if (searchError instanceof DOMException && searchError.name === 'AbortError') {
          return;
        }
        console.error('search members for group failed', searchError);
        setGroupMemberResults([]);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [groupMemberSearch, currentUserId, storedProfile.id]);

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
        const normalizedResults = await searchUsersForSelection(trimmedQuery, currentUserId || storedProfile.id || '');
        setChatSearchResults(normalizedResults);
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

    const socket = connectChatSocket({
      currentUserId: currentUserId || storedProfile.id || '',
      onError: (message) => setChatError(message),
      onMessage: (nextMessage) => {
        setChatMessages((currentMessages) => {
          const alreadyExists = currentMessages.some((message) => message.id === nextMessage.id);
          return alreadyExists ? currentMessages : [...currentMessages, nextMessage];
        });
      },
    });

    setChatSocket(socket);

    return () => {
      socket.close();
      setChatSocket(null);
    };
  }, [chatSelectedUser, currentUserId, storedProfile.id]);

  const toggleGroupMember = (member: { id: string; name: string; username: string; profileImageUrl: string }) => {
    setSelectedGroupMembers((current) => {
      const exists = current.some((candidate) => candidate.id === member.id);
      if (exists) {
        return current.filter((candidate) => candidate.id !== member.id);
      }
      return [...current, member];
    });
  };

  const handleCreateGroup = async () => {
    const cleanedGroupName = groupNameDraft.trim();
    if (!cleanedGroupName) {
      setChatError('Please provide a group name.');
      return;
    }

    if (selectedGroupMembers.length === 0) {
      setChatError('Add at least one participant to create a group.');
      return;
    }

    const payloadMembers = selectedGroupMembers
      .map((member) => Number(member.id))
      .filter((memberId) => Number.isFinite(memberId) && memberId > 0);

    if (payloadMembers.length === 0) {
      setChatError('Selected members are invalid. Please choose valid users.');
      return;
    }

    try {
      await createGroup(cleanedGroupName, payloadMembers);

      setIsCreatingGroup(false);
      setGroupNameDraft('');
      setGroupMemberSearch('');
      setGroupMemberResults([]);
      setSelectedGroupMembers([]);
      setChatError('');
      await fetchRecentChatProfiles();
    } catch (error) {
      console.error('create group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to create group.');
    }
  };

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

    const messagePayload = buildChatMessagePayload(chatSelectedUser, trimmedMessage);

    chatSocket.send(JSON.stringify(messagePayload));
    setChatMessages((currentMessages) => [
      ...currentMessages,
      {
        id: `local-${Date.now()}`,
        senderId: currentUserId || storedProfile.id || 'me',
        senderName: currentUserName || storedProfile.name || 'You',
        senderProfileImageUrl: currentProfileImageUrl || '',
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
      await toggleLikePost(postId, currentLiked);

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

  const loadCommentsForPost = async (postId: string) => {
    if (commentLoading[postId]) return;

    setCommentLoading((current) => ({ ...current, [postId]: true }));

    try {
      const normalizedComments = await fetchCommentsForPost(postId);
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

    await loadCommentsForPost(postId);
    setOpenCommentFor(postId);
  };

  const handleAddComment = async (postId: string, parentCommentId?: string) => {
    const draftKey = parentCommentId ? `${postId}:${parentCommentId}` : postId;
    const commentText = (parentCommentId ? (replyDrafts[draftKey] ?? '') : (commentDrafts[postId] ?? '')).trim();
    if (!commentText) return;

    try {
      await addCommentToPost(postId, commentText, parentCommentId);

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

      await loadCommentsForPost(postId);
    } catch (error) {
      console.error('add comment failed', error);
    }
  };

  const handleDeletePost = async (postId: string) => {
    try {
      await deletePostById(postId);
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
      await logoutCurrentUser();
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
        await updatePostContent(editingPostId, postText.trim());

        setPosts((currentPosts) =>
          currentPosts.map((post) =>
            post.id === editingPostId ? { ...post, content: postText.trim() || post.content } : post,
          ),
        );
        setComposerSuccess('Post updated successfully.');
        closeComposer();
        return;
      }

      const createdPost = await createPostWithMedia(postText.trim(), selectedFiles, {
        author: currentUserName,
        handle: currentUserHandle || '@user',
        profileImageUrl: currentProfileImageUrl || undefined,
      });

      setPosts((currentPosts) => [{
        ...createdPost,
        time: formatFriendlyTime(new Date().toISOString()),
      }, ...currentPosts]);
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

      <div className="fixed bottom-6 right-6 z-40 flex flex-col items-end gap-3">
        <ChatPanel
          isChatOpen={isChatOpen}
          chatSearch={chatSearch}
          setChatSearch={setChatSearch}
          chatSearchResults={chatSearchResults}
          isSearchingChatUsers={isSearchingChatUsers}
          chatSessions={chatSessions}
          openChatSession={openChatSession}
          isLoadingChatHistory={isLoadingChatHistory}
          chatSelectedUser={chatSelectedUser}
          setChatSelectedUser={setChatSelectedUser}
          chatMessages={chatMessages}
          chatDraft={chatDraft}
          setChatDraft={setChatDraft}
          chatError={chatError}
          handleSendChatMessage={handleSendChatMessage}
          isGroupDetailOpen={isGroupDetailOpen}
          setIsGroupDetailOpen={setIsGroupDetailOpen}
          groupMembers={groupMembers}
          groupCreatorId={groupCreatorId}
          groupMemberSearch={groupMemberSearch}
          setGroupMemberSearch={setGroupMemberSearch}
          groupMemberResults={groupMemberResults}
          refreshGroupMembers={refreshGroupMembers}
          handleAddMemberToGroup={handleAddMemberToGroup}
          handleRemoveMemberFromGroup={handleRemoveMemberFromGroup}
          handleGroupProfileImageChange={handleGroupProfileImageChange}
          setChatError={setChatError}
          isCreatingGroup={isCreatingGroup}
          setIsCreatingGroup={setIsCreatingGroup}
          groupNameDraft={groupNameDraft}
          setGroupNameDraft={setGroupNameDraft}
          selectedGroupMembers={selectedGroupMembers}
          setSelectedGroupMembers={setSelectedGroupMembers}
          toggleGroupMember={toggleGroupMember}
          handleCreateGroup={handleCreateGroup}
          onCloseChat={() => setIsChatOpen(false)}
        />

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
