import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { readStoredProfile } from '../../shared/utils/normalizers';
import { ChatPanel } from './ChatPanel';
import { useChatManagement } from './useChatManagement';

export function FloatingChatWidget() {
  const location = useLocation();
  const storedProfile = readStoredProfile();
  const currentUserId = storedProfile.id || '';
  const currentUserName = (storedProfile.name || 'User').trim() || 'User';
  const currentProfileImageUrl = storedProfile.profileImageUrl || '';

  const chat = useChatManagement({
    currentUserId,
    currentUserName,
    currentProfileImageUrl,
    storedProfile,
  });

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const targetUserId = params.get('chatUserId') || params.get('chat_user_id');
    const targetUserName = params.get('chatUserName') || params.get('chat_user_name') || '';
    const targetUserUsername = params.get('chatUserUsername') || params.get('chat_user_username') || '';
    const targetUserImage = params.get('chatUserImage') || params.get('chat_user_image') || '';

    if (!targetUserId) {
      return;
    }

    const startupUser = {
      id: targetUserId,
      name: targetUserName || targetUserUsername || 'User',
      username: targetUserUsername || '@user',
      profileImageUrl: targetUserImage,
    };

    chat.setIsChatOpen(true);
    void chat.openChatSession(startupUser);
  }, [location.key, location.search]);

  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3">
      <ChatPanel
        isChatOpen={chat.isChatOpen}
        chatSearch={chat.chatSearch}
        setChatSearch={chat.setChatSearch}
        chatSearchResults={chat.chatSearchResults}
        isSearchingChatUsers={chat.isSearchingChatUsers}
        chatSessions={chat.chatSessions}
        openChatSession={chat.openChatSession}
        isLoadingChatHistory={chat.isLoadingChatHistory}
        chatSelectedUser={chat.chatSelectedUser}
        setChatSelectedUser={chat.setChatSelectedUser}
        chatMessages={chat.chatMessages}
        chatDraft={chat.chatDraft}
        setChatDraft={chat.setChatDraft}
        chatError={chat.chatError}
        handleSendChatMessage={chat.handleSendChatMessage}
        isGroupDetailOpen={chat.isGroupDetailOpen}
        setIsGroupDetailOpen={chat.setIsGroupDetailOpen}
        groupMembers={chat.groupMembers}
        groupCreatorId={chat.groupCreatorId}
        groupMemberSearch={chat.groupMemberSearch}
        setGroupMemberSearch={chat.setGroupMemberSearch}
        groupMemberResults={chat.groupMemberResults}
        refreshGroupMembers={chat.refreshGroupMembers}
        handleAddMemberToGroup={chat.handleAddMemberToGroup}
        handleRemoveMemberFromGroup={chat.handleRemoveMemberFromGroup}
        handleGroupProfileImageChange={chat.handleGroupProfileImageChange}
        setChatError={chat.setChatError}
        isCreatingGroup={chat.isCreatingGroup}
        setIsCreatingGroup={chat.setIsCreatingGroup}
        groupNameDraft={chat.groupNameDraft}
        setGroupNameDraft={chat.setGroupNameDraft}
        selectedGroupMembers={chat.selectedGroupMembers}
        setSelectedGroupMembers={chat.setSelectedGroupMembers}
        toggleGroupMember={chat.toggleGroupMember}
        handleCreateGroup={chat.handleCreateGroup}
        onCloseChat={() => chat.setIsChatOpen(false)}
      />

      <button
        type="button"
        onClick={() => chat.setIsChatOpen((current) => !current)}
        className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-r from-sky-500 to-blue-600 text-2xl text-white shadow-[0_18px_40px_rgba(59,130,246,0.45)] transition hover:scale-[1.02]"
        aria-label="Open messages"
      >
        💬
      </button>
    </div>
  );
}
