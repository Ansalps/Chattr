import type { ChangeEvent, RefObject } from 'react';
import type { ProfileUser } from './profile.types';
import { getInitials } from './profile.utils';

type ProfileHeaderProps = {
  user: ProfileUser;
  isOwnProfile: boolean;
  isFollowing: boolean;
  isUpdatingFollowState: boolean;
  isUploadingAvatar: boolean;
  avatarInputRef: RefObject<HTMLInputElement | null>;
  summary: Array<{ label: string; value: string; onClick?: () => void }>;
  dedupedCustomLinks: Array<{ id: string; label: string; url: string }>;
  onBack: () => void;
  onGoEdit: () => void;
  onGoSubscription?: () => void;
  onFollowToggle: () => void;
  onAvatarUpload: (event: ChangeEvent<HTMLInputElement>) => Promise<void>;
};

export function ProfileHeader({
  user,
  isOwnProfile,
  isFollowing,
  isUpdatingFollowState,
  isUploadingAvatar,
  avatarInputRef,
  summary,
  dedupedCustomLinks,
  onBack,
  onGoEdit,
  onGoSubscription,
  onFollowToggle,
  onAvatarUpload,
}: ProfileHeaderProps) {
  return (
    <div className="mb-8 space-y-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <button
          type="button"
          onClick={onBack}
          className="inline-flex w-fit items-center justify-center rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-600 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
        >
          ← Back to newsfeed
        </button>

        <div className="flex items-center gap-2 self-end sm:self-auto">
          {!isOwnProfile && (
            <>
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
                onClick={onFollowToggle}
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
            </>
          )}

          {isOwnProfile && (
            <button type="button" onClick={onGoSubscription} className="rounded-2xl bg-sky-600 px-4 py-2 text-sm font-semibold text-white shadow-sm">Get verified</button>
          )}
          {isOwnProfile && (
            <button
              type="button"
              onClick={onGoEdit}
              className="rounded-2xl border border-sky-200 bg-sky-50 px-4 py-2 text-sm font-semibold text-sky-700 shadow-sm transition hover:border-sky-300 hover:bg-sky-100"
            >
              Edit Profile
            </button>
          )}

          <button
            type="button"
            onClick={onBack}
            className="rounded-2xl bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-sky-300/40 transition hover:brightness-105"
          >
            Newsfeed
          </button>
        </div>
      </div>

      <div className="overflow-hidden rounded-[30px] border border-sky-100 bg-white shadow-[0_18px_50px_rgba(14,116,144,0.08)]">
        <div className="bg-gradient-to-r from-sky-50 via-white to-blue-50 p-5 sm:p-7">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-start gap-4">
              <div className="relative">
                <div
                  onClick={() => {
                    if (isOwnProfile && !isUploadingAvatar) {
                      avatarInputRef.current?.click();
                    }
                  }}
                  className={`group relative flex h-24 w-24 items-center justify-center overflow-hidden rounded-[28px] bg-gradient-to-br from-sky-500 to-blue-600 text-2xl font-black text-white shadow-lg shadow-sky-300/40 transition ${
                    isOwnProfile ? 'cursor-pointer' : 'cursor-default'
                  } ${isUploadingAvatar ? 'opacity-80' : ''}`}
                  title={isOwnProfile ? 'Change profile picture' : 'Profile avatar'}
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
                    onChange={onAvatarUpload}
                  />
                )}
              </div>

              <div className="pt-1">
                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-sky-600">Profile</p>
                <div className="mt-2 flex items-center gap-2"><h1 className="text-3xl font-black tracking-tight text-slate-900">{user.name}</h1>{user.blueTick && <span title="Verified account" className="flex h-5 w-5 items-center justify-center rounded-full bg-sky-500 text-[11px] font-black text-white">✓</span>}</div>
                <p className="text-sm text-slate-500">{user.username}</p>
                {user.bio && <p className="mt-2 max-w-xl text-sm leading-6 text-slate-600">{user.bio}</p>}
              </div>
            </div>

            <div className="grid w-full max-w-xs grid-cols-3 gap-3">
              {summary.map((item) => (
                <button
                  key={item.label}
                  type="button"
                  onClick={item.onClick}
                  className="rounded-2xl border border-sky-100 bg-white px-3 py-3 text-center shadow-sm transition hover:border-sky-200 hover:bg-sky-50 disabled:cursor-default disabled:hover:bg-white"
                  disabled={!item.onClick}
                >
                  <div className="text-xl font-black text-slate-900">{item.value}</div>
                  <div className="text-[11px] uppercase tracking-[0.14em] text-slate-500">{item.label}</div>
                </button>
              ))}
            </div>
          </div>
        </div>

        {(user.website || user.linkedin || user.instagram || (user.customLinks && user.customLinks.length > 0)) && (
          <div className="border-t border-sky-100 bg-white px-5 py-4 sm:px-7">
            <div className="flex flex-wrap gap-2 text-xs">
              {user.website && (
                <a
                  href={user.website}
                  target="_blank"
                  rel="noreferrer"
                  className="rounded-full bg-sky-50 px-2.5 py-1.5 font-medium text-sky-700 transition hover:bg-sky-100"
                >
                  Website
                </a>
              )}
              {user.linkedin && (
                <a
                  href={user.linkedin}
                  target="_blank"
                  rel="noreferrer"
                  className="rounded-full bg-sky-50 px-2.5 py-1.5 font-medium text-sky-700 transition hover:bg-sky-100"
                >
                  LinkedIn
                </a>
              )}
              {user.instagram && (
                <a
                  href={user.instagram}
                  target="_blank"
                  rel="noreferrer"
                  className="rounded-full bg-sky-50 px-2.5 py-1.5 font-medium text-sky-700 transition hover:bg-sky-100"
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
                  className="rounded-full bg-sky-50 px-2.5 py-1.5 font-medium text-sky-700 transition hover:bg-sky-100"
                >
                  {link.label || link.url}
                </a>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
