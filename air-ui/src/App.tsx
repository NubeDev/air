import './index.css';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ChatPage } from '@/pages/Chat';
import { FilesPage } from '@/pages/Files';
import { ReportsPage } from '@/pages/Reports';
import { ExecuteReportPage } from '@/pages/ExecuteReport';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/chat" replace />} />
        <Route path="/chat" element={<ChatPage />} />
        <Route path="/files" element={<FilesPage />} />
        <Route path="/reports" element={<ReportsPage />} />
        <Route path="/reports/:id/execute" element={<ExecuteReportPage />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;