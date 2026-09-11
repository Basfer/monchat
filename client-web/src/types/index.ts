export interface User {
  id: string;
  username: string;
  email: string;
}

export type RoomType = 'direct' | 'group' | 'channel' | 'public';
export type RoomRole = 'creator' | 'admin' | 'member' | 'subscriber';

export interface Room {
  id: string;
  name?: string | null;
  description?: string | null;
  type: RoomType;
  createdBy: string;
  isPrivate: boolean;
  createdAt: string;
  updatedAt: string;
  memberCount?: number;
  lastMessage?: Message;
  
  // Для личных чатов - информация о втором участнике
  participant?: UserSearchResult;
}

export type MessageStatus = 'new' | 'delivered' | 'read' | 'deleted';

export interface Message {
  id: string;
  roomId: string;
  senderId: string;
  senderName?: string;
  content: string;
  type: string;
  status?: MessageStatus;
  createdAt: string;
}

export interface WSMessage {
  type: string;
  roomID: string;
  userID?: string;
  senderID?: string;
  senderName?: string;
  content?: string;
  data?: Record<string, unknown>;
  timestamp: string;
}

// Результаты поиска пользователей
export interface UserSearchResult {
  id: string;
  username: string;
  email: string;
}

// Ответ API при создании/получении личного чата
export interface DirectChatResponse {
  room: Room;
  other_participant: UserSearchResult;
}
