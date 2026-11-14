import apiClient from './api'

export const authService = {
  /**
   * Authenticate user with email and password
   * @param {Object} credentials - { email, password }
   * @returns {Promise<Object>} { success, message, data: { token, user } }
   */
  async login(credentials) {
    const response = await apiClient.post('/auth/login', credentials)
    return response.data
  },

  /**
   * Register a new user account (public role forced as requester by backend)
   * @param {Object} userData - { name, email, password }
   * @returns {Promise<Object>} { success, message, data: UserResponse }
   */
  async register(userData) {
    const response = await apiClient.post('/auth/register', userData)
    return response.data
  },

  /**
   * Retrieve the current authenticated user's profile
   * @returns {Promise<Object>} { success, message, data: UserResponse }
   */
  async getMe() {
    const response = await apiClient.get('/me')
    return response.data
  }
}

export default authService
