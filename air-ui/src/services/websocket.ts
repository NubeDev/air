export interface WebSocketMessage {
  type: string;
  channel?: string;
  payload: Record<string, any>;
  timestamp: string;
  user_id?: string;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private messageHandlers: Map<string, (message: WebSocketMessage) => void> = new Map();

  constructor(private url: string) {}

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        console.log('Connecting to WebSocket:', this.url);
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
          console.log('✅ WebSocket connected successfully');
          this.reconnectAttempts = 0;
          window.dispatchEvent(new CustomEvent('ws-connected'));
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            // Handle multiple messages that might be concatenated
            const messages = event.data.toString().split('\n').filter((line: string) => line.trim());
            
            messages.forEach((messageStr: string) => {
              try {
                const message: WebSocketMessage = JSON.parse(messageStr);
                this.handleMessage(message);
              } catch (parseError) {
                console.error('Failed to parse individual WebSocket message:', parseError, 'Message:', messageStr);
              }
            });
          } catch (error) {
            console.error('Failed to process WebSocket message:', error);
          }
        };

        this.ws.onclose = (event) => {
          console.log('WebSocket disconnected:', event.code, event.reason);
          window.dispatchEvent(new CustomEvent('ws-disconnected'));
          if (event.code !== 1000) { // Not a normal closure
            this.reconnect();
          }
        };

        this.ws.onerror = (error) => {
          console.error('WebSocket error:', error);
          console.error('WebSocket readyState:', this.ws?.readyState);
          console.error('WebSocket url:', this.ws?.url);
          // Don't reject immediately, let the close handler handle reconnection
        };
      } catch (error) {
        console.error('Failed to create WebSocket:', error);
        reject(error);
      }
    });
  }

  private reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(`Reconnecting... attempt ${this.reconnectAttempts}`);
      
      setTimeout(() => {
        this.connect().catch(console.error);
      }, this.reconnectDelay * this.reconnectAttempts);
    }
  }

  private handleMessage(message: WebSocketMessage) {
    // Call specific handler if registered
    const handler = this.messageHandlers.get(message.type);
    if (handler) {
      handler(message);
    }

    // Call general handler
    const generalHandler = this.messageHandlers.get('*');
    if (generalHandler) {
      generalHandler(message);
    }
  }

  onMessage(type: string, handler: (message: WebSocketMessage) => void) {
    this.messageHandlers.set(type, handler);
  }

  sendMessage(message: Partial<WebSocketMessage>) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        ...message,
        timestamp: new Date().toISOString(),
      }));
    } else {
      console.error('WebSocket not connected');
    }
  }

  // Alias for convenience
  send(message: Partial<WebSocketMessage>) {
    this.sendMessage(message);
  }

  subscribe(channel: string) {
    this.sendMessage({
      type: 'subscribe',
      payload: { channel },
    });
  }

  unsubscribe(channel: string) {
    this.sendMessage({
      type: 'unsubscribe',
      payload: { channel },
    });
  }

  analyzeFile(fileId: string, query: string, model: string = 'llama') {
    this.sendMessage({
      type: 'file_analysis',
      payload: {
        file_id: fileId,
        query,
        model,
      },
    });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Singleton instance
export const wsService = new WebSocketService('ws://localhost:9000/v1/ws/');
