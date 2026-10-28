import type { Dispatch, SetStateAction } from 'react';
import type { FeedComment, ProfilePost, ProfileUser } from './profile.types';
import { getInitials } from './profile.utils';

type ProfilePostModalProps = {
  user: ProfileUser;
  selectedPost: ProfilePost | null;
  selectedImageIndex: number;
  setSelectedImageIndex: Dispatch<SetStateAction<number>>;
  onClose: () => void;
  onToggleLike: (postId: string, currentLiked: boolean) => Promise<void>;
  onOpenComments: (postId: string) => Promise<void>;
  onAddComment: (postId: string, parentCommentId?: string) => Promise<void>;
  showComments: boolean;
  commentsByPost: Record<string, FeedComment[]>;
  commentDrafts: Record<string, string>;
  setCommentDrafts: Dispatch<SetStateAction<Record<string, string>>>;
  replyDrafts: Record<string, string>;
  setReplyDrafts: Dispatch<SetStateAction<Record<string, string>>>;
  replyingToComment: Record<string, string | null>;
  setReplyingToComment: Dispatch<SetStateAction<Record<string, string | null>>>;
  expandedRepliesByComment: Record<string, boolean>;
  setExpandedRepliesByComment: Dispatch<SetStateAction<Record<string, boolean>>>;
  commentLoading: Record<string, boolean>;
};

export function ProfilePostModal({
  user,
  selectedPost,
  selectedImageIndex,
  setSelectedImageIndex,
  onClose,
  onToggleLike,
  onOpenComments,
  onAddComment,
  showComments,
  commentsByPost,
  commentDrafts,
  setCommentDrafts,
  replyDrafts,
  setReplyDrafts,
  replyingToComment,
  setReplyingToComment,
  expandedRepliesByComment,
  setExpandedRepliesByComment,
  commentLoading,
}: ProfilePostModalProps) {
  if (!selectedPost) {
    return null;
  }

  const selectedComments = commentsByPost[selectedPost.id] ?? [];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 p-4">
      <div className="max-h-[90vh] w-full max-w-5xl overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-[0_25px_80px_rgba(15,23,42,0.25)]">
        <div className="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-sky-600">Full image</p>
            <h3 className="mt-1 text-lg font-bold text-slate-900">{user.name}</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
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
                    onClick={() =>
                      setSelectedImageIndex((prev) => Math.min(selectedPost.imageUrls.length - 1, prev + 1))
                    }
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
                onClick={() => void onToggleLike(selectedPost.id, Boolean(selectedPost.isLiked))}
                className={`flex cursor-pointer items-center gap-2 rounded-full px-3 py-1.5 text-sm font-medium ${
                  selectedPost.isLiked ? 'bg-rose-50 text-rose-600' : 'bg-slate-100 text-slate-600'
                }`}
              >
                <span>{selectedPost.isLiked ? '♥' : '♡'}</span>
                <span>{selectedPost.likes}</span>
              </button>

              <button
                type="button"
                onClick={() => void onOpenComments(selectedPost.id)}
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
                    onClick={() => void onAddComment(selectedPost.id)}
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
                  ) : selectedComments.length > 0 ? (
                    selectedComments.map((comment) => {
                      const replyDraftKey = `${selectedPost.id}:${comment.id}`;
                      const isReplying = replyingToComment[selectedPost.id] === comment.id;
                      const isExpanded = expandedRepliesByComment[`${selectedPost.id}:${comment.id}`] ?? false;

                      return (
                        <div key={comment.id} className="rounded-2xl bg-slate-50 px-3 py-2 ring-1 ring-slate-100">
                          <div className="mb-2 flex items-start gap-3">
                            <div className="flex h-9 w-9 items-center justify-center overflow-hidden rounded-full bg-linear-to-br from-sky-400 to-blue-600 text-[10px] font-bold text-white ring-1 ring-sky-100">
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
                                  onClick={() =>
                                    setReplyingToComment((current) => ({
                                      ...current,
                                      [selectedPost.id]: null,
                                    }))
                                  }
                                  className="cursor-pointer rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-500"
                                >
                                  Close
                                </button>
                                <button
                                  type="button"
                                  onClick={() => void onAddComment(selectedPost.id, comment.id)}
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
                                    <div className="flex h-7 w-7 items-center justify-center overflow-hidden rounded-full bg-linear-to-br from-sky-300 to-blue-500 text-[8px] font-bold text-white ring-1 ring-sky-100">
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
  );
}
