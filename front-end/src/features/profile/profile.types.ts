export type FeedComment = {
  id: string;
  author: string;
  text: string;
  time: string;
  profileImageUrl?: string;
  replyCount?: number;
  replies?: FeedComment[];
};

export type ProfilePost = {
  id: string;
  content: string;
  imageUrls: string[];
  likes: number;
  comments: number;
  time: string;
  isLiked?: boolean;
};

export type CustomLink = {
  id: string;
  label: string;
  url: string;
};

export type ProfileUser = {
  id: string;
  name: string;
  username: string;
  profileImageUrl?: string;
  bio?: string;
  website?: string;
  linkedin?: string;
  instagram?: string;
  customLinks?: CustomLink[];
  followerCount?: number;
  followingCount?: number;
  blueTick?: boolean;
};

export type ProfilePageData = {
  posts: ProfilePost[];
  user: ProfileUser;
  followerCount: number;
  followingCount: number;
  isFollowing: boolean;
};
