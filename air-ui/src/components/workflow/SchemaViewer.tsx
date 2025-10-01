import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Database, 
  FileText, 
  CheckCircle,
  Loader2,
  AlertCircle,
  Eye,
  Download
} from 'lucide-react';

interface SchemaViewerProps {
  sessionId?: string;
  sessionType?: 'file' | 'database';
  onSchemaLoaded: (data: { schemaData: any }) => void;
}

export function SchemaViewer({ sessionId, sessionType, onSchemaLoaded }: SchemaViewerProps) {
  const [schemaData, setSchemaData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (sessionId && sessionType) {
      loadSchema();
    }
  }, [sessionId, sessionType]);

  const loadSchema = async () => {
    if (!sessionId || !sessionType) return;

    setIsLoading(true);
    setError(null);

    try {
      let response;
      
      if (sessionType === 'file') {
        // First get session details to find the file ID
        const sessionResponse = await fetch(`/v1/sessions/${sessionId}`);
        if (!sessionResponse.ok) {
          throw new Error(`Failed to get session: ${sessionResponse.statusText}`);
        }
        
        const sessionData = await sessionResponse.json();
        const options = JSON.parse(sessionData.options || '{}');
        const fileId = options.file_id;
        
        if (!fileId) {
          throw new Error('No file ID found in session');
        }

        // Load file schema using the new endpoint
        response = await fetch(`/v1/upload/file/${fileId}/learn`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        });
      } else {
        // Load database schema
        response = await fetch(`/v1/schema/${sessionId}`, {
          method: 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        });
      }

      if (!response.ok) {
        throw new Error(`Schema loading failed: ${response.statusText}`);
      }

      const data = await response.json();
      
      // Extract schema data from the response
      const schemaData = data.schema_data || data;
      setSchemaData(schemaData);
      onSchemaLoaded({ schemaData: schemaData });

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load schema');
    } finally {
      setIsLoading(false);
    }
  };

  const renderFileSchema = (schema: any) => (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">File Info</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Filename:</span>
                <span className="font-medium">{schema.filename || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Size:</span>
                <span className="font-medium">{schema.file_size || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Type:</span>
                <span className="font-medium">{schema.file_type || 'Unknown'}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Data Overview</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Rows:</span>
                <span className="font-medium">{schema.row_count || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Columns:</span>
                <span className="font-medium">{schema.column_count || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Encoding:</span>
                <span className="font-medium">{schema.encoding || 'Unknown'}</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Quality</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Completeness:</span>
                <Badge variant="outline" className="text-xs">
                  {schema.completeness || 'Unknown'}%
                </Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Duplicates:</span>
                <Badge variant="outline" className="text-xs">
                  {schema.duplicate_count || 'Unknown'}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Column Schema */}
      {schema.columns && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Column Schema</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2">Column</th>
                    <th className="text-left py-2">Type</th>
                    <th className="text-left py-2">Nulls</th>
                    <th className="text-left py-2">Unique</th>
                    <th className="text-left py-2">Sample</th>
                  </tr>
                </thead>
                <tbody>
                  {schema.columns.map((col: any, index: number) => (
                    <tr key={index} className="border-b">
                      <td className="py-2 font-medium">{col.name}</td>
                      <td className="py-2">
                        <Badge variant="outline" className="text-xs">
                          {col.type}
                        </Badge>
                      </td>
                      <td className="py-2">{col.null_count || 0}</td>
                      <td className="py-2">{col.unique_count || 0}</td>
                      <td className="py-2 text-muted-foreground">
                        {col.sample_value || 'N/A'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Sample Data */}
      {schema.sample_data && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Sample Data</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    {schema.sample_data[0] && Object.keys(schema.sample_data[0]).map((key) => (
                      <th key={key} className="text-left py-2">{key}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {schema.sample_data.slice(0, 5).map((row: any, index: number) => (
                    <tr key={index} className="border-b">
                      {Object.values(row).map((value: any, colIndex: number) => (
                        <td key={colIndex} className="py-2">{String(value)}</td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderDatabaseSchema = (schema: any) => (
    <div className="space-y-4">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Database Info</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Type:</span>
                <span className="font-medium">{schema.database_type || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Tables:</span>
                <span className="font-medium">{schema.table_count || 'Unknown'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Connection:</span>
                <Badge variant="outline" className="text-xs">
                  {schema.connection_status || 'Unknown'}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Schema Notes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              {schema.schema_notes || 'No schema notes available'}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tables */}
      {schema.tables && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium">Tables</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {schema.tables.map((table: any, index: number) => (
                <div key={index} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="font-medium">{table.name}</h4>
                    <Badge variant="outline" className="text-xs">
                      {table.row_count || 'Unknown'} rows
                    </Badge>
                  </div>
                  {table.columns && (
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-2 text-sm">
                      {table.columns.map((col: any, colIndex: number) => (
                        <div key={colIndex} className="flex items-center space-x-2">
                          <span className="text-muted-foreground">{col.name}:</span>
                          <Badge variant="outline" className="text-xs">
                            {col.type}
                          </Badge>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );

  if (isLoading) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-center space-x-2">
            <Loader2 className="h-5 w-5 animate-spin" />
            <span>Loading schema...</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center space-x-2 text-destructive">
            <AlertCircle className="h-5 w-5" />
            <span>{error}</span>
          </div>
          <Button 
            onClick={loadSchema} 
            variant="outline" 
            size="sm" 
            className="mt-4"
          >
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (!schemaData) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="text-center text-muted-foreground">
            <Database className="h-12 w-12 mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">No Schema Data</h3>
            <p>Start a session to load schema information</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Database className="h-5 w-5 mr-2" />
            Schema Discovery
            <Badge variant="outline" className="ml-2">
              {sessionType === 'file' ? 'File' : 'Database'}
            </Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {sessionType === 'file' ? renderFileSchema(schemaData) : renderDatabaseSchema(schemaData)}
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button className="flex items-center">
          <CheckCircle className="h-4 w-4 mr-2" />
          Schema Loaded
        </Button>
      </div>
    </div>
  );
}
