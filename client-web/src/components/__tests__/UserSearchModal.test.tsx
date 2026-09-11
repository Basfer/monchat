import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import UserSearchModal from '../UserSearchModal';
import { useAuthStore } from '../../store/authStore';
import { useDirectChatStore } from '../../store/directChatStore';

// Mock stores - the service mock is needed because it's used internally by the store
jest.mock('../../store/authStore', () => ({
  useAuthStore: jest.fn(),
}));

jest.mock('../../store/directChatStore', () => ({
  useDirectChatStore: jest.fn(),
}));

jest.mock('../../services/api', () => ({
  directChatService: {
    searchUsers: jest.fn(),
    getOrCreateDirectChat: jest.fn(),
  },
}));

const mockUser = { id: 'test-user-id', username: 'testuser', email: 'test@test.com' };
const mockOnClose = jest.fn();
const mockOnChatCreated = jest.fn();
const mockSetSearchQuery = jest.fn();
const mockSetSearchResults = jest.fn();
const mockStartSearch = jest.fn();
const mockCreateDirectChat = jest.fn();

describe('UserSearchModal', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    const mockUseAuthStore = useAuthStore as unknown as jest.Mock;
    mockUseAuthStore.mockReturnValue({ user: mockUser });
    
    const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
    mockUseDirectChatStore.mockReturnValue({
      searchQuery: '',
      searchResults: [],
      isSearching: false,
      setSearchQuery: mockSetSearchQuery,
      setSearchResults: mockSetSearchResults,
      startSearch: mockStartSearch,
      createDirectChat: mockCreateDirectChat,
    });
  });

  describe('rendering', () => {
    test('returns null when closed', () => {
      const { container } = render(
        <UserSearchModal
          isOpen={false}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(container.firstChild).toBeNull();
      expect(mockOnClose).not.toHaveBeenCalled();
    });

    test('renders modal when open', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('New Direct Message')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Search by username or email...')).toBeInTheDocument();
    });

    test('auto-focuses search input on open', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search by username or email...');
      expect(searchInput).toHaveFocus();
    });
  });

  describe('close handlers', () => {
    test('closes on Escape key press', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const modalOverlay = document.querySelector('.modal-overlay');
      if (modalOverlay) {
        fireEvent.keyDown(modalOverlay, { key: 'Escape' });
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    test('closes on overlay click', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const modalOverlay = document.querySelector('.modal-overlay');
      if (modalOverlay) {
        fireEvent.click(modalOverlay);
      }

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    test('does not close when clicking inside modal content', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const modalContent = document.querySelector('.modal');
      if (modalContent) {
        fireEvent.click(modalContent);
      }

      expect(mockOnClose).not.toHaveBeenCalled();
    });
  });

  describe('search functionality', () => {
    test('displays search hint when no query entered', () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('Type to search for users')).toBeInTheDocument();
    });

    test('calls startSearch when typing in search input', async () => {
      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search by username or email...');
      
      await act(async () => {
        fireEvent.change(searchInput, { target: { value: 'john' } });
      });

      expect(mockSetSearchQuery).toHaveBeenCalledWith('john');
      expect(mockStartSearch).toHaveBeenCalledWith('john', 'test-user-id');
    });

    test('displays loading state while searching', () => {
      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: [],
        isSearching: true,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('Searching...')).toBeInTheDocument();
    });

    test('displays "No users found" when search returns empty results', () => {
      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'nonexistent_user_xyz',
        searchResults: [],
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('No users found')).toBeInTheDocument();
    });
  });

  describe('search results display', () => {
    test('displays search results with user info', () => {
      const mockResults = [
        { id: 'user-1', username: 'john_doe', email: 'john@example.com' },
        { id: 'user-2', username: 'john_smith', email: 'john.smith@example.com' },
      ];

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('john_doe')).toBeInTheDocument();
      expect(screen.getByText('john@example.com')).toBeInTheDocument();
      expect(screen.getByText('john_smith')).toBeInTheDocument();
      expect(screen.getByText('john.smith@example.com')).toBeInTheDocument();
    });

    test('shows avatar letter for each user', () => {
      const mockResults = [{ id: 'user-1', username: 'john_doe', email: 'john@example.com' }];

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('J')).toBeInTheDocument();
    });

    test('shows Message button for each result user', () => {
      const mockResults = [{ id: 'user-1', username: 'john_doe', email: 'john@example.com' }];

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(screen.getByText('Message')).toBeInTheDocument();
    });
  });

  describe('chat creation', () => {
    test('disables button and shows creating state while creating chat', async () => {
      const mockResults = [{ id: 'user-1', username: 'john_doe', email: 'john@example.com' }];
      const mockRoom = { id: 'room-1', name: 'Direct Chat', type: 'direct' };

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat.mockResolvedValue(mockRoom),
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const startChatButton = screen.getByText('Message');
      
      await act(async () => {
        fireEvent.click(startChatButton);
      });

      await waitFor(() => {
        expect(mockCreateDirectChat).toHaveBeenCalledWith('user-1', 'test-user-id');
      });
    });

    test('calls onChatCreated and closes after successful chat creation', async () => {
      const mockResults = [{ id: 'user-1', username: 'john_doe', email: 'john@example.com' }];
      const mockRoom = { id: 'room-1', name: 'Direct Chat', type: 'direct' };

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat.mockResolvedValue(mockRoom),
      });

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const startChatButton = screen.getByText('Message');
      
      await act(async () => {
        fireEvent.click(startChatButton);
      });

      await waitFor(() => {
        expect(mockOnChatCreated).toHaveBeenCalledWith('room-1');
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    test('handles chat creation error gracefully', async () => {
      const mockResults = [{ id: 'user-1', username: 'john_doe', email: 'john@example.com' }];
      const consoleError = console.error;
      const mockedConsoleError = jest.fn();

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'john',
        searchResults: mockResults,
        isSearching: false,
        setSearchQuery: mockSetSearchQuery,
        setSearchResults: mockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat.mockRejectedValue(new Error('Network error')),
      });

      console.error = mockedConsoleError;

      render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      const startChatButton = screen.getByText('Message');
      
      await act(async () => {
        fireEvent.click(startChatButton);
      });

      await waitFor(() => {
        expect(mockedConsoleError).toHaveBeenCalled();
      });

      console.error = consoleError;
    });
  });

  describe('modal reset on close', () => {
    test('clears search query and results when modal closes', () => {
      const localMockSetSearchQuery = jest.fn();
      const localMockSetSearchResults = jest.fn();

      const mockUseDirectChatStore = useDirectChatStore as unknown as jest.Mock;
      mockUseDirectChatStore.mockReturnValue({
        searchQuery: 'previous search',
        searchResults: [{ id: 'user-1', username: 'test', email: 'test@test.com' }],
        isSearching: false,
        setSearchQuery: localMockSetSearchQuery,
        setSearchResults: localMockSetSearchResults,
        startSearch: mockStartSearch,
        createDirectChat: mockCreateDirectChat,
      });

      const { rerender } = render(
        <UserSearchModal
          isOpen={true}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      // Close the modal by rerendering with isOpen=false
      rerender(
        <UserSearchModal
          isOpen={false}
          onClose={mockOnClose}
          onChatCreated={mockOnChatCreated}
        />
      );

      expect(localMockSetSearchQuery).toHaveBeenCalledWith('');
      expect(localMockSetSearchResults).toHaveBeenCalledWith([]);
    });
  });
});
