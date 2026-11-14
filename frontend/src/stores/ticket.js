import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import ticketService from '@/services/ticketService'

export const useTicketStore = defineStore('ticket', () => {
  const tickets = ref([])
  const currentTicket = ref(null)
  const pagination = ref({
    page: 1,
    limit: 10,
    total: 0,
    totalPages: 0
  })

  const isLoading = ref(false)
  const isCreating = ref(false)
  const error = ref(null)

  const searchKeyword = ref('')
  const sortBy = ref('created_at')
  const sortOrder = ref('desc')

  const filters = reactive({
    status: '',
    priority: '',
    assignee_id: ''
  })

  let inFlightAbortController = null

  /**
   * Fetch paginated tickets according to current filters, search, and sorting
   * @param {number} [targetPage] - Optional page number override
   */
  async function fetchTickets(targetPage) {
    if (targetPage !== undefined) {
      pagination.value.page = targetPage
    }

    isLoading.value = true
    error.value = null

    try {
      const params = {
        page: pagination.value.page,
        limit: pagination.value.limit,
        status: filters.status,
        priority: filters.priority,
        assignee_id: filters.assignee_id,
        search: searchKeyword.value.trim(),
        sort_by: sortBy.value,
        order: sortOrder.value
      }

      const res = await ticketService.listTickets(params)

      // Expected response envelope: { success: true, data: TicketResponse[], pagination: PaginationMetadata }
      tickets.value = res.data || []
      if (res.pagination) {
        pagination.value = {
          page: res.pagination.page,
          limit: res.pagination.limit,
          total: Number(res.pagination.total) || 0,
          totalPages: res.pagination.total_pages || 0
        }
      }
      return tickets.value
    } catch (err) {
      error.value = err.message || 'Failed to fetch tickets from server'
      tickets.value = []
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Fetch single ticket details by UUID
   * @param {string} id
   */
  async function fetchTicketById(id) {
    isLoading.value = true
    error.value = null
    try {
      const res = await ticketService.getTicket(id)
      currentTicket.value = res.data
      return currentTicket.value
    } catch (err) {
      error.value = err.message || 'Failed to load ticket details'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Submit new ticket to POST /api/tickets
   * @param {Object} payload - { title, description, category, priority }
   */
  async function createTicket(payload) {
    isCreating.value = true
    error.value = null

    try {
      const res = await ticketService.createTicket(payload)
      const newTicket = res.data
      // Prepend to ticket list if currently viewing page 1
      if (pagination.value.page === 1) {
        tickets.value.unshift(newTicket)
        pagination.value.total += 1
      }
      return newTicket
    } catch (err) {
      error.value = err.message || 'Failed to create ticket'
      throw err
    } finally {
      isCreating.value = false
    }
  }

  /**
   * Update search keyword and trigger query
   */
  function setSearch(keyword) {
    searchKeyword.value = keyword
    pagination.value.page = 1
    return fetchTickets()
  }

  /**
   * Set filter value and trigger query
   */
  function setFilter(key, value) {
    if (key in filters) {
      filters[key] = value
      pagination.value.page = 1
      return fetchTickets()
    }
  }

  /**
   * Update sorting order and trigger query
   */
  function setSort(field, order = 'desc') {
    sortBy.value = field
    sortOrder.value = order
    pagination.value.page = 1
    return fetchTickets()
  }

  /**
   * Navigate to specific page
   */
  function setPage(newPage) {
    if (newPage < 1 || (pagination.value.totalPages > 0 && newPage > pagination.value.totalPages)) {
      return
    }
    return fetchTickets(newPage)
  }

  /**
   * Change page size limit
   */
  function setLimit(newLimit) {
    pagination.value.limit = newLimit
    pagination.value.page = 1
    return fetchTickets()
  }

  /**
   * Reset all search and filter conditions to defaults
   */
  function resetFilters() {
    searchKeyword.value = ''
    filters.status = ''
    filters.priority = ''
    filters.assignee_id = ''
    sortBy.value = 'created_at'
    sortOrder.value = 'desc'
    pagination.value.page = 1
    return fetchTickets()
  }

  return {
    tickets,
    currentTicket,
    pagination,
    isLoading,
    isCreating,
    error,
    filters,
    searchKeyword,
    sortBy,
    sortOrder,
    fetchTickets,
    fetchTicketById,
    createTicket,
    setSearch,
    setFilter,
    setSort,
    setPage,
    setLimit,
    resetFilters
  }
})

export default useTicketStore
