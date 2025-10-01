import React from 'react';

interface FloatingDotProps {
  visible: boolean;
  text?: string;
  status?: 'waiting' | 'error' | 'ok';
}

export function FloatingDot({ visible, text = 'Waiting on backend…', status = 'waiting' }: FloatingDotProps) {
  if (!visible) return null;

  const color = status === 'error' ? 'bg-red-500' : status === 'ok' ? 'bg-green-500' : 'bg-blue-500';

  return (
    <div className="fixed bottom-24 right-5 z-40 flex items-center gap-2 select-none">
      <div className={`w-3 h-3 rounded-full ${color} animate-pulse`} />
      <span className="text-xs text-gray-600 bg-white/90 backdrop-blur px-2 py-1 rounded shadow">
        {text}
      </span>
    </div>
  );
}


