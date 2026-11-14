import axios from 'axios'
import router from '@/router'

const baseURL = import.meta.env.VITE_API_BASE_URL || import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

export const apiClient = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  },
  timeout: 15000
})

// Request interceptor: automatically attach JWT Bearer token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token && !config.headers.Authorization) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor: standardize error format and handle central 401 unauthorized
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const url = error.config?.url || ''

    // Central 401 handling (exclude login and register attempts to prevent redirect loops)
    const isAuthRequest = url.includes('/auth/login') || url.includes('/auth/register')
    if (status === 401 && !isAuthRequest) {
      localStorage.removeItem('auth_token')

      const currentPath = router.currentRoute?.value?.path
      if (currentPath && currentPath !== '/login' && currentPath !== '/register') {
        router.push({
          path: '/login',
          query: { redirect: router.currentRoute.value.fullPath }
        })
      }
    }

    // Standardize error message and validation fields
    let message = 'An unexpected error occurred'
    let errors = null
    const fieldErrors = {}

    if (error.response) {
      const data = error.response.data
      message = data?.message || `Request failed with status ${status}`
      errors = data?.errors || null

      if (Array.isArray(errors)) {
        errors.forEach((err) => {
          if (err.field && err.message) {
            fieldErrors[err.field.toLowerCase()] = err.message
          }
        })
      }
    } else if (error.request) {
      message = 'Network error: Unable to connect to the backend server. Please verify the service is running.'
    } else {
      message = error.message || 'Unknown request configuration error'
    }

    const normalizedError = {
      status,
      message,
      errors,
      fieldErrors,
      raw: error
    }

    return Promise.reject(normalizedError)
  }
)

export default apiClient
