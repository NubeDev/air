import { useState, useEffect } from 'react';
import { FileUpload } from '@/components/upload/FileUpload';
import { FileList } from '@/components/upload/FileList';

export function FilesPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    // placeholder to potentially sync with chat selected file later
  }, [refreshKey]);

  return (
    <div className="min-h-full bg-white p-6 space-y-6">
      <h2 className="text-2xl font-semibold text-foreground">Files</h2>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <FileUpload onFileUploaded={() => setRefreshKey((k) => k + 1)} />
        <FileList onFileSelect={() => {}} onFileDelete={() => setRefreshKey((k) => k + 1)} />
      </div>
    </div>
  );
}


