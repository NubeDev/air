import './index.css';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { MainLayout } from '@/components/layout/MainLayout';
import { ChatPage } from '@/pages/Chat';
import { WorkflowPage } from '@/pages/Workflow';
import { FilesPage } from '@/pages/Files';
import { ReportsPage } from '@/pages/Reports';
import { ExecuteReportPage } from '@/pages/ExecuteReport';

function App() {
  return (
    <BrowserRouter>
      <MainLayout>
        <Routes>
          <Route path="/" element={<Navigate to="/chat" replace />} />
          <Route path="/chat" element={<ChatPage />} />
          <Route path="/workflow" element={<WorkflowPage />} />
          <Route path="/files" element={<FilesPage />} />
          <Route path="/reports" element={<ReportsPage />} />
          <Route path="/reports/:id/execute" element={<ExecuteReportPage />} />
        </Routes>
      </MainLayout>
    </BrowserRouter>
  );
}

export default App;