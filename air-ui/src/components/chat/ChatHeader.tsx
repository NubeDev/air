import { Button } from '@/components/ui/button';
import { Bug, ChevronDown, Settings, Zap, CheckCircle } from 'lucide-react';
import { type AIModel, type ModelInfo } from './ModelSelector';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu';

interface ChatHeaderProps {
  selectedModel: AIModel;
  onModelChange: (model: AIModel) => void;
  models?: ModelInfo[];
  health?: Record<string, { connected: boolean; error?: string }>;
  rawAIMode?: boolean;
  onToggleRawMode?: (value: boolean) => void;
  onToggleDebug?: () => void;
}

export function ChatHeader({
  selectedModel,
  onModelChange,
  models = [],
  health = {},
  rawAIMode = false,
  onToggleRawMode,
  onToggleDebug,
}: ChatHeaderProps) {
  const currentProvider = (selectedModel || '').split(':')[0] || 'ollama';
  const isConnected = health[currentProvider]?.connected ?? false;

  return (
    <div className="flex-shrink-0 bg-white border-b border-gray-100">
      <div className="max-w-4xl mx-auto px-6 py-4">
        <div className="flex items-center justify-center">
          {/* Single Dropdown Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="h-10 px-4 space-x-2">
                <div className="flex items-center space-x-2">
                  <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-primary' : 'bg-destructive'}`} />
                  <span className="text-sm font-medium">
                    {selectedModel.charAt(0).toUpperCase() + selectedModel.slice(1)}
                  </span>
                </div>
                <ChevronDown className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-64" align="center">
              <DropdownMenuLabel className="flex items-center space-x-2">
                <Settings className="h-4 w-4" />
                <span>AI Settings</span>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              
              {/* Model Selection */}
              <DropdownMenuLabel className="text-xs font-medium text-muted-foreground">
                Model Selection
              </DropdownMenuLabel>
              {models.map((m) => (
                <DropdownMenuItem
                  key={m.id}
                  onClick={() => onModelChange(m.id)}
                  className="flex items-center justify-between"
                >
                  <div className="flex items-center space-x-2">
                    <div className={`w-2 h-2 rounded-full ${health[m.provider]?.connected ? 'bg-primary' : 'bg-destructive'}`} />
                    <span className="capitalize">{m.provider}</span>
                    <span className="text-xs text-muted-foreground">{m.name}</span>
                  </div>
                  {selectedModel === m.id && (
                    <CheckCircle className="h-4 w-4 text-primary" />
                  )}
                </DropdownMenuItem>
              ))}
              
              <DropdownMenuSeparator />
              
              {/* Raw AI Toggle */}
              <DropdownMenuItem
                onClick={() => onToggleRawMode?.(!rawAIMode)}
                className="flex items-center justify-between"
              >
                <div className="flex items-center space-x-2">
                  <Zap className="h-4 w-4" />
                  <span>Raw AI Mode</span>
                </div>
                <div className={`w-4 h-4 rounded border-2 ${rawAIMode ? 'bg-primary border-primary' : 'border-muted'}`}>
                  {rawAIMode && <CheckCircle className="h-3 w-3 text-primary-foreground" />}
                </div>
              </DropdownMenuItem>
              
              <DropdownMenuSeparator />
              
              {/* Debug Button */}
              <DropdownMenuItem onClick={onToggleDebug} className="flex items-center space-x-2">
                <Bug className="h-4 w-4" />
                <span>Debug Console</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  );
}