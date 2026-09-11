import { create } from 'zustand';
import { UserSearchResult, Room, DirectChatResponse } from '../types';
import { directChatService } from '../services/api';

interface DirectChatState {
  searchQuery: string;
  searchResults: UserSearchResult[];
  isSearching: boolean;
  selectedChat: Room | null;
  directChatsWithInfo: DirectChatResponse[];
  
  // Actions for search
  setSearchQuery: (query: string) => void;
  setSearchResults: (results: UserSearchResult[]) => void;
  setIsSearching: (isSearching: boolean) => void;
  startSearch: (query: string, currentUserId: string) => Promise<void>;
  
  // Actions for chat creation
  createDirectChat: (targetUserId: string, currentUserId: string) => Promise<Room>;
  
  // Actions for selected chat
  setSelectedChat: (chat: Room | null) => void;
  
  // Actions for loading direct chats
  setDirectChatsWithInfo: (chats: DirectChatResponse[]) => void;
  addDirectChat: (chat: DirectChatResponse) => void;
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null;

export const useDirectChatStore = create<DirectChatState>((set) => ({
  searchQuery: '',
  searchResults: [],
  isSearching: false,
  selectedChat: null,
  directChatsWithInfo: [],

  setSearchQuery: (query) => {
    set({ searchQuery: query });
    
    // Debounce search with 300ms delay
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }
    
    debounceTimer = setTimeout(() => {
      // Note: actual search is triggered via startSearch() which needs userId
    }, 300);
  },

  setSearchResults: (results) => set({ searchResults: results }),
  
  setIsSearching: (isSearching) => set({ isSearching }),

  startSearch: async (query: string, currentUserId: string) => {
    if (!query.trim()) {
      set({ searchResults: [], isSearching: false });
      return;
    }

    set({ isSearching: true });
    try {
      const results = await directChatService.searchUsers(query, currentUserId);
      set({ searchResults: results, isSearching: false });
    } catch (error) {
      console.error('Failed to search users:', error);
      set({ isSearching: false });
    }
  },

  createDirectChat: async (targetUserId: string, currentUserId: string) => {
    const response = await directChatService.getOrCreateDirectChat(targetUserId, currentUserId);
    set((state) => ({
      directChatsWithInfo: [response, ...state.directChatsWithInfo],
      selectedChat: response.room,
    }));
    return response.room;
  },

  setSelectedChat: (chat) => set({ selectedChat: chat }),

  setDirectChatsWithInfo: (chats) => set({ directChatsWithInfo: chats }),

  addDirectChat: (chat) => set((state) => ({
    directChatsWithInfo: [chat, ...state.directChatsWithInfo],
  })),
}));
