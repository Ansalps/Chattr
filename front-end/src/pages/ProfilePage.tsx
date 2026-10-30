import { useMemo, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';
import { ProfileHeader } from '../features/profile/ProfileHeader';
import { ProfilePostGrid } from '../features/profile/ProfilePostGrid';
import { ProfilePostModal } from '../features/profile/ProfilePostModal';
import type { ProfilePost } from '../features/profile/profile.types';
import { dedupeCustomLinks } from '../features/profile/profile.utils';
import { fetchFollowersList, fetchFollowingList, type SocialConnectionUser } from '../features/profile/profileService';
import { useProfileData } from '../features/profile/useProfileData';
import { usePostInteractions } from '../features/profile/usePostInteractions';

export default function ProfilePage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { userId } = useParams();
  const {
    posts,
    setPosts,
    user,
    isLoading,
    isUploadingAvatar,
    followerCount,
    followingCount,
    isFollowing,
    isUpdatingFollowState,
    error,
    avatarInputRef,
    isOwnProfile,
    handleFollowToggle,
    handleAvatarUpload,
  } = useProfileData(userId);

  const [selectedPost, setSelectedPost] = useState<ProfilePost | null>(null);
  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [relationDialog, setRelationDialog] = useState<{ type: 'followers' | 'following' | null; users: SocialConnectionUser[]; loading: boolean }>({
    type: null,
    users: [],
    loading: false,
  });

  const {
    commentDrafts,
    setCommentDrafts,
    replyDrafts,
    setReplyDrafts,
    replyingToComment,
    setReplyingToComment,
    expandedRepliesByComment,
    setExpandedRepliesByComment,
    commentsByPost,
    commentLoading,
    showComments,
    setShowComments,
    fetchCommentsForPost,
    handleToggleLike,
    handleAddComment,
  } = usePostInteractions({
    setPosts,
    selectedPost,
    setSelectedPost,
  });

  const openRelationDialog = async (type: 'followers' | 'following') => {
    setRelationDialog({ type, users: [], loading: true });

    try {
      const result = type === 'followers' ? await fetchFollowersList(user.id === 'me' ? undefined : user.id) : await fetchFollowingList(user.id === 'me' ? undefined : user.id);
      setRelationDialog({ type, users: result, loading: false });
    } catch (error) {
      console.error(`fetch ${type} failed`, error);
      setRelationDialog({ type, users: [], loading: false });
    }
  };

  const summary = useMemo(
    () => [
      { label: 'Posts', value: String(posts.length) },
      {
        label: 'Followers',
        value: String(followerCount),
        onClick: () => {
          void openRelationDialog('followers');
        },
      },
      {
        label: 'Following',
        value: String(followingCount),
        onClick: () => {
          void openRelationDialog('following');
        },
      },
    ],
    [posts.length, followerCount, followingCount, openRelationDialog],
  );

  const dedupedCustomLinks = useMemo(() => dedupeCustomLinks(user), [user]);

  const handleMessageUser = (targetUser: SocialConnectionUser) => {
    setRelationDialog({ type: null, users: [], loading: false });
    const params = new URLSearchParams({
      chatUserId: targetUser.id,
      chatUserName: targetUser.name,
      chatUserUsername: targetUser.username,
      chatUserImage: targetUser.profileImageUrl ?? '',
    });

    navigate({ pathname: location.pathname, search: params.toString() });
  };

  const handleOpenPost = async (post: ProfilePost, imageIndex = 0) => {
    setSelectedPost(post);
    setSelectedImageIndex(imageIndex);
    setShowComments(false);
    await fetchCommentsForPost(post.id);
  };

  const handleCloseModal = () => {
    setSelectedPost(null);
    setSelectedImageIndex(0);
    setShowComments(false);
  };

  return (
    <div className="min-h-screen bg-[#f3f7fb] px-4 py-8 text-slate-800 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <ProfileHeader
          user={user}
          isOwnProfile={isOwnProfile}
          isFollowing={isFollowing}
          isUpdatingFollowState={isUpdatingFollowState}
          isUploadingAvatar={isUploadingAvatar}
          avatarInputRef={avatarInputRef}
          summary={summary}
          dedupedCustomLinks={dedupedCustomLinks}
          onBack={() => navigate('/newsfeed')}
          onGoEdit={() => navigate('/profile/edit')}
          onGoSubscription={() => navigate('/subscription')}
          onFollowToggle={handleFollowToggle}
          onAvatarUpload={handleAvatarUpload}
        />

        <ProfilePostGrid
          posts={posts}
          isLoading={isLoading}
          error={error}
          onOpenPost={handleOpenPost}
        />
      </div>

      <ProfilePostModal
        user={user}
        selectedPost={selectedPost}
        selectedImageIndex={selectedImageIndex}
        setSelectedImageIndex={setSelectedImageIndex}
        onClose={handleCloseModal}
        onToggleLike={handleToggleLike}
        onOpenComments={async (postId: string) => {
          if (showComments) {
            setShowComments(false);
            return;
          }

          setShowComments(true);
          await fetchCommentsForPost(postId);
        }}
        onAddComment={handleAddComment}
        showComments={showComments}
        commentsByPost={commentsByPost}
        commentDrafts={commentDrafts}
        setCommentDrafts={setCommentDrafts}
        replyDrafts={replyDrafts}
        setReplyDrafts={setReplyDrafts}
        replyingToComment={replyingToComment}
        setReplyingToComment={setReplyingToComment}
        expandedRepliesByComment={expandedRepliesByComment}
        setExpandedRepliesByComment={setExpandedRepliesByComment}
        commentLoading={commentLoading}
      />

      {relationDialog.type && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4">
          <div className="w-full max-w-lg overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-[0_20px_60px_rgba(15,23,42,0.25)]">
            <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
              <div>
                <p className="text-xs font-semibold uppercase tracking-[0.2em] text-sky-600">People</p>
                <h3 className="mt-1 text-xl font-black text-slate-900">
                  {relationDialog.type === 'followers' ? 'Followers' : 'Following'}
                </h3>
              </div>

              <button
                type="button"
                onClick={() => setRelationDialog({ type: null, users: [], loading: false })}
                className="rounded-full border border-slate-200 bg-white px-3 py-1.5 text-sm font-medium text-slate-600 transition hover:bg-slate-50"
              >
                Close
              </button>
            </div>

            <div className="max-h-[70vh] overflow-y-auto p-4">
              {relationDialog.loading ? (
                <div className="py-8 text-center text-sm text-slate-500">Loading…</div>
              ) : relationDialog.users.length === 0 ? (
                <div className="py-8 text-center text-sm text-slate-500">
                  No {relationDialog.type} yet.
                </div>
              ) : (
                <div className="space-y-3">
                  {relationDialog.users.map((person) => (
                    <div key={person.id} className="flex items-center justify-between gap-3 rounded-2xl border border-sky-100 bg-sky-50/40 p-3">
                      <div className="flex min-w-0 items-center gap-3">
                        <div className="flex h-11 w-11 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-sky-500 to-indigo-500 text-xs font-bold text-white">
                          {person.profileImageUrl ? (
                            <img src={person.profileImageUrl} alt={person.name} className="h-full w-full object-cover" />
                          ) : (
                            <span>{person.name.slice(0, 2).toUpperCase() || 'U'}</span>
                          )}
                        </div>

                        <div className="min-w-0">
                          <div className="truncate text-sm font-semibold text-slate-900">{person.name}</div>
                          <div className="truncate text-xs text-slate-500">{person.username}</div>
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={() => handleMessageUser(person)}
                        className="rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 px-3 py-2 text-xs font-semibold text-white shadow-sm transition hover:brightness-105"
                      >
                        Message
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
