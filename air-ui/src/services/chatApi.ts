import { api } from '@/lib/api';
import type { AIModel } from '@/components/chat/ModelSelector';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  model?: AIModel;
  metadata?: {
    type?: 'data_query' | 'report_creation' | 'sql_generation' | 'analysis';
    data?: any;
    error?: string;
  };
}

export interface ChatRequest {
  message: string;
  model: AIModel;
  context?: {
    datasource_id?: string;
    report_id?: string;
    session_id?: string;
  };
}

export interface DataQueryRequest {
  query: string;
  datasource_id: string;
  model: AIModel;
}

export interface ReportCreationRequest {
  description: string;
  datasource_id: string;
  model: AIModel;
  data_context?: any;
}

export interface FileAnalysisRequest {
  file_id: string;
  query: string;
  model: AIModel;
}

export interface UploadedFile {
  file_id: string;
  filename: string;
  file_size: number;
  upload_time: string;
  file_type: string;
  file_path: string;
}

export const chatApi = {
  // Send a chat message to the AI
  sendMessage: (request: ChatRequest) => 
    api.post<ChatMessage>('/v1/chat/message', request),

  // Query data using AI
  queryData: (request: DataQueryRequest) =>
    api.post<{ result: any; sql?: string; explanation: string }>('/v1/chat/query-data', request),

  // Create a report using AI
  createReport: (request: ReportCreationRequest) =>
    api.post<{ report: any; sql: string; explanation: string }>('/v1/chat/create-report', request),

  // Get available datasources
  getDatasources: () =>
    api.get<{ datasources: Array<{ id: string; name: string; type: string; connected: boolean }> }>('/v1/datasources'),

  // Get model status
  getModelStatus: () =>
    api.get<Record<string, { connected: boolean; error?: string }>>('/v1/ai/models/status'),

  // Start a learning session for data exploration
  startLearningSession: (datasource_id: string) =>
    api.post<{ session_id: string }>('/v1/sessions', { datasource_id }),

  // Learn about a datasource
  learnDatasource: (datasource_id: string) =>
    api.post<{ schema: any; sample_data: any }>('/v1/learn', { datasource_id }),

  // Analyze uploaded file
  analyzeFile: (request: FileAnalysisRequest) =>
    api.post<{ analysis: string; insights: string[]; suggestions: string[] }>('/v1/chat/analyze-file', request),

  // Get uploaded files
  getUploadedFiles: () =>
    api.get<{ files: UploadedFile[]; count: number }>('/v1/upload/files'),

  // Get specific uploaded file
  getUploadedFile: (fileId: string) =>
    api.get<UploadedFile>(`/v1/upload/file/${fileId}`),
};
