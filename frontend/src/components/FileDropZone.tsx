'use client';

import React, { useCallback, useState } from 'react';

interface FileDropZoneProps {
  onFileSelect: (file: File) => void;
  acceptedFormats?: string[];
  maxSizeMB?: number;
}

export function FileDropZone({
  onFileSelect,
  acceptedFormats = ['pdf', 'csv', 'xlsx'],
  maxSizeMB = 50,
}: FileDropZoneProps) {
  const [isDragActive, setIsDragActive] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDrag = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setIsDragActive(true);
    } else if (e.type === 'dragleave') {
      setIsDragActive(false);
    }
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent<HTMLDivElement>) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragActive(false);

      const files = e.dataTransfer?.files;
      if (files && files.length > 0) {
        const file = files[0];
        const fileExtension = file.name.split('.').pop()?.toLowerCase();

        if (!fileExtension || !acceptedFormats.includes(fileExtension)) {
          setError(`Invalid file format. Accepted formats: ${acceptedFormats.join(', ')}`);
          return;
        }

        const fileSizeMB = file.size / (1024 * 1024);
        if (fileSizeMB > maxSizeMB) {
          setError(`File size exceeds ${maxSizeMB}MB limit`);
          return;
        }

        setError(null);
        onFileSelect(file);
      }
    },
    [acceptedFormats, maxSizeMB, onFileSelect]
  );

  const handleFileInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = e.target.files;
      if (files && files.length > 0) {
        const file = files[0];
        const fileExtension = file.name.split('.').pop()?.toLowerCase();

        if (!fileExtension || !acceptedFormats.includes(fileExtension)) {
          setError(`Invalid file format. Accepted formats: ${acceptedFormats.join(', ')}`);
          return;
        }

        const fileSizeMB = file.size / (1024 * 1024);
        if (fileSizeMB > maxSizeMB) {
          setError(`File size exceeds ${maxSizeMB}MB limit`);
          return;
        }

        setError(null);
        onFileSelect(file);
      }
    },
    [acceptedFormats, maxSizeMB, onFileSelect]
  );

  return (
    <div
      onDragEnter={handleDrag}
      onDragLeave={handleDrag}
      onDragOver={handleDrag}
      onDrop={handleDrop}
      className={`relative flex flex-col items-center justify-center w-full h-64 border-2 border-dashed rounded-lg cursor-pointer transition-colors ${
        isDragActive
          ? 'border-blue-500 bg-blue-50'
          : 'border-gray-300 bg-gray-50 hover:bg-gray-100'
      }`}
    >
      <div className="flex flex-col items-center justify-center pt-5 pb-6">
        <svg
          className="w-10 h-10 text-gray-400 mb-2"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 4v16m8-8H4"
          />
        </svg>
        <p className="mb-2 text-sm text-gray-500">
          <span className="font-semibold">Click to upload</span> or drag and drop
        </p>
        <p className="text-xs text-gray-500">
          {acceptedFormats.join(', ').toUpperCase()} (Max {maxSizeMB}MB)
        </p>
      </div>

      <input
        type="file"
        className="hidden"
        accept={acceptedFormats.map((fmt) => `.${fmt}`).join(',')}
        onChange={handleFileInputChange}
        onClick={(e) => {
          const input = e.target as HTMLInputElement;
          input.value = ''; // Reset to allow re-uploading the same file
        }}
        id="file-input"
      />
      <label htmlFor="file-input" className="absolute inset-0 cursor-pointer" />

      {error && <p className="text-red-500 text-sm mt-2">{error}</p>}
    </div>
  );
}
