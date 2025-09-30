import { useState, useRef, useEffect } from 'react';
import { Message } from './Message';
import { ChatInput } from './ChatInput';
import { ModelSelector, type AIModel } from './ModelSelector';
import { Button } from '@/components/ui/button';
import { Database, Bug, X, FileText, Upload, ChevronDown } from 'lucide-react';
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
  const [wsConnected, setWsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [typingMessage, setTypingMessage] = useState('');

  useEffect(() => {
    // Load model status and uploaded files
    loadModelStatus();
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
      // Handle slash commands locally first
      if (content.startsWith('/')) {
        await handleSlashCommand(content);
        return; // Don't send to AI, handle locally
      }
      
      // Set loading states for AI interactions
      setIsLoading(true);
      setIsTyping(true);
      setTypingMessage('AI is thinking...');
      
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

  const handleFileSelect = (fileId: string) => {
    setSelectedFile(fileId);
    const file = uploadedFiles.find(f => f.file_id === fileId);
    if (file) {
      const successMessage: ChatMessage = {
        id: `load_${Date.now()}`,
        role: 'assistant',
        content: `✅ Loaded dataset: ${file.filename}`,
        timestamp: new Date().toISOString(),
        report_id: reportId,
      };
      setMessages(prev => [...prev, successMessage]);
    }
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
      setIsTyping(false);
      setTypingMessage('');
    });

    // Handle typing indicators
    wsService.onMessage('chat_typing', (message) => {
      if (message.payload.is_typing) {
        setIsTyping(true);
        setTypingMessage('AI is typing...');
      } else {
        setIsTyping(false);
        setTypingMessage('');
      }
    });

    return () => {
      // Cleanup if needed
    };
  }, [selectedModel]);

  return (
    <div className="h-full flex flex-col bg-gray-50">
      {/* Header - Clean and minimal */}
      <div className="flex-shrink-0 border-b bg-white px-4 py-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <img src={nubeLogo} alt="Nube iO" className="h-8 w-8 object-contain" />
            <h1 className="text-lg font-semibold text-gray-900">AIR Assistant</h1>
          </div>
          
          <div className="flex items-center space-x-3">
            {/* WebSocket Status */}
            <div className="flex items-center space-x-2">
              <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
              <span className="text-xs text-gray-500">
                {wsConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
            
            <ModelSelector 
              selectedModel={selectedModel}
              onModelChange={setSelectedModel}
              modelStatus={modelStatus}
            />
            
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowDebug(!showDebug)}
              className="text-gray-500 hover:text-gray-700"
            >
              <Bug className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-hidden">
        {messages.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center px-4">
            <div className="max-w-2xl w-full text-center">
              <div className="mb-8">
                <img src={nubeLogo} alt="Nube iO" className="h-16 w-16 object-contain mx-auto mb-4" />
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Welcome to AIR Assistant</h2>
                <p className="text-gray-600 mb-8">Your AI-powered data analysis companion. Upload a file or start chatting to begin.</p>
              </div>
              
              {/* Quick Actions */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                <Button
                  onClick={() => setActiveTab('upload')}
                  variant="outline"
                  className="h-20 flex flex-col items-center justify-center space-y-2 hover:bg-blue-50 hover:border-blue-300"
                >
                  <Upload className="h-6 w-6" />
                  <span className="font-medium">Upload New File</span>
                  <span className="text-sm text-gray-500">CSV, JSON, or other data files</span>
                </Button>
                
                <Button
                  onClick={() => setActiveTab('existing')}
                  variant="outline"
                  className="h-20 flex flex-col items-center justify-center space-y-2 hover:bg-blue-50 hover:border-blue-300"
                >
                  <Database className="h-6 w-6" />
                  <span className="font-medium">Load Existing Dataset</span>
                  <span className="text-sm text-gray-500">{uploadedFiles.length} files available</span>
                </Button>
              </div>
              
              {/* File Upload Area */}
              {activeTab === 'upload' && (
                <div className="bg-white rounded-lg border-2 border-dashed border-gray-300 p-8">
                  <div className="text-center">
                    <Upload className="h-12 w-12 mx-auto mb-4 text-gray-400" />
                    <h3 className="text-lg font-semibold mb-2">Upload a file</h3>
                    <p className="text-gray-500 mb-4">Choose a file to upload and analyze</p>
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
                      className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 cursor-pointer"
                    >
                      Choose File
                    </label>
                  </div>
                </div>
              )}
              
              {/* Existing Files */}
              {activeTab === 'existing' && uploadedFiles.length > 0 && (
                <div className="bg-white rounded-lg border p-6">
                  <h3 className="text-lg font-semibold mb-4">Available Datasets</h3>
                  <div className="space-y-2">
                    {uploadedFiles.map((file) => (
                      <Button
                        key={file.file_id}
                        variant="ghost"
                        onClick={() => handleFileSelect(file.file_id)}
                        className="w-full justify-start h-auto p-4 hover:bg-blue-50"
                      >
                        <FileText className="h-5 w-5 mr-3 text-blue-600" />
                        <div className="text-left">
                          <div className="font-medium">{file.filename}</div>
                          <div className="text-sm text-gray-500">{file.file_type} • {Math.round(file.file_size / 1024)}KB</div>
                        </div>
                      </Button>
                    ))}
                  </div>
                </div>
              )}
              
              {activeTab === 'existing' && uploadedFiles.length === 0 && (
                <div className="bg-white rounded-lg border p-8 text-center">
                  <Database className="h-12 w-12 mx-auto mb-4 text-gray-400" />
                  <h3 className="text-lg font-semibold mb-2">No datasets available</h3>
                  <p className="text-gray-500 mb-4">Upload a file to get started with data analysis.</p>
                  <Button onClick={() => setActiveTab('upload')}>
                    <Upload className="h-4 w-4 mr-2" />
                    Upload File
                  </Button>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className="h-full overflow-y-auto" ref={scrollAreaRef} onScroll={handleScroll}>
            <div className="max-w-4xl mx-auto px-4 py-6">
              <div className="space-y-6">
                {messages.map((message) => (
                  <Message key={message.id} message={message} />
                ))}
                
                {/* Typing Indicator */}
                {isTyping && (
                  <div className="flex items-center space-x-2 text-gray-500">
                    <div className="flex space-x-1">
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                      <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                    </div>
                    <span className="text-sm">{typingMessage}</span>
                  </div>
                )}
              </div>
            </div>
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
      <div className="flex-shrink-0 border-t bg-white px-4 py-4 sticky bottom-0 z-50">
        <div className="max-w-4xl mx-auto">
          <div className="relative">
            <ChatInput 
              onSend={handleSendMessage}
              onFileAttach={handleFileAttach}
              onRemoveFile={handleRemoveFile}
              attachedFiles={attachedFiles}
              disabled={isLoading || !modelStatus[selectedModel].connected || !wsConnected}
              placeholder={
                !wsConnected
                  ? 'Connecting to server...'
                  : !modelStatus[selectedModel].connected 
                    ? `Cannot send message - ${modelStatus[selectedModel].error || 'No connection'}`
                    : selectedFile 
                      ? `Ask me about ${uploadedFiles.find(f => f.file_id === selectedFile)?.filename || 'your dataset'}...`
                      : 'Message AIR Assistant...'
              }
            />
            
            {/* Loading overlay */}
            {isLoading && (
              <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center rounded-2xl">
                <div className="flex items-center space-x-2">
                  <div className="w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
                  <span className="text-sm text-gray-600">Sending...</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
