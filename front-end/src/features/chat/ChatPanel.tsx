import type { Dispatch, SetStateAction } from 'react';
import type { ChatSession, ChatUserProfile } from '../../shared/types';
import { formatMessageTimestamp } from '../../shared/utils/formatting';

export type ChatPanelProps = {
  isChatOpen: boolean;
  chatSearch: string;
  setChatSearch: Dispatch<SetStateAction<string>>;
  chatSearchResults: ChatUserProfile[];
  isSearchingChatUsers: boolean;
  chatSessions: ChatSession[];
  openChatSession: (user: ChatUserProfile & { conversationId?: string; isGroup?: boolean; groupId?: string }) => Promise<void>;
  isLoadingChatHistory: boolean;
  chatSelectedUser: (ChatUserProfile & { conversationId?: string; isGroup?: boolean; groupId?: string }) | null;
  setChatSelectedUser: Dispatch<SetStateAction<(ChatUserProfile & { conversationId?: string; isGroup?: boolean; groupId?: string }) | null>>;
  chatMessages: Array<{
    id: string;
    senderId: string;
    senderName: string;
    senderProfileImageUrl: string;
    content: string;
    createdAt: string;
    mine: boolean;
  }>;
  chatDraft: string;
  setChatDraft: Dispatch<SetStateAction<string>>;
  chatError: string;
  handleSendChatMessage: () => void;
  isGroupDetailOpen: boolean;
  setIsGroupDetailOpen: Dispatch<SetStateAction<boolean>>;
  groupMembers: ChatUserProfile[];
  groupCreatorId: string;
  setGroupMemberSearch: Dispatch<SetStateAction<string>>;
  groupMemberSearch: string;
  groupMemberResults: ChatUserProfile[];
  refreshGroupMembers: (groupId: string) => Promise<void>;
  handleAddMemberToGroup: (member: ChatUserProfile) => Promise<void>;
  handleRemoveMemberFromGroup: (member: ChatUserProfile) => Promise<void>;
  handleGroupProfileImageChange: (event: React.ChangeEvent<HTMLInputElement>) => Promise<void>;
  setChatError: Dispatch<SetStateAction<string>>;
  isCreatingGroup: boolean;
  setIsCreatingGroup: Dispatch<SetStateAction<boolean>>;
  groupNameDraft: string;
  setGroupNameDraft: Dispatch<SetStateAction<string>>;
  selectedGroupMembers: ChatUserProfile[];
  setSelectedGroupMembers: Dispatch<SetStateAction<ChatUserProfile[]>>;
  toggleGroupMember: (member: ChatUserProfile) => void;
  handleCreateGroup: () => Promise<void>;
  onCloseChat: () => void;
};

const getInitials = (value: string) =>
  value
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join('')
    .toUpperCase() || 'A';

