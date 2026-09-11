import { useEffect, useRef, useState } from 'react';
import { useAuthStore } from '../store/authStore';
import { useRoomStore } from '../store/roomStore';
import { useDirectChatStore } from '../store/directChatStore';
import { roomService, directChatService } from '../services/api';
import { WebSocketService } from '../services/websocket';
import { Room, WSMessage, Message } from '../types';
import RoomList from './RoomList';
import ChatWindow from './ChatWindow';
import UserSearchModal from './UserSearchModal';

function ChatApp() {
  const { user, logout } = useAuthStore();
  const { rooms, setRooms, setCurrentRoom, currentRoom, addMessage } = useRoomStore();
  const { setDirectChatsWithInfo } = useDirectChatStore();
  const [loading, setLoading] = useState(true);
  const [showUserSearchModal, setShowUserSearchModal] = useState(false);
  const wsRef = useRef<WebSocketService | null>(null);
  const userNameMapRef = useRef<Map<string, string>>(new Map());

  // Global WebSocket connection that stays alive for all rooms
  useEffect(() => {
    if (!user) return;

    // Create global WebSocket service (no roomId - receives messages for all rooms)
    if (!wsRef.current) {
      wsRef.current = new WebSocketService(user.id);
      // Expose WebSocket service globally for ChatWindow to access
      (window as any).__wsService = wsRef.current;
    }

    wsRef.current.connect((wsMessage: WSMessage) => {
      // Handle direct_chat_created events
      if (wsMessage.type === 'direct_chat_created') {
        console.log('Direct chat created via WebSocket:', wsMessage.roomID);
        window.dispatchEvent(new CustomEvent('directChatCreated'));
        return;
      }

      // Handle message_status events - update status in store
      if (wsMessage.type === 'message_status' && wsMessage.data && wsMessage.roomID) {
        const data = wsMessage.data as Record<string, string>;
        const messageId = data.message_id;
        const status = data.status;
        
        if (messageId && status) {
          const roomStore = useRoomStore.getState();
          roomStore.updateMessageStatus(wsMessage.roomID, messageId, status as any);
        }
        return;
      }
      
      // Handle message_deleted events - remove message from store
      if (wsMessage.type === 'message_deleted' && wsMessage.data && wsMessage.roomID) {
        const data = wsMessage.data as Record<string, string>;
        const messageId = data.message_id;
        
        if (messageId) {
          const roomStore = useRoomStore.getState();
          roomStore.removeMessage(wsMessage.roomID, messageId);
        }
        return;
      }

      // Handle incoming messages
      if (wsMessage.type === 'message' && wsMessage.content && wsMessage.roomID) {
        // Skip our own messages - they are added optimistically in ChatWindow
        if (wsMessage.senderID?.trim() === user.id.trim()) {
          // Remove optimistic message with the same content if it exists
          // (the real message from server will have different ID format)
          return;
        }
        
        // Get sender name from userNameMapRef or use the one from WebSocket
        const senderName = wsMessage.senderName || userNameMapRef.current.get(wsMessage.senderID || '') || 'Unknown User';
        
        const message: Message = {
          id: wsMessage.roomID + '-' + new Date(wsMessage.timestamp || new Date().toISOString()).getTime().toString(),
          roomId: wsMessage.roomID,
          senderId: wsMessage.senderID || '',
          senderName: senderName,
          content: wsMessage.content,
          type: 'm.text',
          createdAt: wsMessage.timestamp || new Date().toISOString(),
        };
        addMessage(wsMessage.roomID, message);
      }
    });

    return () => {
      // Don't remove from window on unmount - only on logout
    };
  }, [user]);

  // Keep userNameMapRef in sync with loadData updates
  useEffect(() => {
    if (!user) return;
    const updateUserNameMap = async () => {
      try {
        const directChatsData = await directChatService.getDirectChats(user.id);
        const newNameMap = new Map(userNameMapRef.current);
        directChatsData.forEach(dc => {
          if (dc.other_participant) {
            newNameMap.set(dc.other_participant.id, dc.other_participant.username);
          }
        });
        userNameMapRef.current = newNameMap;
      } catch (error) {
        console.error('Failed to update user name map:', error);
      }
    };
    updateUserNameMap();
  }, [user]);

  const loadData = async () => {
    if (!user) return;
    try {
      // Загружаем все комнаты
      const roomsData = await roomService.getRooms(user.id);
      setRooms(roomsData);

      // Загружаем личные чаты с информацией об участниках
      const directChatsData = await directChatService.getDirectChats(user.id);
      setDirectChatsWithInfo(directChatsData);

      // Создаем мапу userId -> username для всех участников личных чатов
      const newNameMap = new Map(userNameMapRef.current);
      directChatsData.forEach(dc => {
        if (dc.other_participant) {
          newNameMap.set(dc.other_participant.id, dc.other_participant.username);
        }
      });
      userNameMapRef.current = newNameMap;

      // Обновляем комнаты с информацией об участниках для direct чатов
      const updatedRooms = roomsData.map(room => {
        if (room.type === 'direct') {
          const directChatInfo = directChatsData.find(dc => dc.room.id === room.id);
          if (directChatInfo) {
            return {
              ...room,
              participant: directChatInfo.other_participant,
            };
          }
        }
        return room;
      });
      
      setRooms(updatedRooms);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  // Load data when user is set
  useEffect(() => {
    if (user) {
      loadData();
    }
  }, [user]);

  // Listen for direct chat creation via WebSocket event
  useEffect(() => {
    const handleDirectChatCreated = () => {
      console.log('Received directChatCreated event, reloading data...');
      loadData();
    };

    window.addEventListener('directChatCreated', handleDirectChatCreated);
    return () => {
      window.removeEventListener('directChatCreated', handleDirectChatCreated);
    };
  }, [loadData]);

  const handleSelectRoom = (room: Room) => {
    setCurrentRoom(room);
  };

  const handleCreateRoom = async (name: string, description: string, type: string) => {
    if (!user) return;
    try {
      const newRoom = await roomService.createRoom(name, description, type, [], user.id);
      setRooms([newRoom, ...rooms]);
      setCurrentRoom(newRoom);
    } catch (error) {
      console.error('Failed to create room:', error);
    }
  };

  const handleChatCreated = (roomId: string) => {
    // После создания прямого чата, загружаем обновленные данные
    loadData().then(() => {
      // Находим созданный чат и выбираем его
      if (user) {
        directChatService.getDirectChats(user.id).then(chats => {
          const chat = chats.find(c => c.room.id === roomId);
          if (chat) {
            setCurrentRoom(chat.room);
          }
        }).catch(() => {
          // Fallback: просто перезагружаем данные
          loadData();
        });
      }
    });
  };

  const handleOpenDirectMessageModal = () => {
    setShowUserSearchModal(true);
  };

  if (loading) {
    return <div className='loading'>Loading...</div>;
  }

  return (
    <div className='chat-app'>
      <header className='app-header'>
        <h1>Matrix Messenger</h1>
        <div className='user-info'>
          <span>{user?.username}</span>
          <button onClick={logout}>Logout</button>
        </div>
      </header>
      
      <main className='chat-main'>
        <RoomList
          rooms={rooms}
          onSelectRoom={handleSelectRoom}
          onCreateRoom={handleCreateRoom}
          onOpenDirectMessageModal={handleOpenDirectMessageModal}
        />
        
        {currentRoom ? (
          <ChatWindow key={currentRoom.id} room={currentRoom} />
        ) : (
          <div className='no-room-selected'>
            <p>Select a room to start chatting</p>
          </div>
        )}
      </main>

      <UserSearchModal
        isOpen={showUserSearchModal}
        onClose={() => setShowUserSearchModal(false)}
        onChatCreated={handleChatCreated}
      />
    </div>
  );
}

export default ChatApp;
