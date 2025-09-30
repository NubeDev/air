import { useState, useEffect } from 'react';
import { MainLayout } from '@/components/layout/MainLayout';
import { ChatWindow } from '@/components/chat/ChatWindowNew';
import { ReportCard } from '@/components/reports/ReportCard';
import { SchemaForm } from '@/components/forms/SchemaForm';
import { FileUpload } from '@/components/upload/FileUpload';
import { FileList } from '@/components/upload/FileList';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { reportsApi } from '@/services/reportsApi';
import type { Report, ReportSchema } from '@/types/api';

export function Dashboard() {
  const [reports, setReports] = useState<Report[]>([]);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [reportSchema, setReportSchema] = useState<ReportSchema | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('chat');

  useEffect(() => {
    fetchReports();
  }, []);

  const fetchReports = async () => {
    try {
      const response = await reportsApi.list();
      console.log('Reports API response:', response);
      
      // Handle different possible response structures
      let reportsData: Report[] = [];
      if (response.data) {
        if (Array.isArray(response.data)) {
          reportsData = response.data;
        } else if (response.data && typeof response.data === 'object' && 'reports' in response.data) {
          const data = response.data as any;
          if (Array.isArray(data.reports)) {
            reportsData = data.reports;
          }
        }
      }
      
      setReports(reportsData);
    } catch (error) {
      console.error('Failed to fetch reports:', error);
      setReports([]); // Ensure reports is always an array
    }
  };

  const handleReportSelect = async (report: Report) => {
    setSelectedReport(report);
    setActiveTab('execute');
    
    try {
      const response = await reportsApi.getSchema(report.id);
      setReportSchema(response.data);
    } catch (error) {
      console.error('Failed to fetch report schema:', error);
    }
  };

  const handleExecuteReport = async (params: Record<string, any>) => {
    if (!selectedReport) return;
    
    setIsLoading(true);
    try {
      const response = await reportsApi.execute(selectedReport.id, params);
      console.log('Report executed:', response.data);
      // Handle success - maybe show results in a modal or new tab
    } catch (error) {
      console.error('Failed to execute report:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <MainLayout>
      <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
        <div className="px-6 py-4 border-b bg-white">
          <TabsList className="grid w-full grid-cols-4 max-w-2xl">
            <TabsTrigger value="chat">AI Chat</TabsTrigger>
            <TabsTrigger value="files">Files</TabsTrigger>
            <TabsTrigger value="reports">Reports</TabsTrigger>
            <TabsTrigger value="execute" disabled={!selectedReport}>
              Execute Report
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="chat" className="flex-1 h-full">
          <ChatWindow />
        </TabsContent>

          <TabsContent value="files" className="space-y-4 p-6">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <FileUpload 
                onFileUploaded={(file) => {
                  console.log('File uploaded:', file);
                  // Refresh file list or show success message
                }}
              />
              <FileList 
                onFileSelect={(file) => {
                  console.log('File selected:', file);
                  // Handle file selection for analysis
                }}
                onFileDelete={(fileId) => {
                  console.log('File deleted:', fileId);
                  // Handle file deletion
                }}
              />
            </div>
          </TabsContent>

          <TabsContent value="reports" className="space-y-4 p-6">
            <div className="flex justify-between items-center">
              <h2 className="text-2xl font-semibold">Your Reports</h2>
              <Button>Create New Report</Button>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {(reports || []).map((report) => (
                <ReportCard
                  key={report.id}
                  report={report}
                  onSelect={() => handleReportSelect(report)}
                  onEdit={() => console.log('Edit report:', report.id)}
                  onDelete={() => console.log('Delete report:', report.id)}
                  onExecute={() => handleReportSelect(report)}
                />
              ))}
            </div>
          </TabsContent>

          <TabsContent value="execute" className="space-y-4 p-6">
            {selectedReport && reportSchema ? (
              <div className="space-y-4">
                <div>
                  <h2 className="text-2xl font-semibold">{selectedReport.title}</h2>
                  <p className="text-muted-foreground">{selectedReport.description}</p>
                </div>
                
                <SchemaForm
                  schema={reportSchema.schema}
                  onSubmit={handleExecuteReport}
                  isLoading={isLoading}
                />
              </div>
            ) : (
              <Card>
                <CardContent className="flex items-center justify-center h-32">
                  <p className="text-muted-foreground">Select a report to execute</p>
                </CardContent>
              </Card>
            )}
          </TabsContent>
        </Tabs>
    </MainLayout>
  );
}
