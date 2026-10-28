import { useCallback, useEffect, useRef, useState, type ChangeEvent } from 'react';
import { fetchProfilePageData } from './profileService';
import type { ProfilePost, ProfileUser } from './profile.types';
import { PROFILE_STORAGE_KEY, readStoredProfile } from './profile.utils';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

const EMPTY_USER: ProfileUser = {
  id: 'me',
  name: 'User Profile',
  username: '@user',
};

export function useProfileData(userId?: string) {
  const [posts, setPosts] = useState<ProfilePost[]>([]);
  const [user, setUser] = useState<ProfileUser>(EMPTY_USER);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploadingAvatar, setIsUploadingAvatar] = useState(false);
  const [followerCount, setFollowerCount] = useState(0);
  const [followingCount, setFollowingCount] = useState(0);
  const [isFollowing, setIsFollowing] = useState(false);
  const [isUpdatingFollowState, setIsUpdatingFollowState] = useState(false);
  const [error, setError] = useState('');
  const avatarInputRef = useRef<HTMLInputElement | null>(null);
  const storedProfile = readStoredProfile();
  const currentUserId = String(storedProfile.id ?? '').trim();
  const normalizeProfileId = (value?: string) => String(value ?? '').trim().replace(/^@/, '').toLowerCase();

  const isOwnProfile = 
    !userId ||
    userId === 'me' ||
    normalizeProfileId(userId) === normalizeProfileId(currentUserId) ||
    normalizeProfileId(user.id) === normalizeProfileId(currentUserId);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const nextProfileData = await fetchProfilePageData(userId);
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

    void fetchProfile();
  }, [userId]);

  const handleFollowToggle = useCallback(async () => {
    if (!userId || userId === 'me' || isOwnProfile) {
      return;
    }

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

      if (!response.ok || Number((payload as Record<string, unknown>).status_code ?? 0) >= 400) {
        throw new Error(
          ((payload as Record<string, unknown>).message as string) ||
            ((payload as Record<string, unknown>).error as string) ||
            `Unable to ${isFollowing ? 'unfollow' : 'follow'} this user.`,
        );
      }

      setIsFollowing((current) => !current);
      setFollowerCount((current) => Math.max(0, current + (isFollowing ? -1 : 1)));
    } catch (followError) {
      const message = followError instanceof Error ? followError.message : 'Unable to update follow status.';
      setError(message);
    } finally {
      setIsUpdatingFollowState(false);
    }
  }, [isFollowing, userId]);

  const handleAvatarUpload = useCallback(
    async (event: ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0];
      if (!file) {
        return;
      }

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
        const apiPayload = payload as Record<string, unknown>;

        if (!response.ok || Number(apiPayload.status_code ?? 0) >= 400) {
          throw new Error(
            (typeof apiPayload.message === 'string' ? apiPayload.message : '') ||
              (typeof apiPayload.error === 'string' ? apiPayload.error : '') ||
              'Unable to upload profile image.',
          );
        }

        const uploadedUrl =
          (typeof apiPayload.ImageUrl === 'string' ? apiPayload.ImageUrl : '') ||
          (typeof apiPayload.image_url === 'string' ? apiPayload.image_url : '') ||
          (typeof apiPayload.imageUrl === 'string' ? apiPayload.imageUrl : '');

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
    },
    [user.name, user.username],
  );

  return {
    posts,
    setPosts,
    user,
    setUser,
    isLoading,
    isUploadingAvatar,
    followerCount,
    setFollowerCount,
    followingCount,
    setFollowingCount,
    isFollowing,
    setIsFollowing,
    isUpdatingFollowState,
    error,
    setError,
    avatarInputRef,
    isOwnProfile,
    handleFollowToggle,
    handleAvatarUpload,
  };
}
