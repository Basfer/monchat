import { create } from 'zustand';
import { Room, Message, MessageStatus } from '../types';

interface RoomState {
  rooms: Room[];
  currentRoom: Room | null;
  currentRoomId: string | null;
  messages: Record<string, Message[]>;
  setRooms: (rooms: Room[]) => void;
  setCurrentRoom: (room: Room | null) => void;
  addMessage: (roomId: string, message: Message) => void;
  addMessages: (roomId: string, messages: Message[]) => void;
  clearMessages: (roomId: string) => void;
  clearRoomMessages: (roomId: string) => void;
  updateMessageStatus: (roomId: string, messageId: string, status: MessageStatus) => void;
  removeMessage: (roomId: string, messageId: string) => void;
  // Direct chat support
  addRoom: (room: Room) => void;
  getDirectChats: () => Room[];
  getNonDirectRooms: () => Room[];
}

export const useRoomStore = create<RoomState>((set, get) => ({
  rooms: [],
  currentRoom: null,
  currentRoomId: null,
  messages: {},
  setRooms: (rooms) => set({ rooms }),
  setCurrentRoom: (room) => {
    set({ currentRoom: room });
    if (room) {
      set((state) => ({
        currentRoomId: room.id,
        messages: {
          ...state.messages,
          [room.id]: state.messages[room.id] || [],
        },
      }));
    } else {
      set({ currentRoomId: null });
    }
  },
  addMessage: (roomId, message) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: [...(state.messages[roomId] || []), message],
      },
    })),
  addMessages: (roomId, messages) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: [...(state.messages[roomId] || []), ...messages],
      },
    })),
  clearMessages: (roomId) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: [],
      },
    })),
  clearRoomMessages: (roomId) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: [],
      },
    })),
  updateMessageStatus: (roomId, messageId, status) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: (state.messages[roomId] || []).map(msg =>
          msg.id === messageId ? { ...msg, status } : msg
        ),
      },
    })),
  removeMessage: (roomId, messageId) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [roomId]: (state.messages[roomId] || []).filter(msg => msg.id !== messageId),
      },
    })),
  
  // Добавление одной комнаты в список
  addRoom: (room) => set((state) => ({
    rooms: [room, ...state.rooms.filter((r) => r.id !== room.id)],
  })),
  
  // Получение только прямых чатов
  getDirectChats: () => {
    return get().rooms.filter((room) => room.type === 'direct');
  },
  
  // Получение всех комнат кроме прямых чатов
  getNonDirectRooms: () => {
    return get().rooms.filter((room) => room.type !== 'direct');
  },
}));
