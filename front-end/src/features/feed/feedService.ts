import type { FeedComment, FeedPost } from '../../shared/types';
import { getPostIdFromResponse, normalizeFeedPosts } from '../../shared/utils/normalizers';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export async function fetchNewsFeed() {
  const response = await fetch(`${API_BASE_URL}/user/newsfeed`, {
    method: 'GET',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load your news feed right now.');
  }

  const payloadObject = payload && typeof payload === 'object' ? (payload as Record<string, unknown>) : {};
  const nestedData = payloadObject.data && typeof payloadObject.data === 'object' ? (payloadObject.data as Record<string, unknown>) : {};

  const feedData = Array.isArray(payload)
    ? payload
    : Array.isArray(payloadObject.PostUserData)
      ? payloadObject.PostUserData
      : Array.isArray(payloadObject.postUserData)
        ? payloadObject.postUserData
        : Array.isArray(payloadObject.data)
          ? payloadObject.data
          : Array.isArray(nestedData.posts)
            ? nestedData.posts
            : Array.isArray(nestedData.PostUserData)
              ? nestedData.PostUserData
              : Array.isArray(nestedData.newsfeed)
                ? nestedData.newsfeed
                : Array.isArray(nestedData.items)
                  ? nestedData.items
                  : [];

  return normalizeFeedPosts(feedData);
}

export async function toggleLikePost(postId: string, currentLiked: boolean): Promise<void> {
  const endpoint = `${API_BASE_URL}/user/post/${postId}/${currentLiked ? 'unlike' : 'like'}`;
  const response = await fetch(endpoint, {
    method: currentLiked ? 'DELETE' : 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}));
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to update like status.');
  }
}

export async function fetchCommentsForPost(postId: string): Promise<FeedComment[]> {
  const response = await fetch(`${API_BASE_URL}/user/post/${postId}/comment?limit=20&offset=0`, {
    method: 'GET',
    credentials: 'include',
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to load comments.');
  }

  const data = (payload && typeof payload === 'object' ? (payload as Record<string, unknown>).data : undefined) ?? payload;
  const commentList = Array.isArray((data as Record<string, unknown>)?.Comments)
    ? ((data as Record<string, unknown>).Comments as unknown[])
    : Array.isArray((data as Record<string, unknown>)?.comments)
      ? ((data as Record<string, unknown>).comments as unknown[])
      : Array.isArray((payload as Record<string, unknown>)?.Comments)
        ? ((payload as Record<string, unknown>).Comments as unknown[])
        : Array.isArray((payload as Record<string, unknown>)?.comments)
          ? ((payload as Record<string, unknown>).comments as unknown[])
          : [];

  const normalizedComments: FeedComment[] = commentList.map((commentValue, index: number) => {
    const comment = commentValue as Record<string, unknown>;
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

  return normalizedComments;
}

export async function addCommentToPost(postId: string, commentText: string, parentCommentId?: string): Promise<void> {
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

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to add your comment.');
  }
}

export async function deletePostById(postId: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/user/post/${postId}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to delete this post.');
  }
}

export async function updatePostContent(postId: string, content: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/user/post/${postId}`, {
    method: 'PATCH',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ caption: content.trim() }),
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to update your post.');
  }
}

export async function createPostWithMedia(
  content: string,
  files: File[],
  authorInfo?: { author?: string; handle?: string; profileImageUrl?: string },
): Promise<FeedPost> {
  const formData = new FormData();
  if (content.trim()) {
    formData.append('caption', content.trim());
  }
  files.forEach((file) => {
    formData.append('media', file);
  });

  const response = await fetch(`${API_BASE_URL}/user/post`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to create your post.');
  }

  const createdPostId = getPostIdFromResponse(payload);
  if (!createdPostId) {
    throw new Error('The server did not return a valid post id.');
  }

  return {
    id: createdPostId,
    author: authorInfo?.author || 'You',
    handle: authorInfo?.handle || '@you',
    time: 'just now',
    content: content.trim(),
    imageUrls: files.map((file) => URL.createObjectURL(file)),
    profileImageUrl: authorInfo?.profileImageUrl || undefined,
    isLiked: false,
    likes: 0,
    comments: 0,
    shares: 0,
    accent: 'from-sky-400 to-blue-600',
    tag: 'New',
  };
}
