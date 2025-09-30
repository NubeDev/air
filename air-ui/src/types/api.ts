export interface Report {
  id: string;
  key: string;
  title: string;
  description?: string;
  owner: string;
  created_at: string;
  updated_at: string;
  archived: boolean;
}

export interface ReportSchema {
  report_id: string;
  schema: JSONSchema;
}

export interface ReportRun {
  id: string;
  report_id: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  row_count: number;
  results: any[];
  started_at: string;
  finished_at: string;
  error_text?: string;
}

export interface ReportData {
  report_id: string;
  run_id: string;
  status: string;
  row_count: number;
  data: any[] | null;
  executed_at: string;
  completed_at: string;
  sql: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  report_id?: string;
  model?: string;
  metadata?: {
    type?: 'data_query' | 'report_creation' | 'sql_generation' | 'analysis';
    data?: any;
    sql?: string;
    error?: string;
  };
}

export interface CreateReportData {
  key: string;
  title: string;
  description?: string;
}

export interface UpdateReportData {
  title?: string;
  description?: string;
  archived?: boolean;
}

export interface JSONSchema {
  type: string;
  properties: Record<string, SchemaProperty>;
  required: string[];
}

export interface SchemaProperty {
  type: string;
  title?: string;
  description?: string;
  format?: string;
  enum?: string[];
  minimum?: number;
  maximum?: number;
  default?: any;
}
