'use client';

import React, { useState, useRef } from 'react';
import { COLORS } from "@/constants"
import { API_ENDPOINTS } from "@/constants/api"

type FileDropzoneProps = Record<string, never>;


interface UploadFileResponse {
  status: boolean;
  data: {
    url: string;
    files: Array<{
      original_name: string;
      file_name: string;
      size: number;
      path: string;
    }>;
    total_size: number;
    file_count: number;
  } | null;
  error: string | null;
}

async function uploadFiles(
  files: FileList,
): Promise<UploadFileResponse> {
  const formData = new FormData();
  
  // Append all files to the form data
  for (let i = 0; i < files.length; i++) {
    formData.append('file', files[i]);
  }

  const res = await fetch(API_ENDPOINTS.UPLOAD, {
    method: "POST",
    body: formData,
  });

  const data = await res.json();
  
  // Check if the API response indicates an error
  if (!res.ok || !data.status) {
    console.log("API error:", res.status, res.statusText);
    throw new Error(data.error || `Upload failed: ${res.statusText}`);
  }

  return data;
}

export default function FileDropzone(_props: FileDropzoneProps) {
  const [isDragActive, setIsDragActive] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<string>('');
  const [uploadedFiles, setUploadedFiles] = useState<File[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragEnter = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(true);
  };

  const handleDragLeave = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);
  };

  const handleDragOver = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleBrowseClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileUpload = async (files: FileList) => {
    if (files.length === 0) return;
    
    setIsUploading(true);
    setUploadProgress(`Uploading ${files.length} file${files.length > 1 ? 's' : ''}...`);
    
    try {
      const res = await uploadFiles(files);
      console.log("Upload Successful:", res);
      
      // Store uploaded files for display
      const fileArray = Array.from(files);
      setUploadedFiles(prev => [...prev, ...fileArray]);
      
      // Show success message with link - access data from the new response structure
      if (res.data) {
        setUploadProgress(`✅ Upload successful! Access your files at: /files/s/${res.data.url}`);
      }
      
      // Clear the file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
      
      // Reset after 5 seconds
      setTimeout(() => {
        setUploadProgress('');
        setUploadedFiles([]);
      }, 5000);
      
    } catch (error) {
      console.error("Upload error:", error);
      setUploadProgress(`❌ Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
      
      // Reset after 3 seconds
      setTimeout(() => {
        setUploadProgress('');
      }, 3000);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileUpload(files);
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      handleFileUpload(files);
    }
  };

  return (
    <div
      className="rounded-xl p-8 shadow-2xl backdrop-blur-lg bg-slate-500 text-slate-900"
    >
      <div className="text-center mb-8">
        <h2 className="text-3xl font-bold mb-2">
          Share Files Anonymously
        </h2>
        <p>
          No accounts. No tracking. Just upload and share.
        </p>
        {uploadProgress && (
          <div className="mt-4 p-3 rounded-lg" style={{ backgroundColor: uploadProgress.includes('✅') ? '#d1fae5' : uploadProgress.includes('❌') ? '#fee2e2' : '#dbeafe' }}>
            <p className="text-sm font-medium">{uploadProgress}</p>
          </div>
        )}
      </div>
      <div
        className="border-2 border-dashed rounded-lg p-12 mb-6 text-center cursor-pointer transition-all"
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onClick={handleBrowseClick}
        style={{
          borderColor: isDragActive ? COLORS.accent : COLORS.foreground,
          backgroundColor: isDragActive ? COLORS.dragActiveBackground : 'transparent',
        }}
      >
        <div className="flex flex-col items-center justify-center">
          {isUploading ? (
            <div className="w-16 h-16 mb-4 flex items-center justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2" style={{ borderColor: COLORS.accent }}></div>
            </div>
          ) : (
            <svg
              className="w-16 h-16 mb-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
              style={{ color: COLORS.foreground }}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
              />
            </svg>
          )}
          <h3 className="text-xl font-semibold mb-2">
            {isUploading ? 'Uploading Files...' : 'Drag & Drop Files Here'}
          </h3>
          <p className="mb-4">
            {isUploading ? 'Please wait while your files are being uploaded' : 'or click to browse (multiple files supported)'}
          </p>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            multiple
            onChange={handleFileInputChange}
          />
          <button
            className="py-2 px-6 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            style={{
              backgroundColor: COLORS.accent,
              color: COLORS.accentForeground,
            }}
            onClick={(e) => {
              e.stopPropagation();
              handleBrowseClick();
            }}
            type="button"
            disabled={isUploading}
          >
            {isUploading ? 'Uploading...' : 'Select Files'}
          </button>
        </div>
      </div>
      
      {/* Recently uploaded files */}
      {uploadedFiles.length > 0 && (
        <div className="mt-6">
          <h3 className="text-lg font-semibold mb-3">Recently Uploaded Files:</h3>
          <div className="space-y-2">
            {uploadedFiles.map((file, index) => (
              <div key={index} className="flex items-center justify-between p-3 bg-gray-100 rounded-lg">
                <div className="flex items-center">
                  <svg className="w-5 h-5 mr-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <span className="text-sm font-medium">{file.name}</span>
                </div>
                <span className="text-xs text-gray-500">
                  {(file.size / 1024).toFixed(1)} KB
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
