import { useState, useRef } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Upload, File, CheckCircle, AlertCircle } from 'lucide-react';
import { api } from '@/lib/api';

interface UploadedFile {
  file_id: string;
  filename: string;
  file_size: number;
  upload_time: string;
  file_type: string;
}

interface FileUploadProps {
  onFileUploaded?: (file: UploadedFile) => void;
  onFileSelected?: (file: File) => void;
}

export function FileUpload({ onFileUploaded, onFileSelected }: FileUploadProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [filename, setFilename] = useState('');
  const [fileType, setFileType] = useState('');
  const [description, setDescription] = useState('');
  const [isUploading, setIsUploading] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<'idle' | 'success' | 'error'>('idle');
  const [uploadMessage, setUploadMessage] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setFilename(file.name);
      
      // Auto-detect file type from extension
      const extension = file.name.split('.').pop()?.toLowerCase();
      if (extension) {
        setFileType(extension);
      }
      
      if (onFileSelected) {
        onFileSelected(file);
      }
    }
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    setUploadStatus('idle');

    try {
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('filename', filename || selectedFile.name);
      formData.append('file_type', fileType);
      formData.append('description', description);

      const response = await api.post('/v1/upload/file', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      setUploadStatus('success');
      setUploadMessage(`File uploaded successfully: ${response.data.filename}`);
      
      if (onFileUploaded) {
        onFileUploaded({
          file_id: response.data.file_id,
          filename: response.data.filename,
          file_size: response.data.file_size,
          upload_time: response.data.upload_time,
          file_type: response.data.file_type,
        });
      }

      // Reset form
      setSelectedFile(null);
      setFilename('');
      setFileType('');
      setDescription('');
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (error: any) {
      setUploadStatus('error');
      setUploadMessage(error.response?.data?.details || 'Upload failed');
    } finally {
      setIsUploading(false);
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <Card className="w-full max-w-2xl">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Upload className="h-5 w-5" />
          Upload File
        </CardTitle>
        <CardDescription>
          Upload CSV, Parquet, or JSON files for data analysis
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* File Selection */}
        <div className="space-y-2">
          <Label htmlFor="file">Select File</Label>
          <Input
            ref={fileInputRef}
            id="file"
            type="file"
            accept=".csv,.parquet,.json,.jsonl"
            onChange={handleFileSelect}
            className="cursor-pointer"
          />
        </div>

        {/* File Info */}
        {selectedFile && (
          <div className="p-3 bg-muted rounded-lg">
            <div className="flex items-center gap-2 mb-2">
              <File className="h-4 w-4" />
              <span className="font-medium">{selectedFile.name}</span>
              <Badge variant="outline">{formatFileSize(selectedFile.size)}</Badge>
            </div>
            <div className="text-sm text-muted-foreground">
              Type: {selectedFile.type || 'Unknown'}
            </div>
          </div>
        )}

        {/* Filename */}
        <div className="space-y-2">
          <Label htmlFor="filename">Filename</Label>
          <Input
            id="filename"
            value={filename}
            onChange={(e) => setFilename(e.target.value)}
            placeholder="Enter custom filename (optional)"
          />
        </div>

        {/* File Type */}
        <div className="space-y-2">
          <Label htmlFor="fileType">File Type</Label>
          <Select value={fileType} onValueChange={setFileType}>
            <SelectTrigger>
              <SelectValue placeholder="Select file type" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="csv">CSV</SelectItem>
              <SelectItem value="parquet">Parquet</SelectItem>
              <SelectItem value="json">JSON</SelectItem>
              <SelectItem value="jsonl">JSONL</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* Description */}
        <div className="space-y-2">
          <Label htmlFor="description">Description (Optional)</Label>
          <Input
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe your data..."
          />
        </div>

        {/* Upload Status */}
        {uploadStatus !== 'idle' && (
          <div className={`flex items-center gap-2 p-3 rounded-lg ${
            uploadStatus === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
          }`}>
            {uploadStatus === 'success' ? (
              <CheckCircle className="h-4 w-4" />
            ) : (
              <AlertCircle className="h-4 w-4" />
            )}
            <span className="text-sm">{uploadMessage}</span>
          </div>
        )}

        {/* Upload Button */}
        <Button
          onClick={handleUpload}
          disabled={!selectedFile || isUploading}
          className="w-full"
        >
          {isUploading ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Uploading...
            </>
          ) : (
            <>
              <Upload className="h-4 w-4 mr-2" />
              Upload File
            </>
          )}
        </Button>
      </CardContent>
    </Card>
  );
}
