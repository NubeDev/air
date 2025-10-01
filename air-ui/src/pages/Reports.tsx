import { useEffect, useState } from 'react';
import { MainLayout } from '@/components/layout/MainLayout';
import { ReportCard } from '@/components/reports/ReportCard';
import { Button } from '@/components/ui/button';
import { reportsApi } from '@/services/reportsApi';
import type { Report } from '@/types/api';
import { useNavigate } from 'react-router-dom';

export function ReportsPage() {
  const [reports, setReports] = useState<Report[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    (async () => {
      try {
        const response = await reportsApi.list();
        const data = Array.isArray(response.data)
          ? response.data
          : (response.data as any)?.reports ?? [];
        setReports(data);
      } catch (e) {
        setReports([]);
      }
    })();
  }, []);

  return (
    <MainLayout>
      <div className="p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-semibold">Your Reports</h2>
          <Button>Create New Report</Button>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {reports.map((report) => (
            <ReportCard
              key={report.id}
              report={report}
              onSelect={() => navigate(`/reports/${report.id}/execute`)}
              onEdit={() => navigate(`/reports/${report.id}/execute`)}
              onDelete={() => {}}
              onExecute={() => navigate(`/reports/${report.id}/execute`)}
            />
          ))}
        </div>
      </div>
    </MainLayout>
  );
}


