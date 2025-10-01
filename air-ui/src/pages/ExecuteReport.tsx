import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
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
    <div className="min-h-full bg-white p-6 space-y-6">
      {report && schema ? (
        <div className="space-y-4">
          <div>
            <h2 className="text-2xl font-semibold text-foreground">{report.title}</h2>
            {report.description && (
              <p className="text-muted-foreground">{report.description}</p>
            )}
          </div>
          <SchemaForm schema={schema.schema} onSubmit={handleExecute} isLoading={isLoading} />
        </div>
      ) : (
        <Card>
          <CardContent className="flex items-center justify-center h-32">
            <p className="text-muted-foreground">Loading report details or report not found...</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}


