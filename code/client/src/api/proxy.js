
/**
 * Handles the response from the Fetch API.
 * @param {Response} response - The response object from a fetch call.
 * @returns {Promise<any>} - The JSON data from the response.
 * @throws {Error} - Throws an error if the response is not ok.
 */
async function handleResponse(response) {
  const data = await response.json();
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
