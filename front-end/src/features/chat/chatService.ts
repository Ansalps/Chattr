import { normalizeUsername } from '../../shared/utils/formatting';
import type { ChatSession, ChatUserProfile } from '../../shared/types';

const API_BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:3000';

export type ChatMessage = {
  id: string;
  senderId: string;
  senderName: string;
  senderProfileImageUrl: string;
  content: string;
  createdAt: string;
  mine: boolean;
};

export type SearchUserResult = {
  id: string;
  name: string;
  username: string;
  profileImageUrl: string;
};

export async function searchUsersForSelection(query: string, currentUserId: string): Promise<SearchUserResult[]> {
  const results = await searchUsersByUsername(query);
  return results.filter((user) => user.id !== currentUserId);
}

export type BuildChatMessageInput = {
  id: string;
  conversationId?: string;
  groupId?: string;
  isGroup?: boolean;
};

export function buildChatMessagePayload(
  chatUser: BuildChatMessageInput,
  content: string,
): Record<string, string | number> {
  const isGroupChat = Boolean(chatUser.isGroup || chatUser.groupId);

  return isGroupChat
    ? {
        type: 'group',
        group_id: chatUser.groupId || chatUser.conversationId || chatUser.id,
        content,
        created_at: new Date().toISOString(),
      }
    : {
        type: 'individual',
        recipient_id: Number(chatUser.id),
        content,
        created_at: new Date().toISOString(),
      };
}

