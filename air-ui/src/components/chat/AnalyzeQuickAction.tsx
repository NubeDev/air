import React from 'react';
import { Button } from '@/components/ui/button';
import { BarChart3 } from 'lucide-react';

interface AnalyzeQuickActionProps {
  onAnalyze: () => void;
  disabled?: boolean;
}

export function AnalyzeQuickAction({ onAnalyze, disabled }: AnalyzeQuickActionProps) {
  return (
    <div
      role="group"
      aria-label="Quick action"
      className="inline-flex items-center gap-3 bg-blue-50 border border-blue-200 rounded-xl px-3 py-2 shadow-sm"
    >
      <div className="flex items-center text-blue-700 text-sm font-medium">
        <BarChart3 className="h-4 w-4 mr-2" />
        Analyze dataset
      </div>
      <Button
        size="sm"
        onClick={onAnalyze}
        disabled={disabled}
        className="h-8"
        aria-label="Analyze dataset"
        title="Analyze dataset"
      >
        Analyze
      </Button>
    </div>
  );
}


