import { normalizeUsername } from '../../shared/utils/formatting';
import type { ChatUserProfile } from '../../shared/types';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export type GroupMembersResult = {
  members: ChatUserProfile[];
  creatorId: string;
};

function normalizeGroupMember(member: unknown): ChatUserProfile {
  const source = (member ?? {}) as Record<string, unknown>;
  const userId = String(source.UserID ?? source.userId ?? source.user_id ?? source.UserId ?? source.id ?? '');
  const rawName =
    (typeof source.Name === 'string' && source.Name) ||
    (typeof source.name === 'string' && source.name) ||
    (typeof source.FullName === 'string' && source.FullName) ||
    (typeof source.full_name === 'string' && source.full_name) ||
    '';
  const rawUsername =
    (typeof source.UserName === 'string' && source.UserName) ||
    (typeof source.userName === 'string' && source.userName) ||
    (typeof source.username === 'string' && source.username) ||
    (typeof source.user_name === 'string' && source.user_name) ||
    (typeof source.User_Name === 'string' && source.User_Name) ||
    (typeof source.User === 'string' && source.User) ||
    (typeof source.user === 'string' && source.user) ||
    '';
  const displayName = rawName || rawUsername.replace(/^@/, '') || 'User';
  const usernameValue = rawUsername || displayName;
  const profileImageUrl =
    (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
    (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
    (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
    '';

  return {
    id: userId || usernameValue || displayName,
    name: displayName,
    username: normalizeUsername(usernameValue),
    profileImageUrl,
  };
}

export async function createGroup(groupName: string, memberIds: number[]): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/user/group`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      group_name: groupName.trim(),
      group_members: memberIds,
    }),
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to create group.');
  }
}

export async function fetchGroupMembers(groupId: string): Promise<GroupMembersResult> {
  const response = await fetch(`${API_BASE_URL}/user/group/${encodeURIComponent(groupId)}/members?page=1&limit=50`, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  if (!response.ok) {
    throw new Error('Unable to load group members.');
  }

  const payload = await response.json().catch(() => ({}));
  const data =
    (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
    (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
    payload;

  const memberList: unknown[] = Array.isArray(data)
    ? data
    : Array.isArray((data as Record<string, unknown>)?.members)
      ? ((data as Record<string, unknown>).members as unknown[])
      : Array.isArray((data as Record<string, unknown>)?.Members)
        ? ((data as Record<string, unknown>).Members as unknown[])
        : Array.isArray((data as Record<string, unknown>)?.GetGroupMembers)
          ? ((data as Record<string, unknown>).GetGroupMembers as unknown[])
          : Array.isArray((data as Record<string, unknown>)?.getGroupMembers)
            ? ((data as Record<string, unknown>).getGroupMembers as unknown[])
            : [];

  const members = memberList.map(normalizeGroupMember).filter((member) => member.id);
  const creatorId =
    memberList
      .map((member) => {
        const source = (member ?? {}) as Record<string, unknown>;
        const creatorCandidate =
          source.CreatorID ??
          source.creator_id ??
          source.creatorId ??
          source.CreatedBy ??
          source.created_by ??
          source.UserID ??
          source.userId ??
          source.user_id ??
          '';
        return creatorCandidate !== '' ? String(creatorCandidate) : '';
      })
      .find((value) => value && value !== '0') || '';

  return { members, creatorId };
}

export async function addMemberToGroup(groupId: string, memberId: string): Promise<void> {
  const numericId = Number(memberId);
  if (!Number.isFinite(numericId) || numericId <= 0) {
    throw new Error('Selected member is invalid.');
  }

  const response = await fetch(`${API_BASE_URL}/user/group/${encodeURIComponent(groupId)}/add-members`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ group_members: [numericId] }),
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to add member to group.');
  }
}

export async function removeMemberFromGroup(groupId: string, memberId: string): Promise<void> {
  const numericId = Number(memberId);
  if (!Number.isFinite(numericId) || numericId <= 0) {
    throw new Error('Member ID is invalid.');
  }

  const response = await fetch(`${API_BASE_URL}/user/group/${encodeURIComponent(groupId)}/remove-member/${encodeURIComponent(memberId)}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to remove member from group.');
  }
}

export async function updateGroupProfileImage(groupId: string, file: File): Promise<string> {
  const formData = new FormData();
  formData.append('image', file);

  const response = await fetch(`${API_BASE_URL}/user/group/${encodeURIComponent(groupId)}/set-group-profile-image`, {
    method: 'PATCH',
    credentials: 'include',
    body: formData,
  });

  const payload = await response.json().catch(() => ({}));
  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to update group profile image.');
  }

  const rawPayload = payload as Record<string, unknown>;
  const data = typeof rawPayload.data === 'object' && rawPayload.data ? (rawPayload.data as Record<string, unknown>) : undefined;
  const serverImageUrl =
    (typeof rawPayload.group_profile_image_url === 'string' && rawPayload.group_profile_image_url) ||
    (typeof rawPayload.groupProfileImageUrl === 'string' && rawPayload.groupProfileImageUrl) ||
    (typeof rawPayload.image_url === 'string' && rawPayload.image_url) ||
    (typeof rawPayload.ImageURL === 'string' && rawPayload.ImageURL) ||
    (typeof data?.group_profile_image_url === 'string' && data.group_profile_image_url) ||
    (typeof data?.groupProfileImageUrl === 'string' && data.groupProfileImageUrl) ||
    (typeof data?.image_url === 'string' && data.image_url) ||
    (typeof data?.ImageURL === 'string' && data.ImageURL) ||
    '';

  if (!serverImageUrl.trim()) {
    throw new Error('The server did not return a valid group profile image URL.');
  }

  return serverImageUrl;
}
