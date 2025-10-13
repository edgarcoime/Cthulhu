// Gateway API configuration
export const GATEWAY_BASE_URL = "http://localhost:4000";

// API endpoints
export const API_ENDPOINTS = {
  UPLOAD: `${GATEWAY_BASE_URL}/files/upload`,
  FILE_ACCESS: (id: string) => `${GATEWAY_BASE_URL}/files/s/${id}`,
  FILE_DOWNLOAD: (id: string, filename: string) => `${GATEWAY_BASE_URL}/files/s/${id}/d/${filename}`,
} as const;
