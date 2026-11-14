import apiClient from './api'

export const commentService = {
  /**
   * List all comments for a ticket
   * Note: Internal notes are filtered out at the backend level for requesters
   * @param {string} ticketId - UUID of the ticket
   * @returns {Promise<Object>} { success, message, data: CommentResponse[] }
   */
  async listComments(ticketId) {
    const response = await apiClient.get(`/tickets/${ticketId}/comments`)
    return response.data
  },

  /**
   * Add a comment or internal note to a ticket
   * @param {string} ticketId - UUID of the ticket
   * @param {Object} payload - { body: string, is_internal?: boolean }
   * @returns {Promise<Object>} { success, message, data: CommentResponse }
   */
  async createComment(ticketId, payload) {
    const response = await apiClient.post(`/tickets/${ticketId}/comments`, payload)
    return response.data
  }
}

export default commentService
