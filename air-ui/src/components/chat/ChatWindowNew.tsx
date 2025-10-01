import { useState, useRef, useEffect } from 'react';
import { ChatInput } from './ChatInput';
import { type AIModel, type ModelInfo } from './ModelSelector';
import { Button } from '@/components/ui/button';
import { X, Upload, ChevronDown, Copy } from 'lucide-react';
import { EphemeralSystemCard } from './EphemeralSystemCard';
import { FloatingDot } from './FloatingDot';
import { ChatHeader } from './ChatHeader';
import { MessageList } from './MessageList';
import { AnalyzeQuickAction } from './AnalyzeQuickAction';
import { chatApi } from '@/services/chatApi';
import { wsService } from '@/services/websocket';
import type { ChatMessage } from '@/types/api';
// removed inline logo for cleaner header

interface ChatWindowProps {
  reportId?: string;
}

export function ChatWindow({ reportId }: ChatWindowProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState<AIModel>('');
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [modelHealth, setModelHealth] = useState<Record<string, { connected: boolean; error?: string }>>({});
  const [uploadedFiles, setUploadedFiles] = useState<Array<{ file_id: string; filename: string; file_size: number; upload_time: string; file_type: string }>>([]);
  const [selectedFile, setSelectedFile] = useState<string>('');
  const [showScrollButton, setShowScrollButton] = useState(false);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [showDebug, setShowDebug] = useState(false);
  const [debugMessages, setDebugMessages] = useState<any[]>([]);
  const [attachedFiles, setAttachedFiles] = useState<File[]>([]);
  const [wsConnected, setWsConnected] = useState(false);
  const [backendPending, setBackendPending] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [typingMessage, setTypingMessage] = useState('');
  const [rawAIMode, setRawAIMode] = useState(false);
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [ephemeralCard, setEphemeralCard] = useState<{
    type: 'file_needed' | 'file_loaded' | 'uploading';
    files?: Array<{ file_id: string; filename: string; file_size: number; upload_time: string; file_type: string }>;
    selectedFileName?: string;
  } | null>(null);

  useEffect(() => {
    // Load models and defaults
    loadModels();
    loadUploadedFiles();
    
    // Initialize WebSocket connection
    const initWebSocket = async () => {
      try {
        await wsService.connect();
        console.log('WebSocket connected in ChatWindow');
        setWsConnected(true);
      } catch (error) {
        console.error('Failed to connect WebSocket:', error);
        setWsConnected(false);
      }
    };
    
    initWebSocket();
    
    // Listen for WebSocket connection changes
    const handleWsConnect = () => setWsConnected(true);
    const handleWsDisconnect = () => setWsConnected(false);
    
    // Add event listeners for WebSocket status
    window.addEventListener('ws-connected', handleWsConnect);
    window.addEventListener('ws-disconnected', handleWsDisconnect);
    
    return () => {
      window.removeEventListener('ws-connected', handleWsConnect);
      window.removeEventListener('ws-disconnected', handleWsDisconnect);
    };
  }, []);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  }, [messages]);

  // Handle scroll events to show/hide scroll button
  const handleScroll = () => {
    if (scrollAreaRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollAreaRef.current;
      setShowScrollButton(scrollHeight - scrollTop > clientHeight + 100);
    }
  };

  const scrollToBottom = () => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  };

  const loadModels = async () => {
    try {
      const [listRes, statusRes] = await Promise.all([
        chatApi.getModels(),
        chatApi.getModelStatus(),
      ]);
      const defaults = listRes.data.defaults || {};
      const chatDefault = defaults.chat as string | undefined;
      setModels(listRes.data.models || []);
      setModelHealth(statusRes.data.health || {});
      if (chatDefault) setSelectedModel(chatDefault);
      else if ((listRes.data.models || []).length > 0) setSelectedModel(listRes.data.models[0].id);
    } catch (error) {
      console.error('Failed to load models:', error);
    }
  };

  const loadUploadedFiles = async () => {
    try {
      const response = await chatApi.getUploadedFiles();
      setUploadedFiles(response.data.files);
    } catch (error) {
      console.error('Failed to load uploaded files:', error);
    }
  };

  const addDebugMessage = (type: string, data: any) => {
    const debugMsg = {
      id: Date.now().toString(),
      type,
      data,
      timestamp: new Date().toISOString()
    };
    setDebugMessages(prev => [...prev, debugMsg]);
  };

  const copyDebugMessages = async () => {
    try {
      const debugText = JSON.stringify(debugMessages, null, 2);
      await navigator.clipboard.writeText(debugText);
      // You could add a toast notification here if you have one
      console.log('Debug messages copied to clipboard');
    } catch (err) {
      console.error('Failed to copy debug messages:', err);
    }
  };

  const handleSendMessage = async (content: string) => {
    if (!content.trim()) return;

    const userMessage: ChatMessage = {
      id: `user_${Date.now()}`,
      role: 'user',
      content: content,
      timestamp: new Date().toISOString(),
      report_id: reportId,
    };
    
    setMessages(prev => [...prev, userMessage]);

    try {
      // Handle @ commands first
      if (content.startsWith('@')) {
        await handleAtCommand(content);
        return; // Don't send to AI, handle locally
      }
      
      // Handle slash commands locally first
      if (content.startsWith('/')) {
        await handleSlashCommand(content);
        return; // Don't send to AI, handle locally
      }
      
      // Set loading states for AI interactions
      setIsLoading(true);
      setIsTyping(true);
      setTypingMessage('AI is thinking...');
      setBackendPending(true);
      
      // If files are attached, analyze them
      if (attachedFiles.length > 0) {
        for (const file of attachedFiles) {
          await handleFileAnalysis(`Analyze this file: ${file.name}`);
        }
        setAttachedFiles([]); // Clear attached files after processing
      } else if (selectedFile && (content.toLowerCase().includes('analyze') || content.toLowerCase().includes('file') || content.toLowerCase().includes('data'))) {
        await handleFileAnalysis(content);
      } else if (content.toLowerCase().includes('show me') || content.toLowerCase().includes('query')) {
        await handleDataQuery(content);
      } else if (content.toLowerCase().includes('create report') || content.toLowerCase().includes('generate report')) {
        await handleCreateReport(content);
      } else {
        // General chat message via WebSocket
        const messageData = {
          type: rawAIMode ? 'raw_ai_message' : 'chat_message',
          payload: {
            content: content,
            model: selectedModel,
          },
        };
        
        // Capture debug info
        addDebugMessage('user_message', {
          content,
          model: selectedModel,
          rawMode: rawAIMode,
          timestamp: new Date().toISOString()
        });
        
        wsService.sendMessage(messageData);
      }
    } catch (error) {
      console.error('Error sending message:', error);
      const errorMessage: ChatMessage = {
        id: `error_${Date.now()}`,
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
      setBackendPending(false);
    }
  };

  const handleAtCommand = async (command: string) => {
    const [cmd] = command.split(' ');
    
    // Handle empty @ command
    if (cmd === '@' || cmd === '') {
      const helpMessage: ChatMessage = {
        id: `at_help_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: `**Available @ Commands:**

**@files** - List available files
- Example: \`@files\`

**@db** - List available databases
- Example: \`@db\`

**@load-file/<filename>** - Load a specific file
- Example: \`@load-file/abc.csv\`

**Current Status:**
- Loaded dataset: ${selectedFile ? uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'None' : 'None'}
- Available files: ${uploadedFiles.length}`,
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, helpMessage]);
      return;
    }
    
    switch (cmd) {
      case '@files':
        // This is handled by the UI autocomplete, no need to send a message
        break;
        
      case '@db':
        // This is handled by the UI autocomplete, no need to send a message
        break;
        
      default:
        if (cmd.startsWith('@load-file/')) {
          const filename = cmd.replace('@load-file/', '');
          
          // Try to find the file in the current uploaded files
          let file = uploadedFiles.find(f => f.filename === filename);
          
          // If not found, try to load files from backend
          if (!file) {
            try {
              const response = await chatApi.getUploadedFiles();
              const files = response.data.files || [];
              file = files.find((f: any) => f.filename === filename);
            } catch (error) {
              console.error('Failed to load files from backend:', error);
            }
          }
          
          if (file) {
            // Immediately load the dataset via WebSocket and set locally
            wsService.sendMessage({
              type: 'load_dataset',
              payload: {
                filename: file.file_id
              }
            });
            setSelectedFile(file.file_id);
            
            // Brief confirmation inline (non-ephemeral)
            const confirmMessage: ChatMessage = {
              id: `at_load_ok_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              role: 'assistant',
              content: `✅ Loaded dataset: ${filename}`,
              timestamp: new Date().toISOString(),
              report_id: reportId,
            };
            setMessages(prev => [...prev, confirmMessage]);
          } else {
            const errorMessage: ChatMessage = {
              id: `at_load_error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              role: 'assistant',
              content: `❌ **File not found: ${filename}**

Use \`@files\` to see available files.`,
              timestamp: new Date().toISOString(),
              report_id: reportId,
            };
            setMessages(prev => [...prev, errorMessage]);
          }
        } else {
          const unknownMessage: ChatMessage = {
            id: `at_unknown_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            role: 'assistant',
            content: `❌ **Unknown @ command: ${cmd}**

Use \`@files\`, \`@db\`, or \`@load-file/<filename>\``,
            timestamp: new Date().toISOString(),
            report_id: reportId,
          };
          setMessages(prev => [...prev, unknownMessage]);
        }
    }
  };

  const handleSlashCommand = async (command: string) => {
    const [cmd, ...args] = command.split(' ');
    
    // Handle empty slash command
    if (cmd === '/' || cmd === '') {
      const helpMessage: ChatMessage = {
        id: `help_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: `**Available Commands:**

**/load <filename>** - Load an existing dataset
- Example: \`/load your_file.csv\`

**/analyze [query]** - Analyze the currently loaded dataset
- Example: \`/analyze show me trends\`

**/help** - Show this help message

**Current Status:**
- Loaded dataset: ${selectedFile ? uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'None' : 'None'}
- Available datasets: ${uploadedFiles.length}`,
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, helpMessage]);
      return;
    }
    
    switch (cmd) {
      case '/load':
        if (args.length === 0) {
          if (uploadedFiles.length === 0) {
            const noFilesMessage: ChatMessage = {
              id: `load_no_files_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              role: 'assistant',
              content: `📁 No datasets available. Please upload a file first using the paperclip icon above.`,
              timestamp: new Date().toISOString(),
              report_id: reportId,
            };
            setMessages(prev => [...prev, noFilesMessage]);
          } else {
            const helpMessage: ChatMessage = {
              id: `load_help_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              role: 'assistant',
              content: `📁 **Available Datasets (${uploadedFiles.length}):**

${uploadedFiles.map((f, i) => `${i + 1}. **${f.filename}** (${f.file_type})`).join('\n')}

**Usage:** \`/load <filename>\` or \`/load <number>\`

**Examples:**
- \`/load ${uploadedFiles[0]?.filename}\`
- \`/load 1\` (to load the first file)`,
              timestamp: new Date().toISOString(),
              report_id: reportId,
            };
            setMessages(prev => [...prev, helpMessage]);
          }
        } else {
          const filename = args.join(' ');
          let file = null;
          
          // Check if it's a number (index)
          if (/^\d+$/.test(filename)) {
            const index = parseInt(filename) - 1;
            if (index >= 0 && index < uploadedFiles.length) {
              file = uploadedFiles[index];
            }
          } else {
            // Search by filename
            file = uploadedFiles.find(f => f.filename.toLowerCase().includes(filename.toLowerCase()));
          }
          
          if (file) {
            // Send load_dataset message to backend
            wsService.sendMessage({
              type: 'load_dataset',
              payload: {
                filename: file.file_id
              }
            });
          } else {
            const errorMessage: ChatMessage = {
              id: `load_error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
              role: 'assistant',
              content: `❌ **Dataset not found:** \`${filename}\`

**Available datasets:**
${uploadedFiles.map((f, i) => `${i + 1}. ${f.filename}`).join('\n')}

**Usage:** \`/load <filename>\` or \`/load <number>\``,
              timestamp: new Date().toISOString(),
              report_id: reportId,
            };
            setMessages(prev => [...prev, errorMessage]);
          }
        }
        break;
        
      case '/analyze':
        if (selectedFile) {
          const query = args.join(' ') || 'Analyze this dataset and provide insights';
          // Set loading states for file analysis
          setIsLoading(true);
          setIsTyping(true);
          setTypingMessage('Analyzing dataset...');
          await handleFileAnalysis(query);
        } else {
          const errorMessage: ChatMessage = {
            id: `analyze_error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
            role: 'assistant',
            content: '❌ No dataset loaded. Use /load <filename> to load a dataset first.',
            timestamp: new Date().toISOString(),
            report_id: reportId,
          };
          setMessages(prev => [...prev, errorMessage]);
        }
        break;
        
      case '/help':
        const helpMessage: ChatMessage = {
          id: `help_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          role: 'assistant',
          content: `**Available Commands:**

**/load <filename>** - Load an existing dataset
- Example: \`/load your_file.csv\`

**/analyze [query]** - Analyze the currently loaded dataset
- Example: \`/analyze show me trends\`

**/help** - Show this help message

**Current Status:**
- Loaded dataset: ${selectedFile ? uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'None' : 'None'}
- Available datasets: ${uploadedFiles.length}`,
          timestamp: new Date().toISOString(),
          report_id: reportId,
        };
        setMessages(prev => [...prev, helpMessage]);
        break;
        
      default:
        const unknownMessage: ChatMessage = {
          id: `unknown_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
          role: 'assistant',
          content: `❌ Unknown command: ${cmd}\n\nUse /help to see available commands.`,
          timestamp: new Date().toISOString(),
          report_id: reportId,
        };
        setMessages(prev => [...prev, unknownMessage]);
    }
  };

  const handleFileAnalysis = async (query: string) => {
    if (!selectedFile) return;

    const analysisMessage: ChatMessage = {
      id: `analysis_${Date.now()}`,
      role: 'assistant',
      content: `Analyzing ${uploadedFiles.find(f => f.file_id === selectedFile)?.filename}...`,
      timestamp: new Date().toISOString(),
      report_id: reportId,
    };
    setMessages(prev => [...prev, analysisMessage]);

    try {
      const messageData = {
        type: 'file_analysis',
        payload: {
          file_id: selectedFile,
          query: query,
          model: selectedModel,
        },
      };
      
      addDebugMessage('file_analysis', {
        file_id: selectedFile,
        query,
        model: selectedModel,
        timestamp: new Date().toISOString()
      });
      
      wsService.sendMessage(messageData);
    } catch (error) {
      console.error('Error analyzing file:', error);
      const errorMessage: ChatMessage = {
        id: `error_${Date.now()}`,
        role: 'assistant',
        content: 'Sorry, I encountered an error analyzing the file. Please try again.',
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, errorMessage]);
    }
  };

  const handleDataQuery = async (query: string) => {
    // Implementation for data query
    console.log('Data query:', query);
  };

  const handleCreateReport = async (query: string) => {
    // Implementation for report creation
    console.log('Create report:', query);
  };

  const handleFileUpload = async (file: File) => {
    try {
      const formData = new FormData();
      formData.append('file', file);
      
      const response = await fetch('/api/v1/upload', {
        method: 'POST',
        body: formData,
      });
      
      if (response.ok) {
        const result = await response.json();
        setUploadedFiles(prev => [...prev, result]);
        setSelectedFile(result.file_id);
        
        const successMessage: ChatMessage = {
          id: `upload_${Date.now()}`,
          role: 'assistant',
          content: `✅ File uploaded successfully: ${result.filename}`,
          timestamp: new Date().toISOString(),
          report_id: reportId,
        };
        setMessages(prev => [...prev, successMessage]);
      }
    } catch (error) {
      console.error('Error uploading file:', error);
    }
  };


  // Handle file selection from ephemeral card
  const handleEphemeralFileSelect = (fileId: string) => {
    // Send WebSocket message to select file
    wsService.sendMessage({
      type: 'ephemeral_file_select',
      payload: {
        file_id: fileId
      }
    });
    
    // Update local state
    setSelectedFile(fileId);
    setEphemeralCard(null);
  };

  // Handle upload click from ephemeral card
  const handleEphemeralUploadClick = () => {
    // Trigger file input click
    const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement;
    if (fileInput) {
      fileInput.click();
    }
    setEphemeralCard(null);
  };

  // Handle ephemeral card dismiss
  const handleEphemeralDismiss = () => {
    addDebugMessage('user_cancelled_file_prompt', { timestamp: new Date().toISOString() });
    setEphemeralCard(null);
  };

  const handleFileAttach = (file: File) => {
    setAttachedFiles(prev => [...prev, file]);
  };

  const handleRemoveFile = (index: number) => {
    setAttachedFiles(prev => prev.filter((_, i) => i !== index));
  };

  // WebSocket message handlers
  useEffect(() => {
    // File analysis handlers
    wsService.onMessage('file_analysis_started', (message) => {
      addDebugMessage('file_analysis_started', message);
    });

    wsService.onMessage('file_analysis_complete', (message) => {
      addDebugMessage('file_analysis_complete', message);
      
      const resultMessage: ChatMessage = {
        id: `analysis_${Date.now()}`,
        role: 'assistant',
        content: message.payload.analysis,
        timestamp: new Date().toISOString(),
        model: selectedModel,
        metadata: {
          type: 'analysis',
          data: {
            insights: message.payload.insights,
            suggestions: message.payload.suggestions,
            data_info: message.payload.data_info,
          },
        },
      };
      
      setMessages(prev => [...prev, resultMessage]);
      
      // Clear loading states
      setIsLoading(false);
      setIsTyping(false);
      setTypingMessage('');
      setBackendPending(false);
    });

    wsService.onMessage('file_analysis_error', (message) => {
      const errorMessage: ChatMessage = {
        id: `msg_${Date.now()}`,
        role: 'assistant',
        content: `Analysis failed: ${message.payload.error}`,
        timestamp: new Date().toISOString(),
        model: selectedModel,
        metadata: {
          type: 'analysis',
          error: message.payload.error,
        },
      };
      setMessages(prev => [...prev, errorMessage]);
      
      // Clear loading states
      setIsLoading(false);
      setIsTyping(false);
      setTypingMessage('');
      setBackendPending(false);
    });

    // Handle load dataset success
    wsService.onMessage('load_dataset_success', (message) => {
      setSelectedFile(message.payload.filename);
      // toast success and update recent list
      setToast({ type: 'success', message: `Loaded dataset: ${message.payload.filename}` });
      try {
        const recent = JSON.parse(localStorage.getItem('air_recent_files') || '[]');
        const next = [message.payload.filename, ...recent.filter((f: string) => f !== message.payload.filename)].slice(0, 5);
        localStorage.setItem('air_recent_files', JSON.stringify(next));
      } catch {}
    });

    // Handle load dataset error
    wsService.onMessage('load_dataset_error', (message) => {
      // rollback optimistic selection
      setSelectedFile('');
      setToast({ type: 'error', message: `Load failed: ${message.payload.error}` });
    });

    // Handle chat responses
    wsService.onMessage('chat_response', (message) => {
      // Capture debug info
      addDebugMessage('ai_response', {
        content: message.payload.content,
        model: message.payload.model || selectedModel,
        timestamp: new Date().toISOString(),
        fullMessage: message
      });
      
      const aiMessage: ChatMessage = {
        id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: message.payload.content,
        timestamp: new Date().toISOString(),
        model: message.payload.model || selectedModel,
      };
      setMessages(prev => [...prev, aiMessage]);
      setIsLoading(false);
      setBackendPending(false);
      setIsTyping(false);
      setTypingMessage('');
    });

    // Handle raw AI responses
    wsService.onMessage('raw_ai_response', (message) => {
      // Capture debug info
      addDebugMessage('raw_ai_response', {
        content: message.payload.content,
        model: message.payload.model || selectedModel,
        timestamp: new Date().toISOString(),
        fullMessage: message
      });
      
      const aiMessage: ChatMessage = {
        id: `raw_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: message.payload.content,
        timestamp: new Date().toISOString(),
        model: message.payload.model || selectedModel,
        metadata: undefined
      };
      setMessages(prev => [...prev, aiMessage]);
      setIsLoading(false);
      setBackendPending(false);
      setIsTyping(false);
      setTypingMessage('');
      setBackendPending(false);
    });

    // Handle ephemeral file needed message
    wsService.onMessage('ephemeral_file_needed', (message) => {
      setEphemeralCard({
        type: 'file_needed',
        files: message.payload.files || []
      });
      setIsLoading(false);
      setIsTyping(false);
      setTypingMessage('');
    });

    // Handle ephemeral file loaded confirmation
    wsService.onMessage('ephemeral_file_loaded', (message) => {
      setEphemeralCard({
        type: 'file_loaded',
        selectedFileName: message.payload.filename
      });
      // Auto-dismiss after 3 seconds
      setTimeout(() => {
        setEphemeralCard(null);
      }, 3000);
    });

    // Handle typing indicators
    wsService.onMessage('chat_typing', (message) => {
      if (message.payload.is_typing) {
        setIsTyping(true);
        setTypingMessage('AI is typing...');
        setBackendPending(true);
      } else {
        setIsTyping(false);
        setTypingMessage('');
        setBackendPending(false);
      }
    });

    return () => {
      // Cleanup if needed
    };
  }, [selectedModel]);

  return (
    <div className="min-h-full flex flex-col bg-gray-50">
      {/* Header - minimal controls only */}
      <ChatHeader
        selectedModel={selectedModel}
        onModelChange={setSelectedModel}
        models={models.filter(m => m.capabilities.includes('chat'))}
        health={modelHealth}
        rawAIMode={rawAIMode}
        onToggleRawMode={setRawAIMode}
        onToggleDebug={() => setShowDebug(!showDebug)}
      />

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        <FloatingDot visible={backendPending} text={wsConnected ? 'Waiting on backend…' : 'Disconnected'} status={wsConnected ? 'waiting' : 'error'} />
        {messages.length === 0 ? (
          <div className="min-h-full flex flex-col items-center justify-center px-6">
            <div className="max-w-3xl w-full">
              {/* Welcome Section */}
              <div className="text-center mb-12">
                <h1 className="text-3xl font-bold text-gray-900 mb-4">Welcome to AIR Assistant</h1>
                <p className="text-lg text-gray-600 mb-8">Your AI-powered data analysis companion. Upload a file or start chatting to begin.</p>
              </div>
              
              {/* Upload Action */}
              <div className="max-w-sm mx-auto mb-8">
                <div className="relative">
                  <input
                    type="file"
                    accept=".csv,.json,.xlsx,.txt"
                    onChange={(e) => {
                      if (e.target.files?.[0]) {
                        handleFileUpload(e.target.files[0]);
                      }
                    }}
                    className="hidden"
                    id="file-upload"
                  />
                  <label
                    htmlFor="file-upload"
                    className="h-24 flex flex-col items-center justify-center space-y-3 hover:bg-primary/5 hover:border-primary transition-all duration-200 border-2 border-dashed border-muted rounded-xl cursor-pointer bg-white"
                  >
                    <Upload className="h-8 w-8 text-primary" />
                    <div className="text-center">
                      <div className="font-semibold text-gray-900">Upload New File</div>
                      <div className="text-sm text-gray-500">CSV, JSON, or other data files</div>
                    </div>
                  </label>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="min-h-full">
            {ephemeralCard && (
              <div className="max-w-4xl mx-auto px-4 pt-4">
                <EphemeralSystemCard
                  type={ephemeralCard.type}
                  files={ephemeralCard.files}
                  selectedFileName={ephemeralCard.selectedFileName}
                  onFileSelect={handleEphemeralFileSelect}
                  onUploadClick={handleEphemeralUploadClick}
                  onDismiss={handleEphemeralDismiss}
                />
              </div>
            )}
            {/* Messages */}
            <MessageList
              messages={messages}
              isTyping={isTyping}
              typingMessage={typingMessage}
              scrollAreaRef={scrollAreaRef}
              onScroll={handleScroll}
              onCancelTyping={() => {
                setIsTyping(false);
                setTypingMessage('');
                setBackendPending(false);
                addDebugMessage('user_cancelled', { timestamp: new Date().toISOString() });
              }}
              footer={selectedFile && !isTyping ? (
                <div className="flex justify-start">
                  <AnalyzeQuickAction
                    disabled={!wsConnected}
                    onAnalyze={() => {
                      const defaultQuery = 'Analyze this dataset and provide key insights.';
                      wsService.sendMessage({
                        type: 'file_analysis',
                        payload: { file_id: selectedFile, query: defaultQuery, model: selectedModel },
                      });
                      setIsTyping(true);
                      setTypingMessage('Analyzing dataset...');
                      setBackendPending(true);
                    }}
                  />
                </div>
              ) : null}
            />
          </div>
        )}
        
        {/* Scroll to bottom button */}
        {showScrollButton && (
          <button
            onClick={scrollToBottom}
            className="fixed bottom-20 right-8 bg-blue-600 hover:bg-blue-700 text-white rounded-full p-3 shadow-lg transition-all duration-200 z-10"
          >
            <ChevronDown className="w-5 h-5" />
          </button>
        )}
      </div>

      {/* Debug Panel */}
      {showDebug && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg w-full max-w-4xl h-3/4 flex flex-col">
            <div className="flex items-center justify-between p-4 border-b">
              <h3 className="text-lg font-semibold">Debug Information</h3>
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={copyDebugMessages}
                  disabled={debugMessages.length === 0}
                >
                  <Copy className="h-4 w-4 mr-1" />
                  Copy
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setDebugMessages([])}
                >
                  Clear
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowDebug(false)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              <pre className="text-xs bg-gray-100 p-4 rounded">
                {JSON.stringify(debugMessages, null, 2)}
              </pre>
            </div>
          </div>
        </div>
      )}
      
      {/* Input - ChatGPT style */}
      <div className="flex-shrink-0 border-t border-gray-200 bg-white px-6 py-6 sticky bottom-0 z-50">
        <div className="max-w-4xl mx-auto">
          <div className="relative">
            {/* Toast */}
            {toast && (
              <div
                className={`absolute -top-16 left-1/2 -translate-x-1/2 px-4 py-2 rounded-lg text-sm shadow-lg z-10 ${
                  toast.type === 'success' ? 'bg-primary text-primary-foreground' : 'bg-destructive text-destructive-foreground'
                }`}
                onAnimationEnd={() => setToast(null)}
              >
                {toast.message}
              </div>
            )}
            <ChatInput 
              onSend={handleSendMessage}
              onFileAttach={handleFileAttach}
              onRemoveFile={handleRemoveFile}
              onAtCommand={handleAtCommand}
              attachedFiles={attachedFiles}
              disabled={isLoading || !wsConnected || !selectedModel}
              placeholder={
                !wsConnected
                  ? 'Connecting to server...'
                  : selectedFile 
                    ? `Ask me about ${uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'your dataset'}...`
                    : 'Type @files to see available files, @db for databases, or @load-file/filename to load a file...'
              }
            />
            
            {/* Loading overlay */}
            {isLoading && (
              <div className="absolute inset-0 bg-white bg-opacity-90 flex items-center justify-center rounded-2xl backdrop-blur-sm">
                <div className="flex items-center space-x-3">
                  <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin"></div>
                  <span className="text-sm font-medium text-gray-700">Sending...</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
