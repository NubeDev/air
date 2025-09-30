import { useState, useRef, useEffect } from 'react';
import { Message } from './Message';
import { ChatInput } from './ChatInput';
import { ModelSelector, type AIModel } from './ModelSelector';
import { Button } from '@/components/ui/button';
import { Database, Bug, Eye, EyeOff, X } from 'lucide-react';
import { chatApi } from '@/services/chatApi';
import { wsService } from '@/services/websocket';
import type { ChatMessage } from '@/types/api';
import nubeLogo from '@/assets/nube-logo.png';

interface ChatWindowProps {
  reportId?: string;
}

export function ChatWindow({ reportId }: ChatWindowProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedModel, setSelectedModel] = useState<AIModel>('llama');
  const [modelStatus, setModelStatus] = useState<Record<AIModel, { connected: boolean; error?: string }>>({
    llama: { connected: true },
    openai: { connected: false, error: 'No API key configured' },
    sqlcoder: { connected: true }
  });
  const [uploadedFiles, setUploadedFiles] = useState<Array<{ file_id: string; filename: string; file_size: number; upload_time: string; file_type: string }>>([]);
  const [selectedFile, setSelectedFile] = useState<string>('');
  const [activeTab, setActiveTab] = useState<'upload' | 'existing'>('upload');
  const [showScrollButton, setShowScrollButton] = useState(false);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [showDebug, setShowDebug] = useState(false);
  const [debugMessages, setDebugMessages] = useState<any[]>([]);
  const [attachedFiles, setAttachedFiles] = useState<File[]>([]);

  useEffect(() => {
    // Load model status and uploaded files
    loadModelStatus();
    loadUploadedFiles();
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
      setShowScrollButton(scrollTop < scrollHeight - clientHeight - 100);
    }
  };

  // Scroll to bottom function
  const scrollToBottom = () => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  };

  // Debug message capture
  const addDebugMessage = (type: string, data: any) => {
    const debugMessage = {
      id: `debug_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      timestamp: new Date().toISOString(),
      type,
      data
    };
    setDebugMessages(prev => [...prev, debugMessage]);
  };

  // File attachment handlers
  const handleFileAttach = (file: File) => {
    setAttachedFiles(prev => [...prev, file]);
  };

  const handleRemoveFile = (index: number) => {
    setAttachedFiles(prev => prev.filter((_, i) => i !== index));
  };

  const loadModelStatus = async () => {
    try {
      const response = await chatApi.getModelStatus();
      setModelStatus(response.data);
    } catch (error) {
      console.error('Failed to load model status:', error);
    }
  };


  const loadUploadedFiles = async () => {
    try {
      const response = await chatApi.getUploadedFiles();
      setUploadedFiles(response.data.files);
      if (response.data.files.length > 0) {
        setSelectedFile(response.data.files[0].file_id);
      }
    } catch (error) {
      console.error('Failed to load uploaded files:', error);
    }
  };

  const handleDataQuery = async (query: string) => {
    // For now, just send as a regular chat message since we're focusing on file analysis
    
    const messageData = {
      type: 'chat_message',
      payload: {
        content: `Data query: ${query}`,
        model: selectedModel,
      },
    };
    
    wsService.sendMessage(messageData);
  };

  const handleCreateReport = async (description: string) => {
    // For now, just send as a regular chat message since we're focusing on file analysis
    
    const messageData = {
      type: 'chat_message',
      payload: {
        content: `Create report: ${description}`,
        model: selectedModel,
      },
    };
    
    wsService.sendMessage(messageData);
  };

  const handleFileAnalysis = async (query: string) => {
    if (!selectedFile) return;
    
    // Send file analysis request via WebSocket
    wsService.analyzeFile(selectedFile, query, selectedModel);
  };

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachedFiles.length === 0) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date().toISOString(),
      report_id: reportId,
    };

    setMessages(prev => [...prev, userMessage]);
    setIsLoading(true);

    try {
      // Handle slash commands
      if (content.startsWith('/')) {
        await handleSlashCommand(content);
      }
      // If files are attached, analyze them
      else if (attachedFiles.length > 0) {
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
          type: 'chat_message',
          payload: {
            content: content,
            model: selectedModel,
          },
        };
        
        // Capture debug info
        addDebugMessage('user_message', {
          content,
          model: selectedModel,
          timestamp: new Date().toISOString()
        });
        
        wsService.sendMessage(messageData);
      }
    } catch (error) {
      console.error('Failed to send message:', error);
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `Sorry, I encountered an error: ${error instanceof Error ? error.message : 'Unknown error'}`,
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
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

  useEffect(() => {
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  }, [messages]);

  // WebSocket message handlers
  useEffect(() => {
    // Connect to WebSocket with a small delay to ensure backend is ready
    const connectWebSocket = async () => {
      try {
        await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second
        await wsService.connect();
        console.log('WebSocket connected in ChatWindow');
      } catch (error) {
        console.error('Failed to connect WebSocket in ChatWindow:', error);
      }
    };
    
    connectWebSocket();

    // Handle file analysis responses
    wsService.onMessage('file_analysis_started', (message) => {
      // Capture debug info
      addDebugMessage('file_analysis_started', {
        file_id: message.payload.file_id,
        model: message.payload.model,
        timestamp: new Date().toISOString(),
        fullMessage: message
      });
      
      // File analysis started - no hardcoded message needed
    });

    wsService.onMessage('file_analysis_complete', (message) => {
      // Capture debug info
      addDebugMessage('file_analysis_complete', {
        file_id: message.payload.file_id,
        model: message.payload.model,
        analysis: message.payload.analysis,
        insights: message.payload.insights,
        suggestions: message.payload.suggestions,
        timestamp: new Date().toISOString(),
        fullMessage: message
      });
      
      const { analysis, insights, suggestions, data_info } = message.payload;
      
      const resultMessage: ChatMessage = {
        id: `msg_${Date.now()}`,
        role: 'assistant',
        content: analysis,
        timestamp: new Date().toISOString(),
        model: selectedModel,
        metadata: {
          type: 'analysis',
          data: {
            insights,
            suggestions,
            data_info,
          },
        },
      };
      
      setMessages(prev => [...prev, resultMessage]);
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
    });

    // Handle load dataset success
    wsService.onMessage('load_dataset_success', (message) => {
      setSelectedFile(message.payload.filename);
      const successMessage: ChatMessage = {
        id: `load_success_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: message.payload.message,
        timestamp: new Date().toISOString(),
        model: selectedModel,
      };
      setMessages(prev => [...prev, successMessage]);
    });

    // Handle load dataset error
    wsService.onMessage('load_dataset_error', (message) => {
      const errorMessage: ChatMessage = {
        id: `load_error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        role: 'assistant',
        content: `❌ Load failed: ${message.payload.error}`,
        timestamp: new Date().toISOString(),
        model: selectedModel,
      };
      setMessages(prev => [...prev, errorMessage]);
    });

        // File analysis progress - no hardcoded messages needed

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
        });

        // Handle typing indicators
        wsService.onMessage('chat_typing', (message) => {
          if (message.payload.is_typing) {
            setIsLoading(true);
          } else {
            setIsLoading(false);
          }
        });

    // Cleanup on unmount
    return () => {
      wsService.disconnect();
    };
  }, [selectedModel]);

  // No hardcoded quick actions - let the AI suggest what to do

  return (
    <div className="h-screen flex flex-col bg-white">
      {/* Header */}
      <div className="border-b bg-white px-6 py-4 flex-shrink-0">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img 
              src={nubeLogo} 
              alt="Nube Logo" 
              className="h-20 w-20 object-contain rounded-lg"
            />
            <div>
              <h1 className="text-xl font-semibold text-gray-900">AIR Assistant</h1>
            </div>
          </div>
          
          <div className="flex items-center gap-3">
            {/* Debug Button */}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDebug(!showDebug)}
              className="flex items-center gap-2"
            >
              {showDebug ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              <Bug className="h-4 w-4" />
              Debug
            </Button>
            
            {/* Model Selector */}
            <ModelSelector 
              selectedModel={selectedModel}
              onModelChange={setSelectedModel}
              modelStatus={modelStatus}
            />
          </div>
        </div>
      </div>

      {/* Data Source Tabs */}
      <div className="border-b bg-gray-50 px-6 py-4 flex-shrink-0">
        <div className="flex space-x-1 bg-white p-1 rounded-lg border">
          <button
            onClick={() => setActiveTab('upload')}
            className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === 'upload'
                ? 'bg-blue-500 text-white shadow-sm'
                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
            }`}
          >
            📁 Upload New File
          </button>
          <button
            onClick={() => setActiveTab('existing')}
            className={`px-4 py-2 text-sm font-medium rounded-md transition-colors ${
              activeTab === 'existing'
                ? 'bg-blue-500 text-white shadow-sm'
                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
            }`}
          >
            📊 Existing Datasets
          </button>
        </div>
        
        {/* Tab Content */}
        {activeTab === 'upload' && (
          <div className="mt-4">
            <div className="text-sm text-gray-600 mb-2">
              Upload a new file to analyze with AI
            </div>
            <div className="flex items-center gap-2">
              <input
                type="file"
                accept=".csv,.json,.xlsx,.xls,.parquet"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    handleFileAttach(file);
                  }
                }}
                className="text-sm text-gray-600 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"
              />
            </div>
          </div>
        )}
        
        {activeTab === 'existing' && (
          <div className="mt-4">
            <div className="text-sm text-gray-600 mb-2">
              Select from previously uploaded datasets
            </div>
            <select
              value={selectedFile}
              onChange={(e) => setSelectedFile(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">Choose a dataset...</option>
              {uploadedFiles.map((file) => (
                <option key={file.file_id} value={file.file_id}>
                  {file.filename} ({file.file_type}) - {new Date(file.upload_time).toLocaleDateString()}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
      
      {/* Messages - This is the scrollable area */}
      <div className="flex-1 min-h-0 bg-gray-50 relative">
        <div className="h-full overflow-y-auto" ref={scrollAreaRef} onScroll={handleScroll}>
          <div className="max-w-4xl mx-auto">
            {messages.length === 0 && (
              <div className="text-center py-12">
                <div className="w-16 h-16 mx-auto mb-4 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center">
                  <Database className="h-8 w-8 text-white" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">Welcome to AIR</h3>
                <p className="text-gray-600 mb-6 max-w-md mx-auto">
                  Start a conversation to begin analyzing your data.
                </p>
              </div>
            )}
            
            {messages.map((message) => (
              <Message key={message.id} message={message} />
            ))}
            
            {isLoading && (
              <div className="flex items-center justify-start gap-4 py-6">
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white text-sm font-semibold">
                  AI
                </div>
                <div className="bg-gray-100 rounded-2xl px-4 py-3">
                  <div className="flex items-center space-x-2">
                    <div className="flex space-x-1">
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                    </div>
                    <span className="text-sm text-gray-600">{selectedModel} is thinking...</span>
                  </div>
                </div>
              </div>
            )}
            
            {/* Add padding at bottom for better scrolling */}
            <div className="h-20"></div>
          </div>
        </div>
        
        {/* Scroll to bottom button */}
        {showScrollButton && (
          <button
            onClick={scrollToBottom}
            className="absolute bottom-20 right-4 bg-blue-500 hover:bg-blue-600 text-white rounded-full p-2 shadow-lg transition-all duration-200 z-10"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
            </svg>
          </button>
        )}
      </div>
      
      {/* Debug Panel - Fixed position overlay */}
      {showDebug && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between p-4 border-b">
              <h3 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <Bug className="h-5 w-5" />
                Debug Messages
              </h3>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setDebugMessages([])}
                >
                  Clear
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowDebug(false)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto p-4">
              {debugMessages.length === 0 ? (
                <p className="text-sm text-gray-500 italic text-center py-8">No debug messages yet. Send a message to see AI communication details.</p>
              ) : (
                <div className="space-y-3">
                  {debugMessages.map((debugMsg) => (
                    <div key={debugMsg.id} className="bg-gray-50 rounded border p-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="font-mono text-blue-600 font-semibold">{debugMsg.type}</span>
                        <span className="text-gray-400 text-sm">{new Date(debugMsg.timestamp).toLocaleTimeString()}</span>
                      </div>
                      <pre className="whitespace-pre-wrap text-gray-700 bg-white p-3 rounded text-sm overflow-x-auto border">
                        {JSON.stringify(debugMsg.data, null, 2)}
                      </pre>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
      
      {/* Input - Fixed at bottom */}
      <div className="border-t bg-white px-6 py-4 flex-shrink-0 sticky bottom-0 z-20 shadow-lg">
        <ChatInput 
          onSend={handleSendMessage}
          onFileAttach={handleFileAttach}
          onRemoveFile={handleRemoveFile}
          attachedFiles={attachedFiles}
          disabled={isLoading || !modelStatus[selectedModel].connected}
          placeholder={
            !modelStatus[selectedModel].connected 
              ? `Cannot send message - ${modelStatus[selectedModel].error || 'No connection'}`
              : selectedFile 
                ? `Ask me about ${uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'your dataset'}...`
                : "Ask me about your data or request a report... (try / for commands)"
          }
        />
      </div>
    </div>
  );
}
