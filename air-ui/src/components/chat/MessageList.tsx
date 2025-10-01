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
}

export function MessageList({ messages, isTyping, typingMessage, scrollAreaRef, onScroll, footer }: MessageListProps) {
  return (
    <div className="h-full overflow-y-auto" ref={scrollAreaRef} onScroll={onScroll}>
      <div className="max-w-4xl mx-auto px-4 py-6">
        <div className="space-y-6">
          {messages.map((message) => (
            <Message key={message.id} message={message} />
          ))}

          {isTyping && (
            <div className="flex items-center space-x-2 text-gray-500">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }} />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }} />
              </div>
              <span className="text-sm">{typingMessage || 'AI is typing...'}</span>
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


