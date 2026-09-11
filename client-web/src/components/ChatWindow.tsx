import { useEffect, useState, useRef } from 'react';
import { useAuthStore } from '../store/authStore';
import { useRoomStore } from '../store/roomStore';
import { roomService, directChatService } from '../services/api';
import { WebSocketService } from '../services/websocket';
import { Room, Message, MessageStatus, WSMessage } from '../types';

interface ChatWindowProps {
  room: Room;
}

function ChatWindow({ room }: ChatWindowProps) {
  const { user } = useAuthStore();
  const { messages, addMessage, addMessages, clearRoomMessages, updateMessageStatus, removeMessage } = useRoomStore();
  const [messageText, setMessageText] = useState('');
  const [loading, setLoading] = useState(true);
  const [participantName, setParticipantName] = useState<string>('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsServiceRef = useRef<WebSocketService | null>(null);

  // Get WebSocket service instance from window (set by ChatApp)
  useEffect(() => {
    if (user) {
      wsServiceRef.current = (window as any).__wsService;
    }
  }, [user]);

  // Mark messages as read when viewing a room
  useEffect(() => {
    if (user && room.id) {
      markMessagesAsRead();
    }
  }, [room.id, user]);

  const markMessagesAsRead = async () => {
    if (!user || !room.id) return;
    try {
      // Optimistically mark all messages in current room as read
      const roomStore = useRoomStore.getState();
      const roomMessages = roomStore.messages[room.id] || [];
      roomMessages.forEach(msg => {
        if (msg.senderId !== user.id && msg.status !== 'read') {
          roomStore.updateMessageStatus(room.id, msg.id, 'read');
        }
      });
      
      // Also notify server
      await roomService.markRoomMessagesAsRead(room.id, user.id);
    } catch (error) {
      console.error('Failed to mark messages as read:', error);
    }
  };

  // Handle WebSocket events for status updates and message deletion
  useEffect(() => {
    if (!user || !wsServiceRef.current) return;

    const handleWsMessage = (wsMessage: WSMessage) => {
      // Handle message_status events
      if (wsMessage.type === 'message_status' && wsMessage.data && wsMessage.roomID) {
        const data = wsMessage.data as Record<string, string>;
        const messageId = data.message_id;
        const status = data.status as MessageStatus;
        
        if (messageId && status) {
          updateMessageStatus(wsMessage.roomID, messageId, status);
        }
      }
      
      // Handle message_deleted events
      if (wsMessage.type === 'message_deleted' && wsMessage.data && wsMessage.roomID) {
        const data = wsMessage.data as Record<string, string>;
        const messageId = data.message_id;
        
        if (messageId) {
          removeMessage(wsMessage.roomID, messageId);
        }
      }
    };

    wsServiceRef.current.connect(handleWsMessage);

    return () => {
      // Note: We don't disconnect here since the WebSocket is shared across components
    };
  }, [user, room.id, updateMessageStatus, removeMessage]);

  useEffect(() => {
    loadMessages();
    
    // Получаем имя участника для прямых чатов
    if (room.type === 'direct' && room.participant) {
      setParticipantName(room.participant.username);
    } else if (room.type === 'direct') {
      // Если нет информации о участнике в room, загружаем её
      loadDirectChatInfo();
    }
  }, [room.id, room.type, room.participant, user]);

  useEffect(() => {
    scrollToBottom();
  }, [messages[room.id]]);

  const loadDirectChatInfo = async () => {
    if (!user || room.type !== 'direct') return;
    try {
      const directChats = await directChatService.getDirectChats(user.id);
      const chat = directChats.find(dc => dc.room.id === room.id);
      if (chat && chat.other_participant) {
        setParticipantName(chat.other_participant.username);
      }
    } catch (error) {
      console.error('Failed to load direct chat info:', error);
    }
  };

  const loadMessages = async () => {
    if (!user) return;
    try {
      clearRoomMessages(room.id);
      const msgs = await roomService.getMessages(room.id, user.id);
      addMessages(room.id, msgs.reverse());
    } catch (error) {
      console.error('Failed to load messages:', error);
    } finally {
      setLoading(false);
    }
  };


  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!messageText.trim() || !user) return;

    // Create optimistic message for immediate UI update
    const optimisticMessage: Message = {
      id: `optimistic-${Date.now()}`,
      roomId: room.id,
      senderId: user.id,
      senderName: user.username,
      content: messageText,
      type: 'm.text',
      createdAt: new Date().toISOString(),
      status: 'new',
    };
    
    // Add optimistic message immediately
    addMessage(room.id, optimisticMessage);
    setMessageText('');

    // Send via WebSocket if available
    if (wsServiceRef.current) {
      wsServiceRef.current.sendMessage(messageText, room.id);
    } else {
      // Fallback to HTTP API if WebSocket is not available
      try {
        await roomService.sendMessage(room.id, messageText, user.id);
      } catch (error) {
        console.error('Failed to send message:', error);
      }
    }
  };

  const handleDeleteMessage = async (messageId: string) => {
    if (!user) return;
    try {
      await roomService.deleteMessage(messageId, user.id);
      removeMessage(room.id, messageId);
    } catch (error) {
      console.error('Failed to delete message:', error);
    }
  };

  const formatTime = (createdAt: unknown) => {
    if (!createdAt || typeof createdAt !== 'string') return '';
    const date = new Date(createdAt);
    if (isNaN(date.getTime())) return '';
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const getHeaderTitle = () => {
    if (room.type === 'direct') {
      return participantName || 'Loading...';
    }
    return room.name || 'Untitled Room';
  };

  const getHeaderIcon = () => {
    switch (room.type) {
      case 'direct':
        return '👤';
      case 'group':
        return '👥';
      case 'channel':
        return '📢';
      default:
        return '💬';
    }
  };

  const getStatusIndicator = (msg: Message) => {
    if (msg.senderId !== user?.id) return null;
    
    switch (msg.status) {
      case 'new':
        return <span className='status-indicator status-new'>•</span>;
      case 'delivered':
        return <span className='status-indicator status-delivered'>✓</span>;
      case 'read':
        return <span className='status-indicator status-read'>✓✓</span>;
      case 'deleted':
        return null;
      default:
        return null;
    }
  };

  // Filter out deleted messages
  const displayMessages = messages[room.id]?.filter(msg => msg.status !== 'deleted') || [];

  if (loading) {
    return <div className='chat-window loading'>Loading messages...</div>;
  }

  return (
    <div className='chat-window'>
      <div className='chat-header'>
        <h3>
          <span className='header-icon'>{getHeaderIcon()}</span>
          {getHeaderTitle()}
        </h3>
        {room.type === 'direct' && (
          <span className='direct-chat-indicator'>Direct Message</span>
        )}
      </div>

      <div className='messages-container'>
        {displayMessages.length === 0 ? (
          <div className='no-messages'>
            <p>
              {room.type === 'direct'
                ? `Start a conversation with ${participantName}!`
                : 'No messages yet. Start the conversation!'}
            </p>
          </div>
        ) : (
          displayMessages.map((msg) => (
            <div
              key={msg.id}
              className={`message ${msg.senderName === user?.username ? 'sent' : 'received'}`}
              onContextMenu={(e) => {
                // Right-click to delete own messages
                if (msg.senderId === user?.id && msg.status !== 'deleted') {
                  e.preventDefault();
                  if (window.confirm('Delete this message?')) {
                    handleDeleteMessage(msg.id);
                  }
                }
              }}
            >
              {room.type !== 'direct' && (
                <div className='message-sender'>
                  <span>{msg.senderName || msg.senderId}</span>
                </div>
              )}
              <div className='message-content'>{msg.content}</div>
              <div className='message-time'>
                {formatTime(msg.createdAt)}
                {getStatusIndicator(msg)}
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      <form className='message-form' onSubmit={handleSendMessage}>
        <input
          type='text'
          value={messageText}
          onChange={(e) => setMessageText(e.target.value)}
          placeholder={room.type === 'direct' ? `Message ${participantName || 'user'}...` : 'Type a message...'}
        />
        <button type='submit' disabled={!messageText.trim()}>
          Send
        </button>
      </form>
    </div>
  );
}

export default ChatWindow;
