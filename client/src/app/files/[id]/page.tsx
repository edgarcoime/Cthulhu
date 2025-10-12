'use client';

import React, { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { COLORS } from "@/constants";

interface FileInfo {
  name: string;
  filename: string;
  size: number;
  url: string;
}

interface FileResponse {
  status: boolean;
  data: {
    session_id: string;
    files: FileInfo[];
    count: number;
  } | null;
  error: string | null;
}

const API_BASE_URL = "http://localhost:4000";

export default function FileAccessPage() {
  const params = useParams();
  const id = params.id as string;
  const [files, setFiles] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFiles = async () => {
      try {
        setLoading(true);
        setError(null);
        
        const response = await fetch(`${API_BASE_URL}/files/${id}`);
        const data: FileResponse = await response.json();
        
        // Check if the API response indicates an error
        if (!response.ok || !data.status) {
          throw new Error(data.error || 'Failed to fetch files');
        }
        
        // Extract files from the new response structure
        if (data.data) {
          setFiles(data.data.files);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    if (id) {
      fetchFiles();
    }
  }, [id]);

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const handleDownload = (file: FileInfo) => {
    window.open(`${API_BASE_URL}${file.url}`, '_blank');
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 mx-auto mb-4" style={{ borderColor: COLORS.accent }}></div>
          <p className="text-lg">Loading files...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center max-w-md mx-auto p-8">
          <div className="mb-6">
            <svg className="w-16 h-16 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ color: '#ef4444' }}>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 19.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-4" style={{ color: '#ef4444' }}>Error</h2>
          <p className="text-lg mb-6">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="py-2 px-6 rounded-lg transition-colors"
            style={{
              backgroundColor: COLORS.accent,
              color: COLORS.accentForeground,
            }}
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen p-8">
      <div className="max-w-4xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-2">File Access</h1>
          <p className="text-lg text-gray-600">Session ID: <code className="bg-gray-100 px-2 py-1 rounded">{id}</code></p>
        </div>

        {files.length === 0 ? (
          <div className="text-center py-12">
            <svg className="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p className="text-xl text-gray-500">No files found in this session</p>
          </div>
        ) : (
          <div className="space-y-4">
            <div className="mb-6">
              <p className="text-lg font-semibold">{files.length} file{files.length !== 1 ? 's' : ''} available</p>
            </div>
            
            {files.map((file, index) => (
              <div
                key={index}
                className="border rounded-lg p-6 hover:shadow-lg transition-shadow"
                style={{ borderColor: COLORS.foreground }}
              >
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold mb-2">{file.name}</h3>
                    <p className="text-gray-600">{formatFileSize(file.size)}</p>
                  </div>
                  <button
                    onClick={() => handleDownload(file)}
                    className="py-2 px-6 rounded-lg transition-colors"
                    style={{
                      backgroundColor: COLORS.accent,
                      color: COLORS.accentForeground,
                    }}
                  >
                    Download
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
