import React from 'react';
import { Button } from '@/components/ui/button';
import { Bug } from 'lucide-react';
import { ModelSelector, type AIModel } from './ModelSelector';

interface ModelStatus {
  connected: boolean;
  error?: string;
}

interface ChatHeaderProps {
  selectedModel: AIModel;
  onModelChange: (model: AIModel) => void;
  modelStatus: Record<AIModel, ModelStatus | undefined>;
  wsConnected?: boolean;
  rawAIMode?: boolean;
  onToggleRawMode?: (value: boolean) => void;
  showDebug?: boolean;
  onToggleDebug?: () => void;
}

export function ChatHeader({
  selectedModel,
  onModelChange,
  modelStatus,
  wsConnected = true,
  rawAIMode = false,
  onToggleRawMode,
  showDebug = false,
  onToggleDebug,
}: ChatHeaderProps) {
  return (
    <div className="flex-shrink-0 border-b bg-white px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-xs text-gray-500">
            {wsConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>

        <div className="flex items-center space-x-3">
          <ModelSelector
            selectedModel={selectedModel}
            onModelChange={onModelChange}
            modelStatus={modelStatus}
          />

          <div className="flex items-center gap-2 p-2 bg-muted rounded-lg">
            <span className="text-xs font-medium">Raw AI:</span>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={rawAIMode}
                onChange={(e) => onToggleRawMode?.(e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-10 h-5 bg-gray-200 rounded-full peer peer-checked:bg-blue-600 after:content-[''] after:absolute after:-ml-8 after:w-4 after:h-4 after:bg-white after:rounded-full after:translate-x-1 peer-checked:after:translate-x-6 after:transition" />
            </label>
          </div>

          <Button
            variant="ghost"
            size="sm"
            onClick={onToggleDebug}
            className="text-gray-500 hover:text-gray-700"
          >
            <Bug className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}


