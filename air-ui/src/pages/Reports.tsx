import { useEffect, useState } from 'react';
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
    <div className="min-h-full bg-white p-6 space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-semibold text-foreground">Your Reports</h2>
        <Button onClick={() => console.log('Create New Report clicked')}>Create New Report</Button>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {(reports || []).map((report) => (
          <ReportCard
            key={report.id}
            report={report}
            onSelect={() => navigate(`/reports/${report.id}/execute`)}
            onEdit={() => console.log('Edit report:', report.id)}
            onDelete={() => console.log('Delete report:', report.id)}
            onExecute={() => navigate(`/reports/${report.id}/execute`)}
          />
        ))}
      </div>
    </div>
  );
}


