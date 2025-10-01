import { } from 'react';
import { MainLayout } from '@/components/layout/MainLayout';
import { ChatWindow } from '@/components/chat/ChatWindowNew';

export function ChatPage() {
  return (
    <MainLayout>
      <div className="h-full">
        <ChatWindow />
      </div>
    </MainLayout>
  );
}


