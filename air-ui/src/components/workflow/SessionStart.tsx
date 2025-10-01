import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Upload, 
  Database, 
  FileText, 
  CheckCircle,
  Loader2,
  AlertCircle,
  Play,
  ArrowRight
} from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface SessionStartProps {
  onSessionStart: (data: { sessionId: string; sessionType: 'file' | 'database' }) => void;
  onStepComplete?: () => void;
  sessionType?: 'file' | 'database';
}

export function SessionStart({ onSessionStart, onStepComplete, sessionType }: SessionStartProps) {
  const [selectedType, setSelectedType] = useState<'file' | 'database' | null>(sessionType || null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleFileUpload = async (file: File) => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      console.log('Creating FormData for file:', file.name, 'Size:', file.size);
      // Create FormData for file upload
      const formData = new FormData();
      formData.append('file', file);

      console.log('Sending upload request to /v1/upload/file');
      // Upload file to backend
      const uploadResponse = await fetch('/v1/upload/file', {
        method: 'POST',
        body: formData,
      });

      console.log('Upload response status:', uploadResponse.status);

      if (!uploadResponse.ok) {
        throw new Error(`Upload failed: ${uploadResponse.statusText}`);
      }

      const uploadData = await uploadResponse.json();
      console.log('Upload successful:', uploadData);
      
      // Start session with uploaded file
      console.log('Starting session with file:', uploadData.file_path);
      const sessionResponse = await fetch('/v1/sessions/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          file_path: uploadData.file_path,
          session_name: `File Analysis - ${uploadData.filename}`,
          datasource_type: 'file',
          options: {
            file_id: uploadData.file_id,
            filename: uploadData.filename,
            file_type: uploadData.file_type
          }
        }),
      });

      if (!sessionResponse.ok) {
        throw new Error(`Session start failed: ${sessionResponse.statusText}`);
      }

      const sessionData = await sessionResponse.json();
      
      onSessionStart({
        sessionId: sessionData.id.toString(),
        sessionType: 'file'
      });

      setSuccess(`File uploaded and session started successfully! Session ID: ${sessionData.id}`);
      
      // Auto-advance to next step after a short delay
      setTimeout(() => {
        onStepComplete?.();
      }, 2000);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start file session');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDatabaseConnect = async () => {
    setIsLoading(true);
    setError(null);

    try {
      // For now, we'll create a mock database session
      // In a real implementation, this would connect to a registered datasource
      const sessionResponse = await fetch('/v1/sessions/start', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          file_path: 'database://default',
          session_name: 'Database Analysis Session',
          datasource_type: 'database',
          options: {
            datasource_id: 'default'
          }
        }),
      });

      if (!sessionResponse.ok) {
        throw new Error(`Session start failed: ${sessionResponse.statusText}`);
      }

      const sessionData = await sessionResponse.json();
      
      onSessionStart({
        sessionId: sessionData.id.toString(),
        sessionType: 'database'
      });

      setSuccess(`Database connected and session started successfully! Session ID: ${sessionData.id}`);
      
      // Auto-advance to next step after a short delay
      setTimeout(() => {
        onStepComplete?.();
      }, 2000);

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start database session');
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    console.log('File selected:', file);
    if (file) {
      console.log('Starting file upload for:', file.name);
      handleFileUpload(file);
    } else {
      console.log('No file selected');
    }
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Play className="h-5 w-5 mr-2" />
            Start New Session
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Session Type Selection */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* File Upload Option */}
            <Card 
              className={`cursor-pointer transition-all duration-200 hover:shadow-md ${
                selectedType === 'file' ? 'ring-2 ring-primary bg-primary/5' : ''
              }`}
              onClick={() => setSelectedType('file')}
            >
              <CardContent className="p-6 text-center">
                <FileText className="h-12 w-12 mx-auto mb-4 text-primary" />
                <h3 className="text-lg font-semibold mb-2">Upload File</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Upload CSV, JSON, or Excel files for analysis
                </p>
                <Badge variant="outline" className="text-xs">
                  CSV, JSON, XLSX, TXT
                </Badge>
              </CardContent>
            </Card>

            {/* Database Connection Option */}
            <Card 
              className={`cursor-pointer transition-all duration-200 hover:shadow-md ${
                selectedType === 'database' ? 'ring-2 ring-primary bg-primary/5' : ''
              }`}
              onClick={() => setSelectedType('database')}
            >
              <CardContent className="p-6 text-center">
                <Database className="h-12 w-12 mx-auto mb-4 text-primary" />
                <h3 className="text-lg font-semibold mb-2">Connect Database</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Connect to SQLite, PostgreSQL, MySQL, or TimescaleDB
                </p>
                <Badge variant="outline" className="text-xs">
                  SQLite, Postgres, MySQL, Timescale
                </Badge>
              </CardContent>
            </Card>
          </div>

          {/* Action Buttons */}
          {!selectedType && (
            <div className="text-center py-4">
              <p className="text-muted-foreground">Please select a session type above to continue</p>
            </div>
          )}
          {selectedType && (
            <div className="flex justify-center space-x-4">
              {selectedType === 'file' ? (
                <div className="text-center">
                  <input
                    type="file"
                    accept=".csv,.json,.xlsx,.txt"
                    onChange={handleFileSelect}
                    className="hidden"
                    id="file-upload"
                  />
                  <Button
                    onClick={() => document.getElementById('file-upload')?.click()}
                    disabled={isLoading}
                    className="flex items-center"
                  >
                    {isLoading ? (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    ) : (
                      <Upload className="h-4 w-4 mr-2" />
                    )}
                    {isLoading ? 'Uploading...' : 'Choose File'}
                  </Button>
                </div>
              ) : (
                <Button
                  onClick={handleDatabaseConnect}
                  disabled={isLoading}
                  className="flex items-center"
                >
                  {isLoading ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Database className="h-4 w-4 mr-2" />
                  )}
                  {isLoading ? 'Connecting...' : 'Connect Database'}
                </Button>
              )}
            </div>
          )}

          {/* Error Display */}
          {error && (
            <div className="flex items-center space-x-2 text-destructive bg-destructive/10 p-3 rounded-md">
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          {/* Success Display */}
          {success && (
            <div className="space-y-3">
              <div className="flex items-center space-x-2 text-green-800 bg-green-50 p-3 rounded-md">
                <CheckCircle className="h-4 w-4" />
                <span className="text-sm">{success}</span>
              </div>
              <div className="text-center">
                <Button
                  onClick={() => onStepComplete?.()}
                  className="bg-primary text-primary-foreground hover:bg-primary/90"
                >
                  Continue to Next Step
                  <ArrowRight className="h-4 w-4 ml-2" />
                </Button>
              </div>
            </div>
          )}

          {/* Quick Actions */}
          <div className="border-t pt-4">
            <h4 className="text-sm font-medium text-muted-foreground mb-3">Quick Actions</h4>
            <div className="flex space-x-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate('/files')}
                className="flex items-center"
              >
                <FileText className="h-4 w-4 mr-2" />
                Browse Files
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => navigate('/reports')}
                className="flex items-center"
              >
                <Database className="h-4 w-4 mr-2" />
                View Reports
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
