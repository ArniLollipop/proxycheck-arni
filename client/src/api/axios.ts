import axios from 'axios'

// Configure axios defaults
const api = axios.create({
  baseURL: '/',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add request interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Unauthorized - could redirect to login if needed
      console.error('Unauthorized access')
    }
    return Promise.reject(error)
  }
)

export default api
