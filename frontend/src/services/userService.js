import apiClient from './api'

export const userService = {
  /**
   * List users with pagination and search (Admin role required)
   * @param {Object} params - { page, limit, search }
   * @returns {Promise<Object>} { success, message, data: UserResponse[], pagination: PaginationMetadata }
   */
  async listUsers(params = {}) {
    const cleanParams = {}
    Object.entries(params).forEach(([key, value]) => {
      if (value !== '' && value !== null && value !== undefined) {
        cleanParams[key] = value
      }
    })

    const response = await apiClient.get('/users', { params: cleanParams })
    return response.data
  },

  /**
   * Update a user's role (Admin role required)
   * @param {string} userId - Target user UUID
   * @param {string} role - 'requester' | 'agent' | 'admin'
   * @returns {Promise<Object>} { success, message, data: UserResponse }
   */
  async updateUserRole(userId, role) {
    const response = await apiClient.patch(`/users/${userId}/role`, { role })
    return response.data
  },

  async createUser(payload) {
    const response = await apiClient.post('/users', payload)
    return response.data
  },

  async updateUser(userId, payload) {
    const response = await apiClient.patch(`/users/${userId}`, payload)
    return response.data
  },

  async deleteUser(userId) {
    const response = await apiClient.delete(`/users/${userId}`)
    return response.data
  }
}

export default userService
