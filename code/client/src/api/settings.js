
async function handleResponse(response) {
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || 'An unknown error occurred');
  }
  return data.data;
}

/**
 * Fetches the current application settings.
 * Corresponds to: GET /api/settings
 * @returns {Promise<object>} A promise that resolves to the settings object.
 */
export async function getSettings() {
  const response = await fetch('/api/settings');
  return await handleResponse(response);
}

/**
 * Updates the application settings.
 * Corresponds to: POST /api/settings
 * @param {object} settings - The settings object to save.
 * @returns {Promise<object>} A promise that resolves to the updated settings object.
 */
export async function updateSettings(settings) {
  const response = await fetch('/api/settings', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(settings),
  });
  return handleResponse(response);
}