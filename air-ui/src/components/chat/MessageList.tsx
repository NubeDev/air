import React, { RefObject } from 'react';
import type { ChatMessage as MessageType } from '@/types/api';
import { Message } from './Message';

interface MessageListProps {
  messages: MessageType[];
  isTyping?: boolean;
  typingMessage?: string;
  scrollAreaRef: RefObject<HTMLDivElement | null>;
  onScroll?: () => void;
  footer?: React.ReactNode;
  onCancelTyping?: () => void;
}

export function MessageList({ messages, isTyping, typingMessage, scrollAreaRef, onScroll, footer, onCancelTyping }: MessageListProps) {
  return (
    <div className="h-full overflow-y-auto" ref={scrollAreaRef} onScroll={onScroll}>
      <div className="max-w-4xl mx-auto px-4 py-6">
        <div className="space-y-6">
          {messages.map((message) => (
            <Message key={message.id} message={message} />
          ))}

          {isTyping && (
            <div className="flex items-center space-x-3 text-gray-500">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }} />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }} />
              </div>
              <span className="text-sm">{typingMessage || 'AI is typing...'}</span>
              {onCancelTyping && (
                <button
                  onClick={onCancelTyping}
                  className="ml-2 text-xs px-2 py-1 rounded border border-gray-300 hover:bg-gray-100"
                >
                  Cancel
                </button>
              )}
            </div>
          )}

          {footer && (
            <div className="mt-2">
              {footer}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}


