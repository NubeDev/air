import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { MainLayout } from '@/components/layout/MainLayout';
import { SchemaForm } from '@/components/forms/SchemaForm';
import { Card, CardContent } from '@/components/ui/card';
import { reportsApi } from '@/services/reportsApi';
import type { Report, ReportSchema } from '@/types/api';

export function ExecuteReportPage() {
  const { id } = useParams();
  const [report, setReport] = useState<Report | null>(null);
  const [schema, setSchema] = useState<ReportSchema | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!id) return;
    (async () => {
      try {
        const [r, s] = await Promise.all([
          reportsApi.get(id),
          reportsApi.getSchema(id),
        ]);
        setReport(r.data);
        setSchema(s.data);
      } catch (e) {
        setReport(null);
        setSchema(null);
      }
    })();
  }, [id]);

  const handleExecute = async (params: Record<string, any>) => {
    if (!id) return;
    setIsLoading(true);
    try {
      await reportsApi.execute(id, params);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <MainLayout>
      <div className="p-6">
        {report && schema ? (
          <div className="space-y-4">
            <div>
              <h2 className="text-2xl font-semibold">{report.title}</h2>
              {report.description && (
                <p className="text-muted-foreground">{report.description}</p>
              )}
            </div>
            <SchemaForm schema={schema.schema} onSubmit={handleExecute} isLoading={isLoading} />
          </div>
        ) : (
          <Card>
            <CardContent className="flex items-center justify-center h-32">
              <p className="text-muted-foreground">Loading report…</p>
            </CardContent>
          </Card>
        )}
      </div>
    </MainLayout>
  );
}


