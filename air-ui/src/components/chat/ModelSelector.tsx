import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

export type AIModel = string; // model ID like "openai:gpt-4o-mini" or "ollama:llama3:latest"

interface ProviderStatus {
  connected: boolean;
  error?: string;
}

export interface ModelInfo {
  id: string;
  provider: string;
  name: string;
  capabilities: string[];
}

interface ModelSelectorProps {
  selectedModel: AIModel;
  onModelChange: (model: AIModel) => void;
  models: ModelInfo[];
  health: Record<string, ProviderStatus>;
}

export function ModelSelector({ selectedModel, onModelChange, models, health }: ModelSelectorProps) {
  const getStatusIcon = (provider: string) => {
    const status = health[provider];
    if (!status) return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    if (status.connected) return <CheckCircle className="h-4 w-4 text-green-500" />;
    if (status.error) return <XCircle className="h-4 w-4 text-red-500" />;
    return <AlertCircle className="h-4 w-4 text-yellow-500" />;
  };

  return (
    <div className="flex items-center space-x-3">
      <span className="text-sm font-medium text-gray-600">Model</span>
      <Select value={selectedModel} onValueChange={onModelChange}>
        <SelectTrigger className="w-52 h-9 border-gray-200 focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {models.map((m) => (
            <SelectItem key={m.id} value={m.id}>
              <div className="flex items-center gap-2">
                <span className="capitalize">{m.provider}</span>
                <span className="text-xs text-muted-foreground">{m.name}</span>
                {getStatusIcon(m.provider)}
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}