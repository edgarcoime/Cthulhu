"use client";

import { useEffect, useRef, useState } from "react";

// Type declarations for external libraries
declare global {
  interface Window {
    AOS: {
      init: () => void;
      refresh: () => void;
    };
    feather: {
      replace: () => void;
    };
    VANTA: {
      GLOBE: (config: {
        el: HTMLElement;
        mouseControls: boolean;
        touchControls: boolean;
        gyroControls: boolean;
        minHeight: number;
        minWidth: number;
        scale: number;
        scaleMobile: number;
        color: number;
        backgroundColor: number;
        size: number;
      }) => unknown;
    };
  }
}

export default function Home() {
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [showShareLink, setShowShareLink] = useState(false);
  const [shareLink, setShareLink] = useState("");
  const [isDragActive, setIsDragActive] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);
  
  const vantaRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Initialize AOS and Feather icons
    if (typeof window !== "undefined") {
      // Load AOS
      const aosScript = document.createElement("script");
      aosScript.src = "https://unpkg.com/aos@2.3.1/dist/aos.js";
      aosScript.onload = () => {
        if (window.AOS) {
          window.AOS.init();
        }
      };
      document.head.appendChild(aosScript);

      // Load Feather icons
      const featherScript = document.createElement("script");
      featherScript.src = "https://unpkg.com/feather-icons";
      featherScript.onload = () => {
        if (window.feather) {
          window.feather.replace();
        }
      };
      document.head.appendChild(featherScript);

      // Load Vanta.js
      const vantaScript = document.createElement("script");
      vantaScript.src = "https://cdn.jsdelivr.net/npm/vanta@latest/dist/vanta.globe.min.js";
      vantaScript.onload = () => {
        if (window.VANTA && vantaRef.current) {
          window.VANTA.GLOBE({
            el: vantaRef.current,
            mouseControls: true,
            touchControls: true,
            gyroControls: false,
            minHeight: 200.0,
            minWidth: 200.0,
            scale: 1.0,
            scaleMobile: 1.0,
            color: 0x5a3d8a,
            backgroundColor: 0x111827,
            size: 0.8,
          });
        }
      };
      document.head.appendChild(vantaScript);
    }

    return () => {
      // Cleanup
      if (typeof window !== "undefined" && window.AOS) {
        window.AOS.refresh();
      }
    };
  }, []);

  const handleFiles = (files: FileList) => {
    if (files.length === 0) return;
    
    setIsUploading(true);
    setUploadProgress(0);
    setShowShareLink(false);

    // Simulate upload progress
    let progress = 0;
    const interval = setInterval(() => {
      progress += Math.random() * 10;
      if (progress > 100) progress = 100;

      setUploadProgress(progress);

      if (progress === 100) {
        clearInterval(interval);
        setTimeout(() => {
          setIsUploading(false);
          // Generate random link
          const randomString = Math.random().toString(36).substring(2, 10);
          setShareLink(`https://cthulhu.sh/${randomString}`);
          setShowShareLink(true);
        }, 500);
      }
    }, 300);
  };

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragActive(false);
    
    const files = e.dataTransfer.files;
    handleFiles(files);
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      handleFiles(e.target.files);
    }
  };

  const handleBrowseClick = () => {
    fileInputRef.current?.click();
  };

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(shareLink);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement("textarea");
      textArea.value = shareLink;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand("copy");
      document.body.removeChild(textArea);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    }
  };

  const handleNewUpload = () => {
    setShowShareLink(false);
    setUploadProgress(0);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <>
      {/* External Scripts */}
      <link href="https://unpkg.com/aos@2.3.1/dist/aos.css" rel="stylesheet" />
      
      <div className="bg-gray-900 text-white min-h-screen">
        {/* Vanta.js Background */}
        <div ref={vantaRef} className="fixed inset-0 z-0"></div>
        
        <div className="relative z-10 container mx-auto px-4 py-12">
          {/* Header */}
          <header className="mb-16 text-center">
            <h1
              className="text-5xl md:text-7xl font-bold mb-4"
              style={{ textShadow: "0 0 10px rgba(139, 92, 246, 0.7)" }}
              data-aos="fade-down"
            >
              CTHULHU
            </h1>
            <p
              className="text-xl md:text-2xl text-purple-300 mb-8"
              data-aos="fade-down"
              data-aos-delay="100"
            >
              Anonymous File Sharing from the Depths
            </p>
          </header>

          {/* Main Content */}
          <main className="max-w-4xl mx-auto">
            {/* Upload Section */}
            <div
              className="bg-gray-800 bg-opacity-70 backdrop-blur-lg rounded-xl p-8 shadow-2xl"
              data-aos="fade-up"
            >
              <div className="text-center mb-8">
                <h2 className="text-3xl font-bold mb-2">Share Files Anonymously</h2>
                <p className="text-gray-300">
                  No accounts. No tracking. Just upload and share.
                </p>
              </div>

              {/* Dropzone */}
              <div
                className={`border-2 border-dashed rounded-lg p-12 mb-6 text-center cursor-pointer transition-all hover:bg-gray-700 hover:bg-opacity-30 ${
                  isDragActive
                    ? "border-purple-500 bg-purple-500 bg-opacity-10"
                    : "border-gray-300 border-opacity-30"
                }`}
                onDragEnter={handleDragEnter}
                onDragLeave={handleDragLeave}
                onDragOver={handleDragOver}
                onDrop={handleDrop}
                onClick={handleBrowseClick}
              >
                <div className="flex flex-col items-center justify-center">
                  <svg
                    className="w-16 h-16 text-purple-400 mb-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                    />
                  </svg>
                  <h3 className="text-xl font-semibold mb-2">Drag & Drop Files Here</h3>
                  <p className="text-gray-400 mb-4">or click to browse</p>
                  <input
                    ref={fileInputRef}
                    type="file"
                    className="hidden"
                    multiple
                    {...({ webkitdirectory: "" } as React.InputHTMLAttributes<HTMLInputElement>)}
                    onChange={handleFileInputChange}
                  />
                  <button
                    className="bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 px-6 rounded-lg transition-colors"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleBrowseClick();
                    }}
                  >
                    Select Files
                  </button>
                </div>
              </div>

              {/* Upload Progress */}
              {isUploading && (
                <div className="mb-6">
                  <div className="flex justify-between mb-2">
                    <span className="text-sm font-medium">Uploading...</span>
                    <span className="text-sm font-medium">{Math.round(uploadProgress)}%</span>
                  </div>
                  <div className="w-full bg-gray-700 rounded-full h-2.5">
                    <div
                      className="bg-purple-600 h-2.5 rounded-full transition-all duration-300"
                      style={{ width: `${uploadProgress}%` }}
                    ></div>
                  </div>
                </div>
              )}

              {/* Share Link */}
              {showShareLink && (
                <div className="mt-8">
                  <div className="bg-gray-700 rounded-lg p-4 mb-4">
                    <div className="flex items-center">
                      <svg
                        className="w-5 h-5 text-purple-400 mr-2"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                        />
                      </svg>
                      <span className="text-sm font-medium">Your shareable link:</span>
                    </div>
                    <div className="mt-2 flex">
                      <input
                        type="text"
                        value={shareLink}
                        readOnly
                        className="flex-1 bg-gray-800 border border-gray-600 rounded-l-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-purple-500"
                      />
                      <button
                        onClick={handleCopyLink}
                        className="bg-purple-600 hover:bg-purple-700 px-4 py-2 rounded-r-lg transition-colors flex items-center"
                      >
                        {copySuccess ? (
                          <>
                            <svg
                              className="w-4 h-4 mr-1"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                              xmlns="http://www.w3.org/2000/svg"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M5 13l4 4L19 7"
                              />
                            </svg>
                            Copied!
                          </>
                        ) : (
                          "Copy"
                        )}
                      </button>
                    </div>
                  </div>
                  <div className="text-center">
                    <button
                      onClick={handleNewUpload}
                      className="text-purple-400 hover:text-purple-300 flex items-center justify-center mx-auto"
                    >
                      <svg
                        className="w-4 h-4 mr-1"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 4v16m8-8H4"
                        />
                      </svg>
                      Upload New Files
                    </button>
                  </div>
                </div>
              )}

              {/* Features */}
              <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6">
                <div
                  className="bg-gray-800 bg-opacity-50 p-6 rounded-lg"
                  data-aos="fade-up"
                  data-aos-delay="100"
                >
                  <div className="flex items-center mb-4">
                    <div className="bg-purple-600 bg-opacity-20 p-3 rounded-full mr-4">
                      <svg
                        className="w-6 h-6 text-purple-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                        />
                      </svg>
                    </div>
                    <h3 className="text-lg font-semibold">Complete Anonymity</h3>
                  </div>
                  <p className="text-gray-400">
                    No accounts required. We don&apos;t track your activity or store
                    personal information.
                  </p>
                </div>
                <div
                  className="bg-gray-800 bg-opacity-50 p-6 rounded-lg"
                  data-aos="fade-up"
                  data-aos-delay="200"
                >
                  <div className="flex items-center mb-4">
                    <div className="bg-purple-600 bg-opacity-20 p-3 rounded-full mr-4">
                      <svg
                        className="w-6 h-6 text-purple-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"
                        />
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M8 5a2 2 0 012-2h4a2 2 0 012 2v2H8V5z"
                        />
                      </svg>
                    </div>
                    <h3 className="text-lg font-semibold">Folder Support</h3>
                  </div>
                  <p className="text-gray-400">
                    Upload entire folders while preserving their structure for easy
                    sharing.
                  </p>
                </div>
                <div
                  className="bg-gray-800 bg-opacity-50 p-6 rounded-lg"
                  data-aos="fade-up"
                  data-aos-delay="300"
                >
                  <div className="flex items-center mb-4">
                    <div className="bg-purple-600 bg-opacity-20 p-3 rounded-full mr-4">
                      <svg
                        className="w-6 h-6 text-purple-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    </div>
                    <h3 className="text-lg font-semibold">Auto Expiration</h3>
                  </div>
                  <p className="text-gray-400">
                    Files are automatically deleted after 30 days or when they
                    become inactive.
                  </p>
                </div>
              </div>
            </div>
          </main>

          {/* Footer */}
          <footer className="mt-20 text-center text-gray-500 text-sm">
            <p>Files are encrypted during transfer and storage. Use responsibly.</p>
            <p className="mt-2">
              © 2023 CTHULHU - All files will eventually return to the void
            </p>
          </footer>
        </div>
      </div>
    </>
  );
}
