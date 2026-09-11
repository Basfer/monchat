import { useState, useEffect, useRef } from 'react';
import { useAuthStore } from '../store/authStore';
import { useDirectChatStore } from '../store/directChatStore';
import { UserSearchResult } from '../types';

interface UserSearchModalProps {
  isOpen: boolean;
  onClose: () => void;
  onChatCreated: (roomId: string) => void;
}

function UserSearchModal({ isOpen, onClose, onChatCreated }: UserSearchModalProps) {
  const { user } = useAuthStore();
  const { searchQuery, searchResults, isSearching, setSearchQuery, setSearchResults, startSearch, createDirectChat } = useDirectChatStore();
  const [creating, setCreating] = useState<string | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen && searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) {
      setSearchQuery('');
      setSearchResults([]);
    }
  }, [isOpen, setSearchQuery, setSearchResults]);

  const handleSearch = async (query: string) => {
    if (user) {
      await startSearch(query, user.id);
    }
  };

  const handleCreateChat = async (targetUser: UserSearchResult) => {
    if (!user) return;
    
    setCreating(targetUser.id);
    try {
      const room = await createDirectChat(targetUser.id, user.id);
      onChatCreated(room.id);
      onClose();
    } catch (error) {
      console.error('Failed to create direct chat:', error);
    } finally {
      setCreating(null);
    }
  };

  const handleClose = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" onClick={handleClose} onKeyDown={handleKeyDown}>
      <div className="modal user-search-modal">
        <div className="modal-header">
          <h3>New Direct Message</h3>
          <button className="close-btn" onClick={onClose}>&times;</button>
        </div>

        <div className="search-container">
          <input
            ref={searchInputRef}
            type="text"
            className="search-input"
            placeholder="Search by username or email..."
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              handleSearch(e.target.value);
            }}
          />
        </div>

        <div className="search-results">
          {isSearching && (
            <div className="search-loading">
              <div className="spinner"></div>
              <span>Searching...</span>
            </div>
          )}

          {!isSearching && searchQuery && searchResults.length === 0 && (
            <div className="search-empty">
              <p>No users found</p>
            </div>
          )}

          {!isSearching && !searchQuery && (
            <div className="search-hint">
              <p>Type to search for users</p>
            </div>
          )}

          {searchResults.map((result) => (
            <div key={result.id} className="search-result-item">
              <div className="user-avatar">
                <span className="avatar-letter">{result.username.charAt(0).toUpperCase()}</span>
              </div>
              <div className="user-info">
                <div className="username">{result.username}</div>
                <div className="email">{result.email}</div>
              </div>
              <button
                className="start-chat-btn"
                onClick={() => handleCreateChat(result)}
                disabled={creating === result.id}
              >
                {creating === result.id ? (
                  <>
                    <span className="spinner-small"></span>
                    Creating...
                  </>
                ) : (
                  'Message'
                )}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default UserSearchModal;
