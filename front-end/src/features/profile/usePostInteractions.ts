import { useCallback, useState } from 'react';
import type { Dispatch, SetStateAction } from 'react';
import type { FeedComment, ProfilePost } from './profile.types';
import { normalizeCommentTree } from './profile.utils';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

type UsePostInteractionsOptions = {
  setPosts: Dispatch<SetStateAction<ProfilePost[]>>;
  selectedPost: ProfilePost | null;
  setSelectedPost: Dispatch<SetStateAction<ProfilePost | null>>;
};

export function usePostInteractions({
  setPosts,
  selectedPost,
  setSelectedPost,
}: UsePostInteractionsOptions) {
  const [commentDrafts, setCommentDrafts] = useState<Record<string, string>>({});
  const [replyDrafts, setReplyDrafts] = useState<Record<string, string>>({});
  const [replyingToComment, setReplyingToComment] = useState<Record<string, string | null>>({});
  const [expandedRepliesByComment, setExpandedRepliesByComment] = useState<Record<string, boolean>>({});
  const [commentsByPost, setCommentsByPost] = useState<Record<string, FeedComment[]>>({});
  const [commentLoading, setCommentLoading] = useState<Record<string, boolean>>({});
  const [showComments, setShowComments] = useState(false);

  const fetchCommentsForPost = useCallback(
    async (postId: string) => {
      if (commentLoading[postId]) {
        return;
      }

      setCommentLoading((current) => ({ ...current, [postId]: true }));

      try {
        const response = await fetch(`${API_BASE_URL}/user/post/${postId}/comment?limit=20&offset=0`, {
          method: 'GET',
          credentials: 'include',
        });

        const payload = await response.json().catch(() => ({}));

        if (!response.ok || Number((payload as Record<string, unknown>).status_code ?? 0) >= 400) {
          throw new Error(
            ((payload as Record<string, unknown>).message as string) ||
              ((payload as Record<string, unknown>).error as string) ||
              'Unable to load comments.',
          );
        }

        const data = (payload as Record<string, unknown>).data ?? payload;
        const commentList = Array.isArray((data as Record<string, unknown>)?.Comments)
          ? ((data as Record<string, unknown>).Comments as unknown[])
          : Array.isArray((data as Record<string, unknown>)?.comments)
            ? ((data as Record<string, unknown>).comments as unknown[])
            : Array.isArray((payload as Record<string, unknown>).Comments)
              ? ((payload as Record<string, unknown>).Comments as unknown[])
              : Array.isArray((payload as Record<string, unknown>).comments)
                ? ((payload as Record<string, unknown>).comments as unknown[])
                : [];

        setCommentsByPost((current) => ({
          ...current,
          [postId]: normalizeCommentTree(commentList, postId),
        }));
      } catch (error) {
        console.error('fetch comments failed', error);
        setCommentsByPost((current) => ({ ...current, [postId]: [] }));
      } finally {
        setCommentLoading((current) => ({ ...current, [postId]: false }));
      }
    },
    [commentLoading],
  );

  const handleToggleLike = useCallback(
    async (postId: string, currentLiked: boolean) => {
      try {
        const response = await fetch(`${API_BASE_URL}/user/post/${postId}/${currentLiked ? 'unlike' : 'like'}`, {
          method: currentLiked ? 'DELETE' : 'POST',
          credentials: 'include',
        });

        if (!response.ok) {
          const payload = await response.json().catch(() => ({}));
          throw new Error(
            ((payload as Record<string, unknown>).message as string) ||
              ((payload as Record<string, unknown>).error as string) ||
              'Unable to update like status.',
          );
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
          setSelectedPost((current) =>
            current
              ? {
                  ...current,
                  isLiked: !currentLiked,
                  likes: Math.max(0, current.likes + (currentLiked ? -1 : 1)),
                }
              : current,
          );
        }
      } catch (error) {
        console.error('toggle like failed', error);
      }
    },
    [selectedPost, setPosts, setSelectedPost],
  );

  const handleAddComment = useCallback(
    async (postId: string, parentCommentId?: string) => {
      const draftKey = parentCommentId ? `${postId}:${parentCommentId}` : postId;
      const commentText = (parentCommentId ? (replyDrafts[draftKey] ?? '') : (commentDrafts[postId] ?? '')).trim();
      if (!commentText) {
        return;
      }

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
          throw new Error(
            ((payload as Record<string, unknown>).message as string) ||
              ((payload as Record<string, unknown>).error as string) ||
              'Unable to add your comment.',
          );
        }

        setPosts((currentPosts) =>
          currentPosts.map((post) =>
            post.id === postId ? { ...post, comments: post.comments + 1 } : post,
          ),
        );

        if (selectedPost && selectedPost.id === postId) {
          setSelectedPost((current) =>
            current ? { ...current, comments: current.comments + 1 } : current,
          );
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
    },
    [commentDrafts, fetchCommentsForPost, replyDrafts, selectedPost, setPosts, setSelectedPost],
  );

  return {
    commentDrafts,
    setCommentDrafts,
    replyDrafts,
    setReplyDrafts,
    replyingToComment,
    setReplyingToComment,
    expandedRepliesByComment,
    setExpandedRepliesByComment,
    commentsByPost,
    setCommentsByPost,
    commentLoading,
    showComments,
    setShowComments,
    fetchCommentsForPost,
    handleToggleLike,
    handleAddComment,
  };
}
