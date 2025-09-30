import React, { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Send, Paperclip, X, Database, BarChart3, HelpCircle } from 'lucide-react';

interface ChatInputProps {
  onSend: (message: string) => void;
  onFileAttach?: (file: File) => void;
  disabled?: boolean;
  placeholder?: string;
  attachedFiles?: File[];
  onRemoveFile?: (index: number) => void;
}

export function ChatInput({ 
  onSend, 
  onFileAttach, 
  disabled = false, 
  placeholder = "Type your message...",
  attachedFiles = [],
  onRemoveFile
}: ChatInputProps) {
  const [message, setMessage] = useState('');
  const [showCommands, setShowCommands] = useState(false);
  const [selectedCommandIndex, setSelectedCommandIndex] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Define available slash commands
  const commands = [
    {
      command: '/load',
      description: 'Load an existing dataset',
      icon: Database,
      usage: '/load <filename>',
      examples: ['/load <filename>']
    },
    {
      command: '/analyze',
      description: 'Analyze current data',
      icon: BarChart3,
      usage: '/analyze <query>',
      examples: ['/analyze <your question>']
    },
    {
      command: '/help',
      description: 'Show available commands',
      icon: HelpCircle,
      usage: '/help',
      examples: []
    }
  ];

  // Filter commands based on current input
  const filteredCommands = commands.filter(cmd => 
    cmd.command.toLowerCase().includes(message.toLowerCase().replace('/', '')) ||
    cmd.description.toLowerCase().includes(message.toLowerCase().replace('/', ''))
  );

  // Handle slash command detection
  useEffect(() => {
    if (message.startsWith('/') && message.length > 0) {
      setShowCommands(true);
      setSelectedCommandIndex(0);
    } else {
      setShowCommands(false);
    }
  }, [message]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() && !disabled) {
      onSend(message.trim());
      setMessage('');
      setShowCommands(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (showCommands) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedCommandIndex(prev => 
          prev < filteredCommands.length - 1 ? prev + 1 : 0
        );
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedCommandIndex(prev => 
          prev > 0 ? prev - 1 : filteredCommands.length - 1
        );
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (filteredCommands[selectedCommandIndex]) {
          const selectedCmd = filteredCommands[selectedCommandIndex];
          if (selectedCmd.examples.length > 0) {
            setMessage(selectedCmd.examples[0]);
          } else {
            setMessage(selectedCmd.command + ' ');
          }
          setShowCommands(false);
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        setShowCommands(false);
      }
    } else if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  const handleCommandSelect = (command: typeof commands[0]) => {
    if (command.examples.length > 0) {
      setMessage(command.examples[0]);
    } else {
      setMessage(command.command + ' ');
    }
    setShowCommands(false);
    textareaRef.current?.focus();
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0 && onFileAttach) {
      onFileAttach(files[0]);
    }
  };

  const handleAttachClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="space-y-2">
      {/* Attached Files */}
      {attachedFiles.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {attachedFiles.map((file, index) => (
            <div key={index} className="flex items-center gap-2 bg-blue-50 text-blue-700 px-3 py-1 rounded-full text-sm">
              <Paperclip className="h-3 w-3" />
              <span className="truncate max-w-32">{file.name}</span>
              {onRemoveFile && (
                <button
                  type="button"
                  onClick={() => onRemoveFile(index)}
                  className="hover:bg-blue-100 rounded-full p-0.5"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Command Autocomplete Dropdown */}
      {showCommands && filteredCommands.length > 0 && (
        <div className="absolute bottom-full left-0 right-0 mb-2 bg-white border border-gray-200 rounded-lg shadow-lg z-50 max-h-60 overflow-y-auto">
          {filteredCommands.map((command, index) => {
            const IconComponent = command.icon;
            return (
              <div
                key={command.command}
                className={`flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 ${
                  index === selectedCommandIndex 
                    ? 'bg-blue-50 text-blue-700' 
                    : 'hover:bg-gray-50'
                }`}
                onClick={() => handleCommandSelect(command)}
              >
                <IconComponent className="h-4 w-4 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm font-medium">{command.command}</span>
                    <span className="text-xs text-gray-500">{command.usage}</span>
                  </div>
                  <p className="text-xs text-gray-600 mt-1">{command.description}</p>
                  {command.examples.length > 0 && (
                    <div className="mt-1">
                      <p className="text-xs text-gray-500">Examples:</p>
                      {command.examples.slice(0, 2).map((example, idx) => (
                        <p key={idx} className="text-xs font-mono text-gray-600">{example}</p>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
      
      <form onSubmit={handleSubmit} className="relative">
        <div className="flex items-end gap-3 bg-white border border-gray-300 rounded-2xl px-4 py-3 shadow-sm focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-blue-500">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={handleAttachClick}
            disabled={disabled}
            className="rounded-full w-8 h-8 p-0 text-gray-500 hover:text-gray-700 hover:bg-gray-100"
          >
            <Paperclip className="h-4 w-4" />
          </Button>
          
          <textarea
            ref={textareaRef}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            rows={1}
            className="flex-1 resize-none border-none outline-none text-gray-900 placeholder-gray-500 bg-transparent max-h-32"
            style={{
              minHeight: '24px',
              height: 'auto',
            }}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement;
              target.style.height = 'auto';
              target.style.height = target.scrollHeight + 'px';
            }}
          />
          
          <Button 
            type="submit" 
            disabled={disabled || (!message.trim() && attachedFiles.length === 0)} 
            size="sm"
            className="rounded-full w-8 h-8 p-0 bg-blue-500 hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed"
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
        
        <input
          ref={fileInputRef}
          type="file"
          onChange={handleFileSelect}
          className="hidden"
          accept=".csv,.json,.xlsx,.xls,.parquet"
        />
      </form>
    </div>
  );
}
