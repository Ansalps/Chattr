import { useEffect, useState } from 'react';
import type { ChangeEvent } from 'react';
import {
  buildChatMessagePayload,
  connectChatSocket,
  fetchRecentChatProfiles,
  loadChatSessionData,
  searchUsersForSelection,
} from './chatService';
import {
  addMemberToGroup,
  createGroup,
  fetchGroupMembers,
  removeMemberFromGroup,
  updateGroupProfileImage,
} from '../groups/groupService';
import type { ChatSession, ChatUserProfile } from '../../shared/types';

type StoredProfile = {
  id: string;
  name: string;
  username: string;
  profileImageUrl: string;
};

type ChatTargetUser = ChatUserProfile & {
  conversationId?: string;
  isGroup?: boolean;
  groupId?: string;
};

export type UseChatManagementOptions = {
  currentUserId: string;
  currentUserName: string;
  currentProfileImageUrl: string;
  storedProfile: StoredProfile;
};

export function useChatManagement({
  currentUserId,
  currentUserName,
  currentProfileImageUrl,
  storedProfile,
}: UseChatManagementOptions) {
  const [isChatOpen, setIsChatOpen] = useState(false);
  const [chatSearch, setChatSearch] = useState('');
  const [chatSearchResults, setChatSearchResults] = useState<ChatUserProfile[]>([]);
  const [isSearchingChatUsers, setIsSearchingChatUsers] = useState(false);
  const [chatSessions, setChatSessions] = useState<ChatSession[]>([]);
  const [isCreatingGroup, setIsCreatingGroup] = useState(false);
  const [groupNameDraft, setGroupNameDraft] = useState('');
  const [groupMemberSearch, setGroupMemberSearch] = useState('');
  const [groupMemberResults, setGroupMemberResults] = useState<ChatUserProfile[]>([]);
  const [selectedGroupMembers, setSelectedGroupMembers] = useState<ChatUserProfile[]>([]);
  const [chatSelectedUser, setChatSelectedUser] = useState<ChatTargetUser | null>(null);
  const [chatMessages, setChatMessages] = useState<
    Array<{
      id: string;
      senderId: string;
      senderName: string;
      senderProfileImageUrl: string;
      content: string;
      createdAt: string;
      mine: boolean;
    }>
  >([]);
  const [chatDraft, setChatDraft] = useState('');
  const [chatSocket, setChatSocket] = useState<WebSocket | null>(null);
  const [chatError, setChatError] = useState('');
  const [isLoadingChatHistory, setIsLoadingChatHistory] = useState(false);
  const [isGroupDetailOpen, setIsGroupDetailOpen] = useState(false);
  const [groupMembers, setGroupMembers] = useState<ChatUserProfile[]>([]);
  const [groupCreatorId, setGroupCreatorId] = useState('');

  const loadRecentChats = async () => {
    try {
      const normalizedChats = await fetchRecentChatProfiles();
      setChatSessions(normalizedChats);
    } catch (error) {
      console.error('loading recent chats failed', error);
      setChatSessions([]);
    }
  };

  const openChatSession = async (user: ChatTargetUser) => {
    setChatSelectedUser(user);
    setChatSearch('');
    setChatSearchResults([]);
    setChatDraft('');
    setChatError('');
    setChatMessages([]);
    setIsLoadingChatHistory(true);
    setIsGroupDetailOpen(false);

    try {
      const { messages, groupMembers: nextGroupMembers } = await loadChatSessionData({
        conversationId: user.conversationId,
        currentUserId: currentUserId || storedProfile.id || '',
        groupId: user.groupId,
        isGroup: user.isGroup,
      });

      setChatMessages(messages);
      setGroupMembers(nextGroupMembers.filter((member) => member.id));
    } catch (error) {
      console.error('loading chat history failed', error);
      setChatMessages([]);
    } finally {
      setIsLoadingChatHistory(false);
    }
  };

  const refreshGroupMembers = async (groupId: string) => {
    if (!groupId) {
      setGroupMembers([]);
      return;
    }

    try {
      const { members, creatorId } = await fetchGroupMembers(groupId);
      setGroupCreatorId(creatorId || currentUserId || '');
      setGroupMembers(members.filter((member) => member.id));
    } catch (error) {
      console.error('refresh group members failed', error);
      setGroupMembers([]);
    }
  };

  const handleAddMemberToGroup = async (member: ChatUserProfile) => {
    if (!chatSelectedUser?.groupId) {
      return;
    }

    try {
      await addMemberToGroup(chatSelectedUser.groupId, member.id);
      setGroupMemberSearch('');
      setGroupMemberResults([]);
      await refreshGroupMembers(chatSelectedUser.groupId);
      setChatError('');
    } catch (error) {
      console.error('add member to group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to add member to group.');
    }
  };

  const handleRemoveMemberFromGroup = async (member: ChatUserProfile) => {
    if (!chatSelectedUser?.groupId) {
      return;
    }

    try {
      await removeMemberFromGroup(chatSelectedUser.groupId, member.id);
      setGroupMembers((current) => current.filter((candidate) => candidate.id !== member.id));
      setChatError('');
    } catch (error) {
      console.error('remove member from group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to remove member from group.');
    }
  };

  const handleGroupProfileImageChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !chatSelectedUser?.groupId) {
      return;
    }

    try {
      const finalProfileImageUrl = await updateGroupProfileImage(chatSelectedUser.groupId, file);

      setChatSelectedUser((current) =>
        current ? { ...current, profileImageUrl: finalProfileImageUrl } : current,
      );
      setChatSessions((current) =>
        current.map((session) =>
          session.groupId === chatSelectedUser.groupId || session.conversationId === chatSelectedUser.groupId
            ? { ...session, profileImageUrl: finalProfileImageUrl }
            : session,
        ),
      );
      setChatError('');
    } catch (error) {
      console.error('group profile image update failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to update group profile image.');
    } finally {
      event.target.value = '';
    }
  };

  const toggleGroupMember = (member: ChatUserProfile) => {
    setSelectedGroupMembers((current) => {
      const exists = current.some((candidate) => candidate.id === member.id);
      if (exists) {
        return current.filter((candidate) => candidate.id !== member.id);
      }
      return [...current, member];
    });
  };

  const handleCreateGroup = async () => {
    const cleanedGroupName = groupNameDraft.trim();
    if (!cleanedGroupName) {
      setChatError('Please provide a group name.');
      return;
    }

    if (selectedGroupMembers.length === 0) {
      setChatError('Add at least one participant to create a group.');
      return;
    }

    const payloadMembers = selectedGroupMembers
      .map((member) => Number(member.id))
      .filter((memberId) => Number.isFinite(memberId) && memberId > 0);

    if (payloadMembers.length === 0) {
      setChatError('Selected members are invalid. Please choose valid users.');
      return;
    }

    try {
      await createGroup(cleanedGroupName, payloadMembers);

      setIsCreatingGroup(false);
      setGroupNameDraft('');
      setGroupMemberSearch('');
      setGroupMemberResults([]);
      setSelectedGroupMembers([]);
      setChatError('');
      await fetchRecentChatProfiles();
    } catch (error) {
      console.error('create group failed', error);
      setChatError(error instanceof Error ? error.message : 'Unable to create group.');
    }
  };

  const handleSendChatMessage = () => {
    if (!chatSelectedUser) {
      return;
    }

    const trimmedMessage = chatDraft.trim();
    if (!trimmedMessage) {
      return;
    }

    if (!chatSocket || chatSocket.readyState !== WebSocket.OPEN) {
      setChatError('Chat is not connected yet. Please try again.');
      return;
    }

    const messagePayload = buildChatMessagePayload(chatSelectedUser, trimmedMessage);

    chatSocket.send(JSON.stringify(messagePayload));
    setChatMessages((currentMessages) => [
      ...currentMessages,
      {
        id: `local-${Date.now()}`,
        senderId: currentUserId || storedProfile.id || 'me',
        senderName: currentUserName || storedProfile.name || 'You',
        senderProfileImageUrl: currentProfileImageUrl || '',
        content: trimmedMessage,
        createdAt: new Date().toISOString(),
        mine: true,
      },
    ]);
    setChatDraft('');
    setChatError('');
  };

  useEffect(() => {
    const trimmedQuery = groupMemberSearch.trim();
    if (!trimmedQuery) {
      setGroupMemberResults([]);
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        const normalizedResults = await searchUsersForSelection(trimmedQuery, currentUserId || storedProfile.id || '');
        setGroupMemberResults(normalizedResults);
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return;
        }
        console.error('search members for group failed', error);
        setGroupMemberResults([]);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [groupMemberSearch, currentUserId, storedProfile.id]);

  useEffect(() => {
    const trimmedQuery = chatSearch.trim();
    if (!trimmedQuery) {
      setChatSearchResults([]);
      return;
    }

    const controller = new AbortController();
    const timeoutId = window.setTimeout(async () => {
      try {
        setIsSearchingChatUsers(true);
        const normalizedResults = await searchUsersForSelection(trimmedQuery, currentUserId || storedProfile.id || '');
        setChatSearchResults(normalizedResults);
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return;
        }

        console.error('search chat users failed', error);
        setChatSearchResults([]);
      } finally {
        setIsSearchingChatUsers(false);
      }
    }, 250);

    return () => {
      window.clearTimeout(timeoutId);
      controller.abort();
    };
  }, [chatSearch, currentUserId, storedProfile.id]);

  useEffect(() => {
    if (isChatOpen) {
      void loadRecentChats();
    }
  }, [isChatOpen]);

  useEffect(() => {
    if (!chatSelectedUser) {
      return;
    }

    const socket = connectChatSocket({
      currentUserId: currentUserId || storedProfile.id || '',
      onError: (message) => setChatError(message),
      onMessage: (nextMessage) => {
        setChatMessages((currentMessages) => {
          const alreadyExists = currentMessages.some((message) => message.id === nextMessage.id);
          return alreadyExists ? currentMessages : [...currentMessages, nextMessage];
        });
      },
    });

    setChatSocket(socket);

    return () => {
      socket.close();
      setChatSocket(null);
    };
  }, [chatSelectedUser, currentUserId, storedProfile.id]);

  return {
    isChatOpen,
    setIsChatOpen,
    chatSearch,
    setChatSearch,
    chatSearchResults,
    isSearchingChatUsers,
    chatSessions,
    setChatSessions,
    isCreatingGroup,
    setIsCreatingGroup,
    groupNameDraft,
    setGroupNameDraft,
    groupMemberSearch,
    setGroupMemberSearch,
    groupMemberResults,
    setGroupMemberResults,
    selectedGroupMembers,
    setSelectedGroupMembers,
    chatSelectedUser,
    setChatSelectedUser,
    chatMessages,
    setChatMessages,
    chatDraft,
    setChatDraft,
    chatError,
    setChatError,
    isLoadingChatHistory,
    setIsLoadingChatHistory,
    isGroupDetailOpen,
    setIsGroupDetailOpen,
    groupMembers,
    setGroupMembers,
    groupCreatorId,
    setGroupCreatorId,
    openChatSession,
    refreshGroupMembers,
    handleAddMemberToGroup,
    handleRemoveMemberFromGroup,
    handleGroupProfileImageChange,
    toggleGroupMember,
    handleCreateGroup,
    handleSendChatMessage,
    loadRecentChats,
  };
}
