export type FeedComment = {
  id: string;
  author: string;
  text: string;
  time: string;
  profileImageUrl?: string;
  replyCount?: number;
  replies?: FeedComment[];
};

export type FeedPost = {
  id: string;
  userId?: string;
  author: string;
  blueTick?: boolean;
  handle: string;
  time: string;
  content: string;
  imageUrls: string[];
  profileImageUrl?: string;
  isLiked: boolean;
  likes: number;
  comments: number;
  shares: number;
  accent: string;
  tag: string;
};

export type ChatUserProfile = {
  id: string;
  name: string;
  username: string;
  profileImageUrl: string;
};

export type ChatSession = ChatUserProfile & {
  conversationId: string;
  recipientId?: string;
  lastMessage: string;
  lastMessageTime: string;
  isGroup?: boolean;
  groupId?: string;
};

export type CurrentUserProfile = {
  id: string;
  name: string;
  username: string;
  profileImageUrl: string;
  blueTick?: boolean;
  followerCount: number;
  followingCount: number;
  postCount: number;
};
