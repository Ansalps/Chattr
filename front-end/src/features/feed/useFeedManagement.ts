import { useEffect, useState } from 'react';
import {
  addCommentToPost,
  createPostWithMedia,
  deletePostById,
  fetchCommentsForPost,
  fetchNewsFeed,
  toggleLikePost,
  updatePostContent,
} from './feedService';
import type { FeedComment, FeedPost } from '../../shared/types';
import { formatFriendlyTime } from '../../shared/utils/formatting';

type FeedManagementOptions = {
  fallbackPosts: FeedPost[];
  currentUserId: string;
  currentUserHandle: string;
  currentUserName: string;
  currentProfileImageUrl: string;
};

export function useFeedManagement({
  fallbackPosts,
  currentUserId,
  currentUserHandle,
  currentUserName,
  currentProfileImageUrl,
}: FeedManagementOptions) {
  const [posts, setPosts] = useState<FeedPost[]>(() => fallbackPosts);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [commentDrafts, setCommentDrafts] = useState<Record<string, string>>({});
  const [replyDrafts, setReplyDrafts] = useState<Record<string, string>>({});
  const [replyingToComment, setReplyingToComment] = useState<Record<string, string | null>>({});
  const [expandedRepliesByComment, setExpandedRepliesByComment] = useState<Record<string, boolean>>({});
  const [openCommentFor, setOpenCommentFor] = useState<string | null>(null);
  const [commentsByPost, setCommentsByPost] = useState<Record<string, FeedComment[]>>({});
  const [commentLoading, setCommentLoading] = useState<Record<string, boolean>>({});
  const [postMenuOpenFor, setPostMenuOpenFor] = useState<string | null>(null);
  const [isComposerOpen, setIsComposerOpen] = useState(false);
  const [postText, setPostText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isPosting, setIsPosting] = useState(false);
  const [composerError, setComposerError] = useState('');
  const [composerSuccess, setComposerSuccess] = useState('');
  const [editingPostId, setEditingPostId] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    const loadNewsFeed = async () => {
      try {
        const nextPosts = await fetchNewsFeed();
        if (!isMounted) return;
        setPosts(nextPosts);
      } catch (feedError) {
        const message =
          feedError instanceof Error ? feedError.message : 'Unable to load your news feed right now.';

        if (!isMounted) return;
        setError(message);
        setPosts([]);
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    void loadNewsFeed();

    return () => {
      isMounted = false;
    };
  }, []);

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
    } catch (feedError) {
      console.error('toggle like failed', feedError);
    }
  };

  const loadCommentsForPost = async (postId: string) => {
    if (commentLoading[postId]) return;

    setCommentLoading((current) => ({ ...current, [postId]: true }));

    try {
      const normalizedComments = await fetchCommentsForPost(postId);
      setCommentsByPost((current) => ({ ...current, [postId]: normalizedComments }));
    } catch (feedError) {
      console.error('fetch comments failed', feedError);
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
          post.id === postId ? { ...post, comments: post.comments + 1 } : post,
        ),
      );

      if (parentCommentId) {
        setReplyDrafts((currentDrafts) => ({ ...currentDrafts, [draftKey]: '' }));
        setReplyingToComment((current) => ({ ...current, [postId]: null }));
      } else {
        setCommentDrafts((currentDrafts) => ({ ...currentDrafts, [postId]: '' }));
      }

      await loadCommentsForPost(postId);
    } catch (feedError) {
      console.error('add comment failed', feedError);
    }
  };

  const handleDeletePost = async (postId: string) => {
    try {
      await deletePostById(postId);
      setPosts((currentPosts) => currentPosts.filter((post) => post.id !== postId));
      setPostMenuOpenFor(null);
    } catch (feedError) {
      console.error('delete post failed', feedError);
    }
  };

  const openComposerForNewPost = () => {
    setEditingPostId(null);
    setPostText('');
    setSelectedFiles([]);
    setComposerError('');
    setComposerSuccess('');
    setIsComposerOpen(true);
    setPostMenuOpenFor(null);
  };

  const openComposerForEdit = (post: FeedPost) => {
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
    setComposerSuccess('');
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
    } catch (feedError) {
      const message =
        feedError instanceof Error ? feedError.message : 'Unable to create your post right now.';
      setComposerError(message);
    } finally {
      setIsPosting(false);
    }
  };

  return {
    posts,
    setPosts,
    error,
    setError,
    isLoading,
    setIsLoading,
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
    setCommentsByPost,
    commentLoading,
    setCommentLoading,
    postMenuOpenFor,
    setPostMenuOpenFor,
    isComposerOpen,
    setIsComposerOpen,
    postText,
    setPostText,
    selectedFiles,
    setSelectedFiles,
    isPosting,
    setIsPosting,
    composerError,
    setComposerError,
    composerSuccess,
    setComposerSuccess,
    editingPostId,
    setEditingPostId,
    currentUserId,
    currentUserHandle,
    currentUserName,
    currentProfileImageUrl,
    handleToggleLike,
    loadCommentsForPost,
    handleOpenComments,
    handleAddComment,
    handleDeletePost,
    openComposerForNewPost,
    openComposerForEdit,
    closeComposer,
    handleCreatePost,
  };
}