export function ChatPanel({
  isChatOpen,
  chatSearch,
  setChatSearch,
  chatSearchResults,
  isSearchingChatUsers,
  chatSessions,
  openChatSession,
  isLoadingChatHistory,
  chatSelectedUser,
  setChatSelectedUser,
  chatMessages,
  chatDraft,
  setChatDraft,
  chatError,
  handleSendChatMessage,
  isGroupDetailOpen,
  setIsGroupDetailOpen,
  groupMembers,
  groupCreatorId,
  groupMemberSearch,
  setGroupMemberSearch,
  groupMemberResults,
  refreshGroupMembers,
  handleAddMemberToGroup,
  handleRemoveMemberFromGroup,
  handleGroupProfileImageChange,
  setChatError,
  isCreatingGroup,
  setIsCreatingGroup,
  groupNameDraft,
  setGroupNameDraft,
  selectedGroupMembers,
  setSelectedGroupMembers,
  toggleGroupMember,
  handleCreateGroup,
  onCloseChat,
}: ChatPanelProps) {
  if (!isChatOpen) {
    return null;
  }

  return (
    <div className="h-[480px] w-[360px] overflow-hidden rounded-[28px] border border-sky-100 bg-white shadow-[0_24px_60px_rgba(15,23,42,0.18)]">
      <div className="flex items-center justify-between border-b border-slate-100 bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-3 text-white">
        <div className="flex items-center gap-3">
          {chatSelectedUser ? (
            <button
              type="button"
              onClick={() => {
                setChatSelectedUser(null);
                setChatSearch('');
                setChatError('');
              }}
              className="rounded-full border border-white/30 px-2 py-1 text-xs font-medium text-white/90"
            >
              Back
            </button>
          ) : (
            <div className="flex h-9 w-9 items-center justify-center rounded-full border border-white/60 bg-white/15 text-lg">
              💬
            </div>
          )}

          {chatSelectedUser ? (
            chatSelectedUser.isGroup ? (
              <button
                type="button"
                onClick={() => {
                  setIsGroupDetailOpen(true);
                  const groupId = chatSelectedUser.groupId || chatSelectedUser.conversationId || '';
                  if (groupId) {
                    void refreshGroupMembers(groupId);
                  }
                }}
                className="flex items-center gap-3 text-left"
              >
                {chatSelectedUser.profileImageUrl ? (
                  <img src={chatSelectedUser.profileImageUrl} alt={chatSelectedUser.name} className="h-9 w-9 rounded-full border border-white/60 object-cover" />
                ) : (
                  <div className="flex h-9 w-9 items-center justify-center rounded-full border border-white/60 bg-white/15 text-xs font-bold">
                    {getInitials(chatSelectedUser.name)}
                  </div>
                )}
                <div>
                  <p className="text-sm font-bold">{chatSelectedUser.name}</p>
                  <p className="text-[10px] text-sky-100">{chatSelectedUser.username}</p>
                </div>
              </button>
            ) : (
              <>
                {chatSelectedUser.profileImageUrl ? (
                  <img src={chatSelectedUser.profileImageUrl} alt={chatSelectedUser.name} className="h-9 w-9 rounded-full border border-white/60 object-cover" />
                ) : (
                  <div className="flex h-9 w-9 items-center justify-center rounded-full border border-white/60 bg-white/15 text-xs font-bold">
                    {getInitials(chatSelectedUser.name)}
                  </div>
                )}
                <div>
                  <p className="text-sm font-bold">{chatSelectedUser.name}</p>
                  <p className="text-[10px] text-sky-100">{chatSelectedUser.username}</p>
                </div>
              </>
            )
          ) : (
            <div>
              <p className="text-sm font-bold">Messages</p>
              <p className="text-[10px] text-sky-100">Recent chats</p>
            </div>
          )}
        </div>

        <button type="button" onClick={onCloseChat} className="rounded-full border border-white/30 px-2 py-1 text-xs font-medium text-white/90">
          Close
        </button>
      </div>

      {!chatSelectedUser ? (
        <div className="flex h-[calc(480px-64px)] flex-col bg-slate-50">
          <div className="relative border-b border-slate-100 bg-slate-50 p-3">
            <div className="mb-3 flex items-center justify-between gap-2">
              <div className="text-sm font-semibold text-slate-700">Recent chats</div>
              <button
                type="button"
                onClick={() => {
                  setIsCreatingGroup(true);
                  setGroupNameDraft('');
                  setChatError('');
                  setGroupMemberSearch('');
                }}
                className="rounded-full bg-sky-100 px-2.5 py-1.5 text-[10px] font-semibold uppercase tracking-[0.12em] text-sky-700 transition hover:bg-sky-200"
              >
                New group
              </button>
            </div>
            <div className="relative">
              <input
                value={chatSearch}
                onChange={(event) => setChatSearch(event.target.value)}
                placeholder="Search messages or people"
                className="w-full rounded-full border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
              />
              {isSearchingChatUsers && (
                <span className="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-medium uppercase tracking-[0.18em] text-sky-600">
                  Searching
                </span>
              )}
            </div>

            {chatSearch && chatSearchResults.length > 0 && (
              <div className="absolute left-3 right-3 top-[calc(100%+6px)] z-20 max-h-52 overflow-y-auto rounded-2xl border border-slate-200 bg-white p-2 shadow-lg">
                {chatSearchResults.map((user) => (
                  <button
                    key={user.id}
                    type="button"
                    onClick={() => {
                      setChatSearch('');
                      void openChatSession({
                        id: user.id,
                        name: user.name,
                        username: user.username,
                        profileImageUrl: user.profileImageUrl,
                        conversationId: '',
                      });
                    }}
                    className="flex w-full items-center gap-3 rounded-xl px-2 py-2 text-left transition hover:bg-slate-50"
                  >
                    {user.profileImageUrl ? (
                      <img src={user.profileImageUrl} alt={user.name} className="h-9 w-9 rounded-full object-cover" />
                    ) : (
                      <div className="flex h-9 w-9 items-center justify-center rounded-full bg-sky-100 text-xs font-bold text-sky-700">
                        {getInitials(user.name)}
                      </div>
                    )}
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-semibold text-slate-800">{user.name}</p>
                      <p className="truncate text-xs text-slate-500">{user.username}</p>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="flex-1 overflow-y-auto px-2 py-2">
            {chatSessions.length === 0 ? (
              <div className="flex h-full items-center justify-center px-4 text-center text-sm text-slate-500">
                No recent chats yet. Search for someone to start one.
              </div>
            ) : (
              <div className="space-y-2">
                {chatSessions
                  .filter((session) => {
                    const query = chatSearch.trim().toLowerCase();
                    if (!query) return true;
                    return (
                      session.name.toLowerCase().includes(query) ||
                      session.username.toLowerCase().includes(query.replace(/^@/, '')) ||
                      session.lastMessage.toLowerCase().includes(query)
                    );
                  })
                  .map((session) => (
                    <button
                      key={session.conversationId || session.id}
                      type="button"
                      onClick={() => void openChatSession(session)}
                      className="flex w-full items-center gap-3 rounded-2xl bg-white p-3 text-left shadow-sm ring-1 ring-slate-100 transition hover:bg-slate-50"
                    >
                      {session.profileImageUrl ? (
                        <img src={session.profileImageUrl} alt={session.name} className="h-12 w-12 rounded-full object-cover" />
                      ) : (
                        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                          {getInitials(session.name)}
                        </div>
                      )}
                      <div className="min-w-0 flex-1">
                        <div className="mb-1 flex items-center justify-between gap-2">
                          <p className="truncate text-sm font-semibold text-slate-800">{session.name}</p>
                          <span className="text-[10px] text-slate-400">{session.lastMessageTime}</span>
                        </div>
                        <p className="truncate text-xs text-slate-500">{session.username}</p>
                        <p className="mt-1 truncate text-xs text-slate-600">{session.lastMessage || 'Start a conversation'}</p>
                      </div>
                    </button>
                  ))}
              </div>
            )}
          </div>
        </div>
      ) : (
        <div className="flex h-[calc(480px-64px)] flex-col bg-slate-50">
          <div className="flex-1 space-y-2 overflow-y-auto px-3 py-3">
            {isLoadingChatHistory ? (
              <div className="flex h-full items-center justify-center text-sm text-slate-500">Loading conversation...</div>
            ) : chatMessages.length > 0 ? (
              chatMessages.map((message) => (
                <div key={message.id} className={`flex ${message.mine ? 'justify-end' : 'justify-start'}`}>
                  <div className={`flex max-w-[80%] flex-col ${message.mine ? 'items-end' : 'items-start'}`}>
                    {!message.mine && (
                      <div className="mb-1 flex items-center gap-2 px-1">
                        {message.senderProfileImageUrl ? (
                          <img src={message.senderProfileImageUrl} alt={message.senderName} className="h-6 w-6 rounded-full object-cover" />
                        ) : (
                          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[9px] font-bold text-white">
                            {getInitials(message.senderName || 'User')}
                          </div>
                        )}
                        <span className="text-[11px] font-medium text-slate-600">{message.senderName || 'User'}</span>
                      </div>
                    )}
                    {message.mine && <div className="mb-1 px-1 text-[11px] font-medium text-slate-500">You</div>}
                    <div
                      className={`rounded-2xl px-3 py-2 text-sm ${
                        message.mine ? 'bg-gradient-to-r from-sky-500 to-blue-600 text-white' : 'bg-white text-slate-700 ring-1 ring-slate-200'
                      }`}
                    >
                      {message.content}
                    </div>
                    <span className="mt-1 text-[10px] text-slate-400">{formatMessageTimestamp(message.createdAt)}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="flex h-full items-center justify-center text-center text-sm text-slate-500">
                Start the conversation with {chatSelectedUser.name}
              </div>
            )}
          </div>

          <div className="border-t border-slate-200 bg-white p-3">
            {chatError && (
              <div className="mb-2 rounded-xl border border-red-200 bg-red-50 px-2 py-1 text-xs text-red-600">
                {chatError}
              </div>
            )}
            <div className="flex items-center gap-2">
              <input
                value={chatDraft}
                onChange={(event) => setChatDraft(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === 'Enter' && !event.shiftKey) {
                    event.preventDefault();
                    handleSendChatMessage();
                  }
                }}
                placeholder="Type a message"
                className="flex-1 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
              />
              <button type="button" onClick={handleSendChatMessage} className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white">
                Send
              </button>
            </div>
          </div>
        </div>
      )}

      {isGroupDetailOpen && chatSelectedUser?.isGroup && (
        <div className="absolute inset-0 z-40 flex items-center justify-center bg-slate-900/30 p-4">
          <div className="w-full max-w-md rounded-3xl bg-white p-4 shadow-2xl">
            <div className="mb-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                {chatSelectedUser.profileImageUrl ? (
                  <img src={chatSelectedUser.profileImageUrl} alt={chatSelectedUser.name} className="h-11 w-11 rounded-full object-cover" />
                ) : (
                  <div className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-sm font-bold text-white">
                    {getInitials(chatSelectedUser.name)}
                  </div>
                )}
                <div>
                  <p className="text-sm font-semibold text-slate-800">{chatSelectedUser.name}</p>
                  <p className="text-xs text-slate-500">{chatSelectedUser.username}</p>
                </div>
              </div>
              <button type="button" onClick={() => { setIsGroupDetailOpen(false); setChatError(''); }} className="text-sm text-slate-500">
                Close
              </button>
            </div>

            <div className="mb-4 rounded-2xl border border-dashed border-sky-200 bg-sky-50 p-3">
              <label className="block text-xs font-semibold uppercase tracking-[0.15em] text-sky-600">Update group photo</label>
              <input type="file" accept="image/*" onChange={handleGroupProfileImageChange} className="mt-2 block w-full text-xs text-slate-600 file:mr-3 file:rounded-full file:border-0 file:bg-sky-100 file:px-3 file:py-2 file:text-xs file:font-semibold file:text-sky-700" />
            </div>

            <div className="mb-3">
              <label className="mb-1 block text-xs font-medium text-slate-600">Add members</label>
              <input
                value={groupMemberSearch}
                onChange={(event) => setGroupMemberSearch(event.target.value)}
                placeholder="Search users to add"
                className="w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
              />
            </div>

            {groupMemberResults.length > 0 && (
              <div className="mb-3 max-h-40 space-y-2 overflow-y-auto">
                {groupMemberResults.map((user) => (
                  <button key={user.id} type="button" onClick={() => void handleAddMemberToGroup(user)} className="flex w-full items-center justify-between rounded-xl border border-slate-200 bg-white px-2 py-2 text-left">
                    <div className="flex items-center gap-2">
                      {user.profileImageUrl ? (
                        <img src={user.profileImageUrl} alt={user.name} className="h-8 w-8 rounded-full object-cover" />
                      ) : (
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[10px] font-bold text-white">
                          {getInitials(user.name)}
                        </div>
                      )}
                      <div>
                        <p className="text-sm font-medium text-slate-700">{user.name}</p>
                        <p className="text-xs text-slate-500">{user.username}</p>
                      </div>
                    </div>
                    <span className="text-xs font-medium text-sky-600">Add</span>
                  </button>
                ))}
              </div>
            )}

            <div className="mb-2 flex items-center justify-between">
              <p className="text-sm font-semibold text-slate-700">Members</p>
              <span className="text-xs text-slate-500">{groupMembers.length}</span>
            </div>

            <div className="max-h-64 space-y-2 overflow-y-auto">
              {groupMembers.length > 0 ? (
                groupMembers.map((member) => {
                  const isCreatorMember = !!groupCreatorId && String(member.id) === String(groupCreatorId);
                  return (
                    <div key={member.id} className="flex items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-2 py-2">
                      <div className="flex items-center gap-2">
                        {member.profileImageUrl ? (
                          <img src={member.profileImageUrl} alt={member.name} className="h-8 w-8 rounded-full object-cover" />
                        ) : (
                          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[10px] font-bold text-white">
                            {getInitials(member.name)}
                          </div>
                        )}
                        <div>
                          <p className="text-sm font-medium text-slate-700">{member.name}</p>
                          <p className="text-xs text-slate-500">{member.username}</p>
                        </div>
                      </div>
                      <button
                        type="button"
                        disabled={isCreatorMember}
                        onClick={() => {
                          if (!isCreatorMember) {
                            void handleRemoveMemberFromGroup(member);
                          }
                        }}
                        className={`rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.12em] ${
                          isCreatorMember
                            ? 'cursor-not-allowed border-slate-200 bg-slate-100 text-slate-400'
                            : 'border-red-200 bg-red-50 text-red-600'
                        }`}
                      >
                        {isCreatorMember ? 'Creator' : 'Remove'}
                      </button>
                    </div>
                  );
                })
              ) : (
                <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50 px-3 py-4 text-center text-xs text-slate-500">
                  No members yet.
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {isCreatingGroup && (
        <div className="absolute inset-0 z-40 flex items-center justify-center bg-slate-900/30 p-4">
          <div className="w-full max-w-md rounded-3xl bg-white p-4 shadow-2xl">
            <div className="mb-3 flex items-center justify-between">
              <h3 className="text-base font-semibold text-slate-800">Create group</h3>
              <button
                type="button"
                onClick={() => {
                  setIsCreatingGroup(false);
                  setGroupNameDraft('');
                  setGroupMemberSearch('');
                  setSelectedGroupMembers([]);
                  setChatError('');
                }}
                className="text-sm text-slate-500"
              >
                Close
              </button>
            </div>

            <div className="space-y-3">
              <input
                value={groupNameDraft}
                onChange={(event) => setGroupNameDraft(event.target.value)}
                placeholder="Group name"
                className="w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
              />

              <div>
                <label className="mb-1 block text-xs font-medium text-slate-600">Add members</label>
                <input
                  value={groupMemberSearch}
                  onChange={(event) => setGroupMemberSearch(event.target.value)}
                  placeholder="Search users by username"
                  className="w-full rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700 placeholder:text-slate-400 focus:border-sky-400 focus:outline-none"
                />
              </div>

              {selectedGroupMembers.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {selectedGroupMembers.map((member) => (
                    <button key={member.id} type="button" onClick={() => toggleGroupMember(member)} className="rounded-full border border-sky-200 bg-sky-50 px-2.5 py-1 text-xs font-medium text-sky-700">
                      {member.username}
                    </button>
                  ))}
                </div>
              )}

              <div className="max-h-52 space-y-2 overflow-y-auto">
                {groupMemberResults.length > 0 ? (
                  groupMemberResults.map((user) => {
                    const isSelected = selectedGroupMembers.some((member) => member.id === user.id);
                    return (
                      <button
                        key={user.id}
                        type="button"
                        onClick={() => toggleGroupMember(user)}
                        className={`flex w-full items-center justify-between rounded-xl border px-2 py-2 text-left ${
                          isSelected ? 'border-sky-200 bg-sky-50' : 'border-slate-200 bg-white'
                        }`}
                      >
                        <div className="flex items-center gap-2">
                          {user.profileImageUrl ? (
                            <img src={user.profileImageUrl} alt={user.name} className="h-8 w-8 rounded-full object-cover" />
                          ) : (
                            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-sky-400 to-blue-600 text-[10px] font-bold text-white">
                              {getInitials(user.name)}
                            </div>
                          )}
                          <div>
                            <p className="text-sm font-medium text-slate-700">{user.name}</p>
                            <p className="text-xs text-slate-500">{user.username}</p>
                          </div>
                        </div>
                        <span className="text-xs font-medium text-sky-600">{isSelected ? 'Added' : 'Add'}</span>
                      </button>
                    );
                  })
                ) : (
                  <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50 px-3 py-4 text-center text-xs text-slate-500">
                    Search for users to add to the group.
                  </div>
                )}
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => {
                    setIsCreatingGroup(false);
                    setGroupNameDraft('');
                    setGroupMemberSearch('');
                    setSelectedGroupMembers([]);
                  }}
                  className="rounded-full border border-slate-200 px-3 py-2 text-sm font-medium text-slate-600"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={() => void handleCreateGroup()}
                  disabled={!groupNameDraft.trim() || selectedGroupMembers.length === 0}
                  className="rounded-full bg-gradient-to-r from-sky-500 to-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
                >
                  Create group
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}
