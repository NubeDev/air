import { useState, useEffect } from 'react';
import { MainLayout } from '@/components/layout/MainLayout';
import { FileUpload } from '@/components/upload/FileUpload';
import { FileList } from '@/components/upload/FileList';

export function FilesPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    // placeholder to potentially sync with chat selected file later
  }, [refreshKey]);

  return (
    <MainLayout>
      <div className="p-6 space-y-4">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <FileUpload onFileUploaded={() => setRefreshKey((k) => k + 1)} />
          <FileList onFileSelect={() => {}} onFileDelete={() => setRefreshKey((k) => k + 1)} />
        </div>
      </div>
    </MainLayout>
  );
}


