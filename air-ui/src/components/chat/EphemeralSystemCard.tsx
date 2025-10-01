import React from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Upload, FileText, X, Loader2 } from 'lucide-react';

interface UploadedFile {
  file_id: string;
  filename: string;
  file_size: number;
  upload_time: string;
  file_type: string;
}

interface EphemeralSystemCardProps {
  type: 'file_needed' | 'file_loaded' | 'uploading';
  files?: UploadedFile[];
  onFileSelect?: (fileId: string) => void;
  onUploadClick?: () => void;
  onDismiss?: () => void;
  isUploading?: boolean;
  selectedFileName?: string;
}

export function EphemeralSystemCard({
  type,
  files = [],
  onFileSelect,
  onUploadClick,
  onDismiss,
  isUploading = false,
  selectedFileName
}: EphemeralSystemCardProps) {
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  if (type === 'file_loaded') {
    return (
      <Card className="bg-green-50 border-green-200 p-4 mb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
              <FileText className="h-4 w-4 text-green-600" />
            </div>
            <div>
              <p className="text-sm font-medium text-green-800">
                ✅ Dataset loaded: {selectedFileName}
              </p>
              <p className="text-xs text-green-600">
                Ready for analysis
              </p>
            </div>
          </div>
          {onDismiss && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onDismiss}
              className="text-green-600 hover:text-green-700"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
      </Card>
    );
  }

  if (type === 'uploading') {
    return (
      <Card className="bg-blue-50 border-blue-200 p-4 mb-4">
        <div className="flex items-center gap-3">
          <Loader2 className="h-4 w-4 text-blue-600 animate-spin" />
          <p className="text-sm font-medium text-blue-800">
            Uploading file...
          </p>
        </div>
      </Card>
    );
  }

  // type === 'file_needed'
  return (
    <Card className="bg-amber-50 border-amber-200 p-4 mb-4">
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <div className="w-8 h-8 bg-amber-100 rounded-full flex items-center justify-center">
            <FileText className="h-4 w-4 text-amber-600" />
          </div>
          <div className="flex-1">
            <h4 className="text-sm font-medium text-amber-800 mb-2">
              📁 No dataset loaded
            </h4>
            <p className="text-xs text-amber-700 mb-3">
              To analyze data, you need to load a file first. Choose from your uploaded files or upload a new one.
            </p>
            
            {files.length > 0 && (
              <div className="space-y-2">
                <p className="text-xs font-medium text-amber-800">Available files:</p>
                <div className="space-y-1 max-h-32 overflow-y-auto">
                  {files.map((file) => (
                    <div
                      key={file.file_id}
                      className="flex items-center justify-between p-2 bg-white rounded border border-amber-200 hover:border-amber-300 cursor-pointer group"
                      onClick={() => onFileSelect?.(file.file_id)}
                    >
                      <div className="flex items-center gap-2">
                        <FileText className="h-3 w-3 text-amber-600" />
                        <span className="text-xs font-medium text-gray-700 group-hover:text-amber-700">
                          {file.filename}
                        </span>
                        <span className="text-xs text-gray-500">
                          ({formatFileSize(file.file_size)})
                        </span>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="opacity-0 group-hover:opacity-100 text-amber-600 hover:text-amber-700 h-6 px-2"
                      >
                        Load
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            <div className="flex gap-2 mt-3">
              <Button
                variant="outline"
                size="sm"
                onClick={onUploadClick}
                className="text-amber-700 border-amber-300 hover:bg-amber-100"
              >
                <Upload className="h-3 w-3 mr-1" />
                Upload File
              </Button>
              {files.length > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onDismiss}
                  className="text-amber-600 hover:text-amber-700"
                >
                  Cancel
                </Button>
              )}
            </div>
          </div>
        </div>
        {onDismiss && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onDismiss}
            className="text-amber-600 hover:text-amber-700"
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>
    </Card>
  );
}
