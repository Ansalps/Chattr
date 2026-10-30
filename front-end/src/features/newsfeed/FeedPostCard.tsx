import type { Dispatch, SetStateAction } from 'react';
import type { FeedComment, FeedPost } from '../../shared/types';
import { formatCount } from '../../shared/utils/formatting';

const getInitials = (value: string) =>
  value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'A';

export type FeedPostCardProps = {
  post: FeedPost;
  currentUserId: string;
  currentUserHandle: string;
  currentUserName: string;
  currentProfileImageUrl: string;
  postMenuOpenFor: string | null;
  setPostMenuOpenFor: Dispatch<SetStateAction<string | null>>;
  openCommentFor: string | null;
  setOpenCommentFor: Dispatch<SetStateAction<string | null>>;
  onGoToProfile: (userId?: string) => void;
  onToggleLike: (postId: string, currentLiked: boolean) => void;
  onOpenComments: (postId: string) => void;
  onEditPost: (post: FeedPost) => void;
  onDeletePost: (postId: string) => void;
  commentDrafts: Record<string, string>;
  setCommentDrafts: Dispatch<SetStateAction<Record<string, string>>>;
  commentLoading: Record<string, boolean>;
  commentsByPost: Record<string, FeedComment[]>;
  replyDrafts: Record<string, string>;
  setReplyDrafts: Dispatch<SetStateAction<Record<string, string>>>;
  replyingToComment: Record<string, string | null>;
  setReplyingToComment: Dispatch<SetStateAction<Record<string, string | null>>>;
  expandedRepliesByComment: Record<string, boolean>;
  setExpandedRepliesByComment: Dispatch<SetStateAction<Record<string, boolean>>>;
  onAddComment: (postId: string, parentCommentId?: string) => void;
};

export function FeedPostCard({
  post,
  currentUserId,
  currentUserHandle,
  currentUserName,
  currentProfileImageUrl,
  postMenuOpenFor,
  setPostMenuOpenFor,
  openCommentFor,
  setOpenCommentFor,
  onGoToProfile,
  onToggleLike,
  onOpenComments,
  onEditPost,
  onDeletePost,
  commentDrafts,
  setCommentDrafts,
  commentLoading,
  commentsByPost,
  replyDrafts,
  setReplyDrafts,
  replyingToComment,
  setReplyingToComment,
  expandedRepliesByComment,
  setExpandedRepliesByComment,
  onAddComment,
}: FeedPostCardProps) {
  const normalizedPostUserId = String(post.userId ?? '').trim();
  const normalizedCurrentHandle = currentUserHandle.toLowerCase();
  const normalizedPostHandle = post.handle.toLowerCase();
  const nestedUserIdFromPayload = String(
    (post as FeedPost & { UserData?: { UserID?: string | number } }).UserData?.UserID ?? '',
  ).trim();

  const isOwnPost =
    !!currentUserId &&
    (normalizedPostUserId === currentUserId ||
      nestedUserIdFromPayload === currentUserId ||
      (!normalizedPostUserId &&
        (post.author === currentUserName ||
          normalizedPostHandle === normalizedCurrentHandle ||
          post.profileImageUrl === currentProfileImageUrl)));

  return (
    <article key={post.id} className="rounded-[28px] border border-sky-100 bg-white p-5 shadow-sm">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <button
            type="button"
            onClick={() => onGoToProfile(post.userId)}
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
                onClick={() => onGoToProfile(post.userId)}
                className="cursor-pointer text-left font-bold text-slate-900 transition hover:text-sky-700"
                aria-label={`View ${post.author}'s profile`}
              >
                {post.author}
              </button>
              {post.blueTick && <span title="Verified account" aria-label="Verified account" className="inline-flex h-4 w-4 shrink-0 items-center justify-center rounded-full bg-sky-500 text-[10px] font-black leading-none text-white">✓</span>}
              <button
                type="button"
                onClick={() => onGoToProfile(post.userId)}
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

        {isOwnPost && (
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
                  onClick={() => onEditPost(post)}
                  className="block w-full px-3 py-2 text-left text-sm text-slate-700 transition hover:bg-sky-50 hover:text-sky-700"
                >
                  Edit
                </button>
                <button
                  type="button"
                  onClick={() => onDeletePost(post.id)}
                  className="block w-full px-3 py-2 text-left text-sm text-red-600 transition hover:bg-red-50"
                >
                  Delete
                </button>
              </div>
            )}
          </div>
        )}
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
            onClick={() => onToggleLike(post.id, post.isLiked)}
            className={`flex cursor-pointer items-center gap-2 rounded-full px-3 py-1.5 transition ${post.isLiked ? 'bg-rose-50 text-rose-600' : 'bg-slate-50 text-slate-600'} `}
          >
            <span>{post.isLiked ? '♥' : '♡'}</span>
            <span>{formatCount(post.likes)}</span>
          </button>
          <button
            type="button"
            onClick={() => onOpenComments(post.id)}
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
                onClick={() => onAddComment(post.id)}
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
                            onClick={() => onAddComment(post.id, comment.id)}
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
  );
}
