import { WSMessage } from '../types';

type MessageCallback = (message: WSMessage) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private onMessageCallbacks: MessageCallback[] = [];
  private userId: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;

  constructor(userId: string) {
    this.userId = userId;
  }

  connect(onMessage: MessageCallback) {
    this.onMessageCallbacks.push(onMessage);
    
    if (!this.ws || this.ws.readyState === WebSocket.CLOSED || this.ws.readyState === WebSocket.CONNECTING) {
      this.connectWebSocket();
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  sendMessage(content: string, roomId: string) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WSMessage = {
        type: 'message',
        roomID: roomId,
        userID: this.userId,
        content,
        timestamp: new Date().toISOString(),
      };
      this.ws.send(JSON.stringify(message));
    }
  }

  private connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const url = `${protocol}//${host}/ws?userId=${this.userId}`;
    console.log('Connecting to WebSocket:', url);
    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.reconnectDelay = 1000;
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data);
        this.onMessageCallbacks.forEach(callback => callback(message));
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.tryReconnect();
    };
  }

  private tryReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
      this.reconnectAttempts++;
      
      setTimeout(() => {
        console.log(`Attempting to reconnect... (attempt ${this.reconnectAttempts})`);
        this.connectWebSocket();
      }, delay);
    }
  }

  hasCallback(onMessage: MessageCallback): boolean {
    return this.onMessageCallbacks.includes(onMessage);
  }

  removeCallback(onMessage: MessageCallback) {
    const index = this.onMessageCallbacks.indexOf(onMessage);
    if (index > -1) {
      this.onMessageCallbacks.splice(index, 1);
    }
    
    // Disconnect if no more callbacks
    if (this.onMessageCallbacks.length === 0 && this.ws) {
      this.disconnect();
    }
  }
}
