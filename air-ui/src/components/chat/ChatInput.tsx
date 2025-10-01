import React, { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Send, Paperclip, X, Database, BarChart3, HelpCircle, FileText } from 'lucide-react';
import { chatApi } from '@/services/chatApi';

interface ChatInputProps {
  onSend: (message: string) => void;
  onFileAttach?: (file: File) => void;
  disabled?: boolean;
  placeholder?: string;
  attachedFiles?: File[];
  onRemoveFile?: (index: number) => void;
  onAtCommand?: (command: string, args: string[]) => void;
}

interface AtCommand {
  command: string;
  description: string;
  icon: React.ComponentType<any>;
  usage: string;
  examples: string[];
}

interface FileItem {
  file_id: string;
  filename: string;
  file_type: string;
  file_size: number;
}

interface DatasourceItem {
  id: string;
  name: string;
  type: string;
  connected: boolean;
}

export function ChatInput({ 
  onSend, 
  onFileAttach, 
  disabled = false, 
  placeholder = "Type your message...",
  attachedFiles = [],
  onRemoveFile,
  onAtCommand
}: ChatInputProps) {
  const [message, setMessage] = useState('');
  const [showCommands, setShowCommands] = useState(false);
  const [selectedCommandIndex, setSelectedCommandIndex] = useState(0);
  const [showAtCommands, setShowAtCommands] = useState(false);
  const [selectedAtCommandIndex, setSelectedAtCommandIndex] = useState(0);
  const [atCommandType, setAtCommandType] = useState<'files' | 'db' | 'load-file' | null>(null);
  const [atCommandQuery, setAtCommandQuery] = useState('');
  const [availableFiles, setAvailableFiles] = useState<FileItem[]>([]);
  const [availableDatasources, setAvailableDatasources] = useState<DatasourceItem[]>([]);
  const [isLoadingFiles, setIsLoadingFiles] = useState(false);
  const [isLoadingDatasources, setIsLoadingDatasources] = useState(false);
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

  // Define available @ commands
  const atCommands: AtCommand[] = [
    {
      command: '@files',
      description: 'List available files',
      icon: FileText,
      usage: '@files',
      examples: ['@files']
    },
    {
      command: '@recent',
      description: 'Recently used files',
      icon: FileText,
      usage: '@recent',
      examples: ['@recent']
    },
    {
      command: '@db',
      description: 'List available databases',
      icon: Database,
      usage: '@db',
      examples: ['@db']
    },
    {
      command: '@load-file',
      description: 'Load a specific file',
      icon: FileText,
      usage: '@load-file/<filename>',
      examples: ['@load-file/abc.csv']
    }
  ];

  // Filter commands based on current input
  const filteredCommands = commands.filter(cmd => 
    cmd.command.toLowerCase().includes(message.toLowerCase().replace('/', '')) ||
    cmd.description.toLowerCase().includes(message.toLowerCase().replace('/', ''))
  );

  // Filter @ commands based on current input
  const filteredAtCommands = atCommands.filter(cmd => 
    cmd.command.toLowerCase().includes(message.toLowerCase().replace('@', '')) ||
    cmd.description.toLowerCase().includes(message.toLowerCase().replace('@', ''))
  );

  // Load available files
  const loadFiles = async () => {
    setIsLoadingFiles(true);
    try {
      const response = await chatApi.getUploadedFiles();
      setAvailableFiles(response.data.files || []);
    } catch (error) {
      console.error('Failed to load files:', error);
      setAvailableFiles([]);
    } finally {
      setIsLoadingFiles(false);
    }
  };

  // Load available datasources
  const loadDatasources = async () => {
    setIsLoadingDatasources(true);
    try {
      const response = await chatApi.getDatasources();
      setAvailableDatasources(response.data.datasources || []);
    } catch (error) {
      console.error('Failed to load datasources:', error);
      setAvailableDatasources([]);
    } finally {
      setIsLoadingDatasources(false);
    }
  };

  // Handle @ command detection (Cursor-like): show commands + files when typing '@'
  useEffect(() => {
    if (message.startsWith('@')) {
      setShowAtCommands(true);
      setSelectedAtCommandIndex(0);

      // Always load files so we can list them next to commands
      loadFiles();

      // Parse specific sub-commands if user types them (still supported)
      const atPart = message.split(' ')[0];
      if (atPart.startsWith('@load-file/')) {
        setAtCommandType('load-file');
        setAtCommandQuery(atPart.replace('@load-file/', ''));
      } else if (atPart.startsWith('@db')) {
        setAtCommandType('db');
        setAtCommandQuery(atPart.replace('@db', ''));
        loadDatasources();
      } else if (atPart.startsWith('@files')) {
        setAtCommandType('files');
        setAtCommandQuery(atPart.replace('@files', ''));
      } else {
        setAtCommandType(null);
        setAtCommandQuery(message.slice(1));
      }
    } else {
      setShowAtCommands(false);
      setAtCommandType(null);
      setAtCommandQuery('');
    }
  }, [message]);

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
      // Handle @ commands
      if (message.startsWith('@') && onAtCommand) {
        const parts = message.split(' ');
        const command = parts[0];
        const args = parts.slice(1);
        onAtCommand(command, args);
        setMessage('');
        setShowAtCommands(false);
        return;
      }
      
      onSend(message.trim());
      setMessage('');
      setShowCommands(false);
      setShowAtCommands(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (showAtCommands) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedAtCommandIndex(prev => 
          prev < filteredAtCommands.length - 1 ? prev + 1 : 0
        );
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedAtCommandIndex(prev => 
          prev > 0 ? prev - 1 : filteredAtCommands.length - 1
        );
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (filteredAtCommands[selectedAtCommandIndex]) {
          const selectedCmd = filteredAtCommands[selectedAtCommandIndex];
          if (selectedCmd.examples.length > 0) {
            setMessage(selectedCmd.examples[0]);
          } else {
            setMessage(selectedCmd.command + ' ');
          }
          setShowAtCommands(false);
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        setShowAtCommands(false);
      } else if (e.key === 'Tab') {
        e.preventDefault();
        if (filteredAtCommands[selectedAtCommandIndex]) {
          const selectedCmd = filteredAtCommands[selectedAtCommandIndex];
          setMessage(selectedCmd.command + ' ');
          setShowAtCommands(false);
        }
      }
    } else if (showCommands) {
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
      } else if (e.key === 'Tab') {
        e.preventDefault();
        if (filteredCommands[selectedCommandIndex]) {
          const selectedCmd = filteredCommands[selectedCommandIndex];
          setMessage(selectedCmd.command + ' ');
          setShowCommands(false);
        }
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

  const handleAtCommandSelect = (command: AtCommand) => {
    if (command.examples.length > 0) {
      setMessage(command.examples[0]);
    } else {
      setMessage(command.command + ' ');
    }
    setShowAtCommands(false);
    textareaRef.current?.focus();
  };

  const handleFileSelect = (file: FileItem) => {
    // Immediate load on selection (Cursor-like)
    if (onAtCommand) {
      onAtCommand(`@load-file/${file.filename}`, []);
      setMessage('');
      setShowAtCommands(false);
      return;
    }
    setMessage(`@load-file/${file.filename} `);
    setShowAtCommands(false);
    textareaRef.current?.focus();
  };

  const handleDatasourceSelect = (datasource: DatasourceItem) => {
    setMessage(`@db/${datasource.id} `);
    setShowAtCommands(false);
    textareaRef.current?.focus();
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
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

      {/* @ Command Autocomplete Dropdown (Commands + Files) */}
      {showAtCommands && (
        <div className="absolute bottom-full left-0 right-0 mb-2 bg-white border border-gray-200 rounded-lg shadow-lg z-50 max-h-72 overflow-y-auto">
          <div className="p-3 border-b border-gray-100">
            <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
              <HelpCircle className="h-4 w-4" />
              Commands
            </div>
          </div>
          {filteredAtCommands.map((command, index) => {
            const IconComponent = command.icon;
            return (
              <div
                key={command.command}
                className={`flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 ${
                  index === selectedAtCommandIndex 
                    ? 'bg-blue-50 text-blue-700' 
                    : 'hover:bg-gray-50'
                }`}
                onClick={() => handleAtCommandSelect(command)}
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
          <div className="p-3 border-b border-gray-100">
            <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
              <FileText className="h-4 w-4" />
              Files
              {isLoadingFiles && <div className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>}
            </div>
          </div>
          {/* Recent files from localStorage */}
          {(() => {
            try {
              const recent: string[] = JSON.parse(localStorage.getItem('air_recent_files') || '[]');
              if (recent.length > 0 && (!atCommandType || message.startsWith('@recent'))) {
                return (
                  <>
                    <div className="px-4 py-2 text-xs text-gray-500">Recent</div>
                    {recent.map((name) => (
                      <div
                        key={`recent-${name}`}
                        className="flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 hover:bg-gray-50"
                        onClick={() => handleFileSelect({ file_id: name, filename: name, file_type: '', file_size: 0 })}
                      >
                        <FileText className="h-4 w-4 flex-shrink-0 text-amber-600" />
                        <div className="flex-1 min-w-0">
                          <div className="font-medium text-sm">{name}</div>
                          <div className="text-xs text-gray-500">recent</div>
                        </div>
                      </div>
                    ))}
                  </>
                );
              }
            } catch {}
            return null;
          })()}
          {availableFiles
            .filter(file => file.filename.toLowerCase().includes(atCommandQuery.toLowerCase()))
            .map((file) => (
            <div
              key={file.file_id}
              className="flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 hover:bg-gray-50"
              onClick={() => handleFileSelect(file)}
            >
              <FileText className="h-4 w-4 flex-shrink-0 text-blue-600" />
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm">{file.filename}</div>
                <div className="text-xs text-gray-500">{file.file_type} • {Math.round(file.file_size / 1024)}KB</div>
              </div>
            </div>
          ))}
          {/* Datasources section remains for completeness */}
          {atCommandType === 'db' && availableDatasources.map((datasource) => (
            <div
              key={datasource.id}
              className="flex items-center gap-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 hover:bg-gray-50"
              onClick={() => handleDatasourceSelect(datasource)}
            >
              <Database className="h-4 w-4 flex-shrink-0 text-green-600" />
              <div className="flex-1 min-w-0">
                <div className="font-medium text-sm">{datasource.name}</div>
                <div className="text-xs text-gray-500">{datasource.type} • {datasource.connected ? 'Connected' : 'Disconnected'}</div>
              </div>
            </div>
          ))}
          {/* Empty states */}
          {!isLoadingFiles && availableFiles.length === 0 && (
            <div className="p-4 text-center text-gray-500 text-sm">
              No files available. Upload a file first.
            </div>
          )}

          {!isLoadingFiles && availableFiles.length > 0 && availableFiles.filter(file => file.filename.toLowerCase().includes(atCommandQuery.toLowerCase())).length === 0 && (
            <div className="p-4 text-center text-gray-500 text-sm">
              No files match "{atCommandQuery}". Try a different search.
            </div>
          )}

          {atCommandType === 'db' && !isLoadingDatasources && availableDatasources.length === 0 && (
            <div className="p-4 text-center text-gray-500 text-sm">
              No databases available. Configure a datasource first.
            </div>
          )}

          {atCommandType === 'load-file' && !isLoadingFiles && availableFiles.filter(file => file.filename.toLowerCase().includes(atCommandQuery.toLowerCase())).length === 0 && (
            <div className="p-4 text-center text-gray-500 text-sm">
              No files match "{atCommandQuery}". Try a different search.
            </div>
          )}
        </div>
      )}

      {/* Slash Command Autocomplete Dropdown */}
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
          onChange={handleFileUpload}
          className="hidden"
          accept=".csv,.json,.xlsx,.xls,.parquet"
        />
      </form>
    </div>
  );
}
