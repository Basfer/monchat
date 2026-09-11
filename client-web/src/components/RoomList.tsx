import { useState } from 'react';
import { Room } from '../types';

interface RoomListProps {
  rooms: Room[];
  onSelectRoom: (room: Room) => void;
  onCreateRoom: (name: string, description: string, type: string) => void;
  onOpenDirectMessageModal: () => void;
}

function RoomList({ rooms, onSelectRoom, onCreateRoom, onOpenDirectMessageModal }: RoomListProps) {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newRoomName, setNewRoomName] = useState('');
  const [newRoomDescription, setNewRoomDescription] = useState('');
  const [newRoomType, setNewRoomType] = useState<'group' | 'channel'>('group');

  const handleCreateRoom = (e: React.FormEvent) => {
    e.preventDefault();
    if (newRoomName.trim()) {
      onCreateRoom(newRoomName, newRoomDescription, newRoomType);
      setNewRoomName('');
      setNewRoomDescription('');
      setShowCreateModal(false);
    }
  };

  const getRoomIcon = (type: string) => {
    switch (type) {
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

  // Разделяем комнаты на прямые чаты и остальные
  const directChats = rooms.filter((room) => room.type === 'direct');
  const otherRooms = rooms.filter((room) => room.type !== 'direct');

  const getRoomDisplayName = (room: Room) => {
    if (room.type === 'direct' && room.participant) {
      return room.participant.username;
    }
    return room.name || 'Untitled Room';
  };

  return (
    <div className='room-list'>
      <div className='room-list-header'>
        <h2>Rooms</h2>
        <div className='create-room-buttons'>
          <button className='btn-direct-message' onClick={onOpenDirectMessageModal} title='New Direct Message'>
            + DM
          </button>
          <button className='btn-create-room' onClick={() => setShowCreateModal(true)}>
            + Room
          </button>
        </div>
      </div>

      {/* Секция прямых чатов */}
      {directChats.length > 0 && (
        <div className='room-section'>
          <h3 className='section-title'>Direct Messages</h3>
          <ul className='rooms direct-chats'>
            {directChats.map((room) => (
              <li
                key={room.id}
                className={`room-item ${room.type}`}
                onClick={() => onSelectRoom(room)}
              >
                <span className='room-icon'>{getRoomIcon(room.type)}</span>
                <div className='room-info'>
                  <span className='room-name'>{getRoomDisplayName(room)}</span>
                  {room.lastMessage && (
                    <span className='last-message'>{room.lastMessage.content.substring(0, 30)}...</span>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Секция групповых чатов и каналов */}
      {otherRooms.length > 0 && (
        <div className='room-section'>
          <h3 className='section-title'>Teams & Channels</h3>
          <ul className='rooms'>
            {otherRooms.map((room) => (
              <li
                key={room.id}
                className={`room-item ${room.type}`}
                onClick={() => onSelectRoom(room)}
              >
                <span className='room-icon'>{getRoomIcon(room.type)}</span>
                <div className='room-info'>
                  <span className='room-name'>{getRoomDisplayName(room)}</span>
                  {room.lastMessage && (
                    <span className='last-message'>{room.lastMessage.content.substring(0, 30)}...</span>
                  )}
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Пустое состояние, если нет комнат */}
      {rooms.length === 0 && (
        <div className='empty-rooms'>
          <p>No rooms yet. Create a room or start a direct message!</p>
        </div>
      )}

      {showCreateModal && (
        <div className='modal-overlay'>
          <div className='modal'>
            <h3>Create New Room</h3>
            <form onSubmit={handleCreateRoom}>
              <div className='form-group'>
                <label>Name</label>
                <input
                  type='text'
                  value={newRoomName}
                  onChange={(e) => setNewRoomName(e.target.value)}
                  required
                />
              </div>
              <div className='form-group'>
                <label>Description</label>
                <input
                  type='text'
                  value={newRoomDescription}
                  onChange={(e) => setNewRoomDescription(e.target.value)}
                />
              </div>
              <div className='form-group'>
                <label>Type</label>
                <select
                  value={newRoomType}
                  onChange={(e) => setNewRoomType(e.target.value as 'group' | 'channel')}
                >
                  <option value='group'>Group Chat</option>
                  <option value='channel'>Channel</option>
                </select>
              </div>
              <div className='modal-actions'>
                <button type='button' onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type='submit'>Create</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default RoomList;
