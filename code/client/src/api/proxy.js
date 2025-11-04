
/**
 * Handles the response from the Fetch API.
 * @param {Response} response - The response object from a fetch call.
 * @returns {Promise<any>} - The JSON data from the response.
 * @throws {Error} - Throws an error if the response is not ok.
 */
async function handleResponse(response) {
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || 'An unknown error occurred');
  }
  return data.data;
}

/**
 * Fetches the list of all proxies.
 * Corresponds to: GET /api/proxy
 * @returns {Promise<Array>} A promise that resolves to an array of proxies.
 */
export async function getProxies() {
  const response = await fetch('/api/proxy');
  return handleResponse(response);
}

/**
 * Creates a new proxy.
 * Corresponds to: POST /api/proxy
 * @param {object} proxyData - The data for the new proxy (e.g., {ip, port, username, password}).
 * @returns {Promise<object>} A promise that resolves to the newly created proxy.
 */
export async function createProxy(proxyData) {
  const response = await fetch('/api/proxy', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(proxyData),
  });
  return handleResponse(response);
}

/**
 * Updates an existing proxy.
 * Corresponds to: PUT /api/proxy/:id
 * @param {string} id - The ID of the proxy to update.
 * @param {object} proxyData - The new data for the proxy.
 * @returns {Promise<object>} A promise that resolves to the updated proxy object.
 */
export async function updateProxy(id, proxyData) {
  const response = await fetch(`/api/proxy/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(proxyData),
  });
  return handleResponse(response);
}

/**
 * Deletes a proxy by its ID.
 * Corresponds to: DELETE /api/proxy/:id
 * @param {string} id - The ID of the proxy to delete.
 * @returns {Promise<object>} A promise that resolves to a confirmation message.
 */
export async function deleteProxy(id) {
  const response = await fetch(`/api/proxy/${id}`, {
    method: 'DELETE',
  });
  return handleResponse(response);
}

/**
 * Verifies a single proxy by its ID.
 * Corresponds to: GET /api/proxy/:id/verify
 * @param {string} id - The ID of the proxy to verify.
 * @returns {Promise<object>} A promise that resolves to the updated proxy data.
 */
export async function verifyProxy(id) {
  const response = await fetch(`/api/proxy/${id}/verify`);
  return handleResponse(response);
}

/**
 * Verifies a batch of proxies by their IDs.
 * Corresponds to: POST /api/proxy/verify-batch
 * @param {Array<string>} ids - An array of proxy IDs to verify.
 * @returns {Promise<object>} A promise that resolves to a confirmation message.
 */
export async function verifyBatch(ids) {
  const response = await fetch(`/api/proxy/verify-batch?ids=${ids.join(',')}`, {
    method: 'POST',
  });
  return handleResponse(response);
}

/**
 * Imports proxies from a text file.
 * Corresponds to: POST /import
 * @param {File} file - The text file containing proxy data.
 * @returns {Promise<object>} A promise that resolves to the import summary.
 */
export async function importProxies(file) {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('api/import', {
    method: 'POST',
    body: formData,
  });

  // Для этого эндпоинта ответ не содержит поля 'data', поэтому обрабатываем его отдельно.
  const result = await response.json();
  if (!response.ok) {
    throw new Error(result.error || 'Failed to import proxies');
  }
  return result;
}

/**
 * Helper function to trigger file download from a fetch response.
 * @param {Response} response - The fetch response object.
 * @param {string} defaultFilename - A default filename if one isn't provided in the response headers.
 */
async function handleFileDownload(response, defaultFilename) {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Failed to download file' }));
    throw new Error(errorData.error);
  }

  const disposition = response.headers.get('Content-Disposition');
  let filename = defaultFilename;
  if (disposition && disposition.includes('attachment')) {
    const filenameMatch = /filename="([^"]+)"/.exec(disposition);
    if (filenameMatch && filenameMatch[1]) {
      filename = filenameMatch[1];
    }
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.style.display = 'none';
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}

/**
 * Exports all proxies to a CSV file.
 * Corresponds to: GET /api/export/all
 * @returns {Promise<void>} A promise that resolves when the download is initiated.
 */
export async function exportAllProxies() {
  const response = await fetch('/api/export/all');
  await handleFileDownload(response, 'proxies.csv');
}

/**
 * Exports selected proxies to a CSV file.
 * Corresponds to: GET /api/export/selected
 * @param {Array<string>} ids - An array of proxy IDs to export.
 * @returns {Promise<void>} A promise that resolves when the download is initiated.
 */
export async function exportSelectedProxies(ids) {
  if (!ids || ids.length === 0) {
    throw new Error('No proxy IDs provided for export.');
  }
  const response = await fetch(`/api/export/selected?ids=${ids.join(',')}`);
  await handleFileDownload(response, 'selected_proxies.csv');
}

export async function getSpeedLogs(params) {
  try {
    const url = new URL('/api/speedLogs', window.location.origin);
    if (params) {
      Object.keys(params).forEach(key => {
        if (params[key] !== null && params[key] !== undefined) {
          url.searchParams.append(key, params[key]);
        }
      });
    }
    const response = await fetch(url);
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch speed logs:', error);
    throw error;
  }
}

export async function getIpLogs(params) {
  try {
    const url = new URL('/api/ipLogs', window.location.origin);
    if (params) {
      Object.keys(params).forEach(key => {
        if (params[key] !== null && params[key] !== undefined) {
          url.searchParams.append(key, params[key]);
        }
      });
    }
    const response = await fetch(url);
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch IP logs:', error);
    throw error;
  }
}

export async function getProxyVisits(params) {
  try {
    const url = new URL('/api/proxyVisits', window.location.origin);
    if (params) {
      Object.keys(params).forEach(key => {
        if (params[key] !== null && params[key] !== undefined) {
          url.searchParams.append(key, params[key]);
        }
      });
    }
    const response = await fetch(url);
    return await response.json();
  } catch (error) {
    console.error('Failed to fetch proxy visit logs:', error);
    throw error;
  }
}