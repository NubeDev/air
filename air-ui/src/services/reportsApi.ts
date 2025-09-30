import { api } from '@/lib/api';
import type { Report, ReportSchema, ReportRun, ReportData, CreateReportData, UpdateReportData } from '@/types/api';

export const reportsApi = {
  list: () => api.get<Report[]>('/v1/reports'),
  get: (id: string) => api.get<Report>(`/v1/reports/${id}`),
  create: (data: CreateReportData) => api.post<Report>('/v1/reports', data),
  update: (id: string, data: UpdateReportData) => api.put<Report>(`/v1/reports/${id}`, data),
  delete: (id: string) => api.delete(`/v1/reports/${id}`),
  execute: (id: string, params: Record<string, any>) => 
    api.post<ReportRun>(`/v1/reports/${id}/execute`, { params, datasource_id: 'sqlite-dev' }),
  getSchema: (id: string) => api.get<ReportSchema>(`/v1/reports/${id}/schema`),
  getData: (id: string) => api.get<ReportData>(`/v1/reports/${id}/data`),
};
