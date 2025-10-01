import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { CheckCircle, XCircle, AlertCircle } from 'lucide-react';

export type AIModel = 'llama' | 'openai' | 'sqlcoder';

interface ModelStatus {
  connected: boolean;
  error?: string;
}

interface ModelSelectorProps {
  selectedModel: AIModel;
  onModelChange: (model: AIModel) => void;
  modelStatus: Record<AIModel, ModelStatus | undefined>;
}

export function ModelSelector({ selectedModel, onModelChange, modelStatus }: ModelSelectorProps) {
  const getStatusIcon = (status: ModelStatus | undefined) => {
    if (!status) {
      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    }
    if (status.connected) {
      return <CheckCircle className="h-4 w-4 text-green-500" />;
    } else if (status.error) {
      return <XCircle className="h-4 w-4 text-red-500" />;
    } else {
      return <AlertCircle className="h-4 w-4 text-yellow-500" />;
    }
  };

  const getStatusText = (status: ModelStatus | undefined) => {
    if (!status) {
      return 'Loading...';
    }
    if (status.connected) {
      return 'Connected';
    } else if (status.error) {
      return status.error;
    } else {
      return 'No valid connection';
    }
  };

  return (
    <div className="flex items-center space-x-3">
      <span className="text-sm font-medium text-gray-600">Model</span>
      <Select value={selectedModel} onValueChange={onModelChange}>
        <SelectTrigger className="w-36 h-9 border-gray-200 focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="llama">
            <div className="flex items-center gap-2">
              <span>Llama</span>
              {getStatusIcon(modelStatus.llama)}
            </div>
          </SelectItem>
          <SelectItem value="openai">
            <div className="flex items-center gap-2">
              <span>OpenAI</span>
              {getStatusIcon(modelStatus.openai)}
            </div>
          </SelectItem>
          <SelectItem value="sqlcoder">
            <div className="flex items-center gap-2">
              <span>SQLCoder</span>
              {getStatusIcon(modelStatus.sqlcoder)}
            </div>
          </SelectItem>
        </SelectContent>
      </Select>

    </div>
  );
}