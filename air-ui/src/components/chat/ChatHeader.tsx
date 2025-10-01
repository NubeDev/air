import { Button } from '@/components/ui/button';
import { Bug, ChevronDown, Settings, Zap, CheckCircle } from 'lucide-react';
import { type AIModel } from './ModelSelector';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface ModelStatus {
  connected: boolean;
  error?: string;
}

interface ChatHeaderProps {
  selectedModel: AIModel;
  onModelChange: (model: AIModel) => void;
  modelStatus: Record<AIModel, ModelStatus | undefined>;
  rawAIMode?: boolean;
  onToggleRawMode?: (value: boolean) => void;
  onToggleDebug?: () => void;
}

export function ChatHeader({
  selectedModel,
  onModelChange,
  modelStatus,
  rawAIMode = false,
  onToggleRawMode,
  onToggleDebug,
}: ChatHeaderProps) {
  const currentStatus = modelStatus[selectedModel];
  const isConnected = currentStatus?.connected ?? false;

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
              {(['openai', 'llama', 'sqlcoder'] as AIModel[]).map((model) => {
                const status = modelStatus[model];
                return (
                  <DropdownMenuItem
                    key={model}
                    onClick={() => onModelChange(model)}
                    className="flex items-center justify-between"
                  >
                    <div className="flex items-center space-x-2">
                      <div className={`w-2 h-2 rounded-full ${status?.connected ? 'bg-primary' : 'bg-destructive'}`} />
                      <span className="capitalize">{model}</span>
                    </div>
                    {selectedModel === model && (
                      <CheckCircle className="h-4 w-4 text-primary" />
                    )}
                  </DropdownMenuItem>
                );
              })}
              
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