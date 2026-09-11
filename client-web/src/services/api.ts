import { User, Room, Message, MessageStatus, UserSearchResult, DirectChatResponse } from '../types';

const API_BASE = '/api/v1';

export const authService = {
  register: async (username: string, email: string, password: string): Promise<User> => {
    const response = await fetch(`${API_BASE}/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, email, password }),
    });
    if (!response.ok) throw new Error('Registration failed');
    return response.json();
  },

  login: async (username: string, password: string): Promise<User> => {
    const response = await fetch(`${API_BASE}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) throw new Error('Login failed');
    return response.json();
  },
};

export const roomService = {
  createRoom: async (
    name: string,
    description: string,
    type: string,
    members: string[],
    userId: string
  ): Promise<Room> => {
    const response = await fetch(`${API_BASE}/rooms`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': userId,
      },
      body: JSON.stringify({ name, description, type, members }),
    });
    if (!response.ok) throw new Error('Failed to create room');
    return response.json();
  },

  getRooms: async (userId: string): Promise<Room[]> => {
    const response = await fetch(`${API_BASE}/rooms`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to get rooms');
    return response.json();
  },

  getDirectChats: async (userId: string): Promise<DirectChatResponse[]> => {
    const response = await fetch(`${API_BASE}/direct-chats`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to get direct chats');
    return response.json();
  },

  getMessages: async (roomId: string, userId: string): Promise<Message[]> => {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/messages`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to get messages');
    const msgs = await response.json();
    // Map sender_name to senderName for frontend
    return msgs.map((msg: Message & { sender_name?: string }) => ({
      ...msg,
      senderName: msg.sender_name || msg.senderName,
    }));
  },

  sendMessage: async (
    roomId: string,
    content: string,
    userId: string
  ): Promise<Message> => {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': userId,
      },
      body: JSON.stringify({ content, type: 'm.text' }),
    });
    if (!response.ok) throw new Error('Failed to send message');
    return response.json();
  },

  updateMessageStatus: async (
    messageId: string,
    status: MessageStatus,
    roomId: string,
    userId: string
  ): Promise<void> => {
    const response = await fetch(`${API_BASE}/messages/status`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': userId,
      },
      body: JSON.stringify({ message_id: messageId, status, room_id: roomId }),
    });
    if (!response.ok) throw new Error('Failed to update message status');
  },

  markRoomMessagesAsRead: async (roomId: string, userId: string): Promise<void> => {
    const response = await fetch(`${API_BASE}/rooms/${roomId}/messages/read`, {
      method: 'POST',
      headers: {
        'X-User-ID': userId,
      },
    });
    if (!response.ok) throw new Error('Failed to mark messages as read');
  },

  deleteMessage: async (messageId: string, userId: string): Promise<void> => {
    const response = await fetch(`${API_BASE}/messages/${messageId}/delete`, {
      method: 'DELETE',
      headers: {
        'X-User-ID': userId,
      },
    });
    if (!response.ok) throw new Error('Failed to delete message');
  },
};

export const directChatService = {
  // Поиск пользователей для начала личного чата
  searchUsers: async (query: string, currentUserId: string): Promise<UserSearchResult[]> => {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    params.set('current_user', currentUserId);
    
    const response = await fetch(`${API_BASE}/users?${params.toString()}`, {
      headers: { 'X-User-ID': currentUserId },
    });
    if (!response.ok) throw new Error('Failed to search users');
    return response.json();
  },

  // Получить или создать личный чат
  getOrCreateDirectChat: async (targetUserId: string, currentUserId: string): Promise<DirectChatResponse> => {
    const response = await fetch(`${API_BASE}/direct-chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': currentUserId,
      },
      body: JSON.stringify({ user_id: targetUserId }),
    });
    if (!response.ok) throw new Error('Failed to get/create direct chat');
    return response.json();
  },

  // Получить все личные чаты
  getDirectChats: async (userId: string): Promise<DirectChatResponse[]> => {
    const response = await fetch(`${API_BASE}/direct-chats`, {
      headers: { 'X-User-ID': userId },
    });
    if (!response.ok) throw new Error('Failed to get direct chats');
    return response.json();
  },
};