export function connectChatSocket({
  currentUserId,
  onMessage,
  onError,
}: {
  currentUserId: string;
  onMessage: (message: ChatMessage) => void;
  onError?: (message: string) => void;
}): WebSocket {
  const socketUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.hostname}:3000/user/ws`;
  const socket = new WebSocket(socketUrl);

  socket.onopen = () => {
    onError?.('');
  };

  socket.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data as string);

      if (payload && payload.error) {
        onError?.(String(payload.error));
        return;
      }

      const senderId = String(payload.sender_id ?? payload.SenderID ?? payload.senderId ?? '');
      const content = String(payload.content ?? payload.Content ?? '');
      if (!senderId && !content) {
        return;
      }

      const senderName = String(
        payload.sender_name ??
          payload.SenderName ??
          payload.senderName ??
          payload.sender_username ??
          payload.senderUsername ??
          'User',
      );
      const senderProfileImageUrl =
        (typeof payload.sender_profile === 'string' && payload.sender_profile) ||
        (typeof payload.sender_profile_url === 'string' && payload.sender_profile_url) ||
        (typeof payload.SenderProfileImgUrl === 'string' && payload.SenderProfileImgUrl) ||
        (typeof payload.senderProfileImageUrl === 'string' && payload.senderProfileImageUrl) ||
        (typeof payload.profile_image_url === 'string' && payload.profile_image_url) ||
        (typeof payload.ProfileImgUrl === 'string' && payload.ProfileImgUrl) ||
        '';

      onMessage({
        id: String(payload.message_id ?? payload.MessageID ?? payload.id ?? `${senderId}-${content}-${Date.now()}`),
        senderId,
        senderName,
        senderProfileImageUrl,
        content,
        createdAt: String(payload.created_at ?? payload.CreatedAt ?? payload.createdAt ?? new Date().toISOString()),
        mine: senderId === currentUserId,
      });
    } catch (error) {
      console.error('failed to parse chat websocket payload', error);
    }
  };

  socket.onerror = () => {
    onError?.('Unable to connect to chat right now.');
  };

  return socket;
}

export async function fetchRecentChatProfiles(): Promise<ChatSession[]> {
  const response = await fetch(`${API_BASE_URL}/user/get-recent-chat-profiles?page=1&limit=20`, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  if (!response.ok) {
    throw new Error('Unable to load recent chats.');
  }

  const payload = await response.json().catch(() => ({}));
  const data =
    (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
    (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
    payload;

  const chatProfiles = Array.isArray((data as Record<string, unknown>)?.ChatProfiles)
    ? ((data as Record<string, unknown>).ChatProfiles as unknown[])
    : Array.isArray((data as Record<string, unknown>)?.chatProfiles)
      ? ((data as Record<string, unknown>).chatProfiles as unknown[])
      : Array.isArray((data as Record<string, unknown>)?.data)
        ? ((data as Record<string, unknown>).data as unknown[])
        : [];

  return chatProfiles
    .map((conversation) => {
      const source = (conversation ?? {}) as Record<string, unknown>;
      const conversationId = String(
        source.chat_id ?? source.ChatID ?? source.chatId ?? source.ConversationID ?? source.conversation_id ?? '',
      );

      const rawUsername =
        (typeof source.user_name === 'string' && source.user_name) ||
        (typeof source.UserName === 'string' && source.UserName) ||
        (typeof source.userName === 'string' && source.userName) ||
        (typeof source.username === 'string' && source.username) ||
        (typeof source.Username === 'string' && source.Username) ||
        (typeof source.chat_name === 'string' && source.chat_name) ||
        (typeof source.ChatName === 'string' && source.ChatName) ||
        (typeof source.chatName === 'string' && source.chatName) ||
        (typeof source.name === 'string' && source.name) ||
        (typeof source.Name === 'string' && source.Name) ||
        '@user';

      const name = String(
        (typeof source.name === 'string' && source.name) ||
          (typeof source.Name === 'string' && source.Name) ||
          (typeof source.chat_name === 'string' && source.chat_name) ||
          (typeof source.ChatName === 'string' && source.ChatName) ||
          (typeof source.chatName === 'string' && source.chatName) ||
          rawUsername.replace(/^@/, '') ||
          'User',
      );
      const username = normalizeUsername(String(rawUsername));
      const profileImageUrl =
        (typeof source.chat_image === 'string' && source.chat_image) ||
        (typeof source.ChatImage === 'string' && source.ChatImage) ||
        (typeof source.chatImage === 'string' && source.chatImage) ||
        (typeof source.group_image_url === 'string' && source.group_image_url) ||
        (typeof source.groupImageUrl === 'string' && source.groupImageUrl) ||
        (typeof source.groupimageurl === 'string' && source.groupimageurl) ||
        (typeof source.GroupImageURL === 'string' && source.GroupImageURL) ||
        (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
        (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
        '';
      const lastMessage = String(source.last_message ?? source.LastMessage ?? '');
      const lastMessageTime = String(
        source.last_message_time ?? source.LastMessageTime ?? source.lastMessageTime ?? new Date().toISOString(),
      );
      const isGroup = Boolean(
        source.is_group || source.IsGroup || source.isGroup || source.type === 'group' || source.Type === 'group',
      );
      const groupId = String(source.group_id ?? source.GroupID ?? source.groupId ?? '');

      if (!conversationId && !name) return null;

      return {
        id: String(source.user_id ?? source.UserID ?? source.userId ?? source.UserId ?? name ?? conversationId),
        conversationId,
        name,
        username,
        profileImageUrl,
        lastMessage,
        lastMessageTime,
        isGroup,
        groupId,
      };
    })
    .filter(Boolean) as ChatSession[];
}

export async function searchUsersByUsername(username: string): Promise<SearchUserResult[]> {
  const trimmed = username.trim();
  if (!trimmed) return [];

  const response = await fetch(`${API_BASE_URL}/user/search?username=${encodeURIComponent(trimmed)}&limit=5&offset=0`, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  const payload = await response.json().catch(() => ({}));

  if (!response.ok || (payload && typeof payload === 'object' && 'status_code' in payload && Number((payload as Record<string, unknown>).status_code) >= 400)) {
    throw new Error((payload as Record<string, unknown>).message ? String((payload as Record<string, unknown>).message) : 'Unable to search users.');
  }

  const dataPayload =
    (payload && typeof payload === 'object' && 'data' in payload ? (payload as Record<string, unknown>).data : undefined) ||
    (payload && typeof payload === 'object' && 'response' in payload ? (payload as Record<string, unknown>).response : undefined) ||
    payload;

  const userList: unknown[] =
    Array.isArray((dataPayload as Record<string, unknown>)?.UserMeataData)
      ? ((dataPayload as Record<string, unknown>).UserMeataData as unknown[])
      : Array.isArray((dataPayload as Record<string, unknown>)?.userMeataData)
        ? ((dataPayload as Record<string, unknown>).userMeataData as unknown[])
        : Array.isArray((dataPayload as Record<string, unknown>)?.data)
          ? ((dataPayload as Record<string, unknown>).data as unknown[])
          : Array.isArray((dataPayload as Record<string, unknown>)?.Data)
            ? ((dataPayload as Record<string, unknown>).Data as unknown[])
            : [];

  return userList
    .map((user) => {
      const source = (user ?? {}) as Record<string, unknown>;
      const userId = String(
        source.UserId ?? source.userId ?? source.user_id ?? source.UserID ?? source.userID ?? '',
      );
      const name = String(
        source.Name ?? source.name ?? source.FullName ?? source.full_name ?? source.UserName ?? source.userName ?? 'User',
      );
      const username = String(source.UserName ?? source.userName ?? source.username ?? '@user');
      const profileImageUrl =
        (typeof source.ProfileImgUrl === 'string' && source.ProfileImgUrl) ||
        (typeof source.profile_image_url === 'string' && source.profile_image_url) ||
        (typeof source.profileImageUrl === 'string' && source.profileImageUrl) ||
        '';

      if (!userId && !name && !username) return null;

      return {
        id: userId || name || username,
        name,
        username: normalizeUsername(username),
        profileImageUrl,
      };
    })
    .filter(Boolean) as SearchUserResult[];
}

export async function fetchChatHistory(conversationId: string, currentUserId: string): Promise<ChatMessage[]> {
  const response = await fetch(`${API_BASE_URL}/user/chat/${encodeURIComponent(conversationId)}?page=1&limit=50`, {
    method: 'GET',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  });

  if (!response.ok) {
    throw new Error('Unable to load chat history.');
  }

  const historyPayload = await response.json().catch(() => ({}));
  const historyData =
    (historyPayload && typeof historyPayload === 'object' && 'data' in historyPayload
      ? (historyPayload as Record<string, unknown>).data
      : historyPayload) ?? {};

  const historyMessages = Array.isArray((historyData as Record<string, unknown>)?.messages)
    ? ((historyData as Record<string, unknown>).messages as unknown[])
    : Array.isArray((historyData as Record<string, unknown>)?.Messages)
      ? ((historyData as Record<string, unknown>).Messages as unknown[])
      : [];

  return historyMessages.map((message, index) => {
    const item = (message ?? {}) as Record<string, unknown>;
    const senderId = String(item.sender_id ?? item.SenderID ?? item.senderId ?? '');
    const senderName = String(
      item.sender_name ?? item.SenderName ?? item.senderName ?? item.sender_username ?? item.senderUsername ?? 'User',
    );
    const senderProfileImageUrl =
      (typeof item.sender_profile === 'string' && item.sender_profile) ||
      (typeof item.SenderProfileImgUrl === 'string' && item.SenderProfileImgUrl) ||
      (typeof item.senderProfileImageUrl === 'string' && item.senderProfileImageUrl) ||
      (typeof item.profile_image_url === 'string' && item.profile_image_url) ||
      (typeof item.ProfileImgUrl === 'string' && item.ProfileImgUrl) ||
      '';
    const content = String(item.content ?? item.Content ?? '');
    const createdAt = String(item.created_at ?? item.CreatedAt ?? item.createdAt ?? new Date().toISOString());

    return {
      id: String(item.message_id ?? item.MessageID ?? item.id ?? `${senderId}-${createdAt}-${index}`),
      senderId,
      senderName,
      senderProfileImageUrl,
      content,
      createdAt,
      mine: senderId === currentUserId,
    };
  });
}

export async function loadChatSessionData({
  conversationId,
  currentUserId,
  groupId,
  isGroup,
}: {
  conversationId?: string;
  currentUserId: string;
  groupId?: string;
  isGroup?: boolean;
}): Promise<{ messages: ChatMessage[]; groupMembers: ChatUserProfile[] }> {
  if (!conversationId) {
    return { messages: [], groupMembers: [] };
  }

  const messages = await fetchChatHistory(conversationId, currentUserId);
  let groupMembers: ChatUserProfile[] = [];

  if (isGroup && groupId) {
    groupMembers = await fetchGroupMembers(groupId);
  }

  return {
    messages: messages.filter((message) => message.content.trim()),
    groupMembers,
  };
}

export async function fetchGroupMembers(groupId: string): Promise<ChatUserProfile[]> {
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

  return memberList.map((member) => {
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
  });
}
