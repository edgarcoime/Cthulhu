'use client';

import React, { useState, useRef } from 'react';
import { COLORS } from "@/constants"

type FileDropzoneProps = {
};

const API_UPLOAD_URL = "https://api.escuelajs.co/api/v1/files/upload";

interface UploadFilePayload {
}

interface UploadFileResponse {
}

async function uploadFiles(
  files: FileList,
): Promise<any> {
  const formData = new FormData();
  formData.append('file', files[0])

  const res = await fetch(API_UPLOAD_URL, {
    method: "POST",
    body: formData,
  });

  if (!res.ok) {
    console.log("API error:", res.status, res.statusText);
    throw new Error(`Upload failed: ${res.statusText}`)
  }

  const data = await res.json()
  return data;
}

export default function FideDropzone({ }: FileDropzoneProps) {
  const [isDragActive, setIsDragActive] = useState(false);
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

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      uploadFiles(files)
        .then(res => {
          console.log("Upload Successful:", res)
        })
        .catch(error => {
          console.error("Upload error:", error)
        })
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      uploadFiles(files)
        .then(res => {
          console.log("Upload Successful:", res)
        })
        .catch(error => {
          console.error("Upload error:", error)
        })
    }
  };

  return (
    <div
      className="rounded-xl p-8 shadow-2xl backdrop-blur-lg"
      style={{
        backgroundColor: COLORS.secondaryBackground,
        color: COLORS.foreground,
      }}
    >
      <div className="text-center mb-8">
        <h2 className="text-3xl font-bold mb-2" style={{ color: COLORS.foreground }}>
          Share Files Anonymously
        </h2>
        <p style={{ color: COLORS.foreground }}>
          No accounts. No tracking. Just upload and share.
        </p>
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
          <h3 className="text-xl font-semibold mb-2" style={{ color: COLORS.foreground }}>
            Drag & Drop Files Here
          </h3>
          <p className="mb-4" style={{ color: COLORS.foreground }}>
            or click to browse
          </p>
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            multiple
            onChange={handleFileInputChange}
          />
          <button
            className="py-2 px-6 rounded-lg transition-colors"
            style={{
              backgroundColor: COLORS.accent,
              color: COLORS.accentForeground,
            }}
            onClick={(e) => {
              e.stopPropagation();
              handleBrowseClick();
            }}
            type="button"
          >
            Select Files
          </button>
        </div>
      </div>
    </div>
  );
}
