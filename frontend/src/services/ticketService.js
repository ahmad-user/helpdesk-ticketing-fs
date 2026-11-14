import apiClient from './api'

export const ticketService = {
  /**
   * List tickets with filtering, searching, sorting, and pagination
   * @param {Object} params - Query parameters (page, limit, status, priority, search, sort_by, order, assignee_id)
   * @returns {Promise<Object>} { success, message, data: TicketResponse[], pagination: PaginationMetadata }
   */
  async listTickets(params = {}) {
    const cleanParams = {}
    Object.entries(params).forEach(([key, value]) => {
      if (value !== '' && value !== null && value !== undefined) {
        cleanParams[key] = value
      }
    })

    const response = await apiClient.get('/tickets', { params: cleanParams })
    return response.data
  },

  /**
   * Retrieve a specific ticket by ID
   * @param {string} id - Ticket UUID
   * @returns {Promise<Object>} { success, message, data: TicketResponse }
   */
  async getTicket(id) {
    const response = await apiClient.get(`/tickets/${id}`)
    return response.data
  },

  /**
   * Create a new ticket
   * @param {Object} payload - { title, description, category, priority }
   * @returns {Promise<Object>} { success, message, data: TicketResponse }
   */
  async createTicket(payload) {
    const response = await apiClient.post('/tickets', payload)
    return response.data
  },

  /**
   * Update ticket fields (status, priority, assignee_id)
   * @param {string} id - Ticket UUID
   * @param {Object} payload - Partial update payload { status?, priority?, assignee_id? }
   * @returns {Promise<Object>} { success, message, data: TicketResponse }
   */
  async updateTicket(id, payload) {
    const response = await apiClient.patch(`/tickets/${id}`, payload)
    return response.data
  }
}

export default ticketService
