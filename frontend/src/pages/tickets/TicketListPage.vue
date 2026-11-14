<template>
  <div class="space-y-6">
    <!-- 1. Header with Role-Aware Context & Primary Action -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-slate-200">
      <div>
        <div class="flex items-center gap-2.5">
          <h1 class="text-xl sm:text-2xl font-extrabold tracking-tight text-slate-900">
            {{ authStore.isRequester ? 'My Support Tickets' : 'Enterprise Ticket Directory' }}
          </h1>
          <BaseBadge :variant="authStore.role" size="xs">
            {{ authStore.role }}
          </BaseBadge>
        </div>
        <p class="text-xs sm:text-sm text-slate-500 mt-1">
          <span v-if="authStore.isRequester">
            Track inquiries, monitor troubleshooting progress, and collaborate directly with support engineers.
          </span>
          <span v-else>
            Triage, assign, and manage enterprise support requests across all customer queues.
          </span>
        </p>
      </div>

      <div class="flex items-center gap-2.5 shrink-0">
        <router-link v-if="authStore.isRequester || authStore.isAdmin" to="/tickets/create">
          <BaseButton variant="primary" size="md">
            <template #icon-left>
              <Plus class="h-4 w-4 mr-1.5" />
            </template>
            Create Ticket
          </BaseButton>
        </router-link>
      </div>
    </div>

    <!-- 2. Segmented Status Tabs (Linear / Zendesk Style) -->
    <div class="flex items-center gap-1.5 overflow-x-auto pb-1 border-b border-slate-200/80 scrollbar-none">
      <button
        v-for="tab in statusTabs"
        :key="tab.value"
        type="button"
        @click="selectStatusTab(tab.value)"
        :class="[
          'px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all duration-150 flex items-center gap-2 cursor-pointer',
          selectedStatus === tab.value
            ? 'bg-slate-900 text-white shadow-xs'
            : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
        ]"
      >
        <span>{{ tab.label }}</span>
        <span
          v-if="tab.value === '' && ticketStore.pagination.total"
          :class="[
            'text-[10px] px-1.5 py-0.2 rounded-full font-bold',
            selectedStatus === tab.value ? 'bg-slate-800 text-slate-200' : 'bg-slate-200/80 text-slate-700'
          ]"
        >
          {{ ticketStore.pagination.total }}
        </span>
      </button>
    </div>

    <!-- 3. Unified Filter & Search Toolbar -->
    <div class="bg-white rounded-2xl border border-slate-200/90 shadow-card p-4 space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-12 gap-3 items-center">
        <!-- Search Input -->
        <div :class="!authStore.isRequester ? 'lg:col-span-4' : 'lg:col-span-5'">
          <BaseInput
            v-model="searchInput"
            placeholder="Search tickets by title, keywords..."
            @input="onSearchInput"
          >
            <template #prefix>
              <Search class="h-4 w-4 text-slate-400" />
            </template>
            <template #suffix v-if="searchInput">
              <button
                type="button"
                @click="clearSearch"
                class="text-slate-400 hover:text-slate-600 p-1 rounded"
                aria-label="Clear search"
              >
                <X class="h-4 w-4" />
              </button>
            </template>
          </BaseInput>
        </div>

        <!-- Priority Filter -->
        <div class="lg:col-span-2">
          <BaseSelect
            v-model="selectedPriority"
            placeholder="All Priorities"
            :options="priorityOptions"
            @change="onFilterChange"
          />
        </div>

        <!-- Assignee Filter (Staff Only) -->
        <div v-if="!authStore.isRequester" class="lg:col-span-3">
          <BaseSelect
            v-model="selectedAssignee"
            placeholder="All Assignees"
            :options="assigneeOptions"
            @change="onFilterChange"
          />
        </div>

        <!-- Sort Control -->
        <div :class="!authStore.isRequester ? 'lg:col-span-2' : 'lg:col-span-3'">
          <BaseSelect
            v-model="selectedSort"
            :options="sortOptions"
            @change="onSortChange"
          />
        </div>

        <!-- Reset Button -->
        <div class="lg:col-span-1 flex justify-end">
          <BaseButton
            variant="secondary"
            full-width
            size="md"
            :disabled="!hasActiveFilters"
            @click="handleReset"
            title="Reset active filters"
          >
            <RotateCcw class="h-4 w-4 text-slate-500" />
          </BaseButton>
        </div>
      </div>

      <!-- Active Filter Tags Bar -->
      <div v-if="hasActiveFilters" class="pt-2 border-t border-slate-100 flex items-center gap-2 flex-wrap text-xs text-slate-500">
        <span class="font-medium text-slate-400 text-[11px]">Active Filters:</span>

        <span
          v-if="selectedStatus"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-xs font-medium"
        >
          Status: <strong class="capitalize">{{ selectedStatus.replace('_', ' ') }}</strong>
          <button type="button" @click="selectStatusTab('')" class="hover:text-rose-600 ml-1">✕</button>
        </span>

        <span
          v-if="selectedPriority"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-xs font-medium"
        >
          Priority: <strong class="capitalize">{{ selectedPriority }}</strong>
          <button type="button" @click="clearPriority" class="hover:text-rose-600 ml-1">✕</button>
        </span>

        <span
          v-if="selectedAssignee"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-xs font-medium"
        >
          Assignee filtered
          <button type="button" @click="clearAssignee" class="hover:text-rose-600 ml-1">✕</button>
        </span>

        <span
          v-if="searchInput"
          class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-slate-100 text-slate-700 text-xs font-medium"
        >
          Search: <strong>"{{ searchInput }}"</strong>
          <button type="button" @click="clearSearch" class="hover:text-rose-600 ml-1">✕</button>
        </span>

        <button
          type="button"
          @click="handleReset"
          class="text-xs text-brand-600 hover:text-brand-700 font-semibold underline ml-1 cursor-pointer"
        >
          Clear all
        </button>
      </div>
    </div>

    <!-- 4. Error State -->
    <ErrorState
      v-if="ticketStore.error && !ticketStore.isLoading"
      title="Failed to retrieve tickets"
      :description="ticketStore.error"
      @retry="loadTickets"
    />

    <!-- 5. Loading Skeleton Table -->
    <BaseCard v-else-if="ticketStore.isLoading" no-padding class="overflow-hidden">
      <div class="p-6 space-y-4">
        <div v-for="i in 6" :key="i" class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0">
          <div class="space-y-2 flex-1 max-w-sm">
            <BaseSkeleton width="85%" height="1.125rem" />
            <BaseSkeleton width="40%" height="0.75rem" />
          </div>
          <BaseSkeleton width="5rem" height="1.5rem" rounded-full />
          <BaseSkeleton width="5rem" height="1.5rem" rounded-full />
          <BaseSkeleton width="4rem" height="1.5rem" />
        </div>
      </div>
    </BaseCard>

    <!-- 6. Empty State -->
    <EmptyState
      v-else-if="ticketStore.tickets.length === 0"
      :title="hasActiveFilters ? 'No matching tickets found' : 'No support tickets recorded yet'"
      :description="
        hasActiveFilters
          ? 'No tickets match your active filter criteria. Try adjusting your query or resetting filters.'
          : (authStore.isAgent ? 'There are currently no tickets in the queue.' : 'You have not submitted any support tickets yet. Click below to open your first ticket.')
      "
    >
      <template #action>
        <div class="flex items-center justify-center gap-3">
          <BaseButton
            v-if="hasActiveFilters"
            variant="secondary"
            size="sm"
            @click="handleReset"
          >
            <RotateCcw class="h-3.5 w-3.5 mr-1.5" />
            Reset Filters
          </BaseButton>
          <router-link v-if="authStore.isRequester || authStore.isAdmin" to="/tickets/create">
            <BaseButton variant="primary" size="sm">
              <Plus class="h-3.5 w-3.5 mr-1.5" />
              Create Support Ticket
            </BaseButton>
          </router-link>
        </div>
      </template>
    </EmptyState>

    <!-- 7. Data Table View (Tailored to fit 1366px laptop without overflow!) -->
    <BaseCard v-else no-padding class="overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm divide-y divide-slate-100">
          <thead>
            <tr class="bg-slate-50/80 text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
              <th scope="col" class="w-12 py-3 px-3 text-center">#</th>
              <th scope="col" class="py-3 px-4 min-w-[200px]">Ticket Title & ID</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Category</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Requester</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Assignee</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Priority</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Status</th>
              <th scope="col" class="py-3 px-3 text-right whitespace-nowrap">Created</th>
              <th scope="col" class="py-3 px-4 text-right whitespace-nowrap">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 bg-white">
            <tr
              v-for="(ticket, index) in ticketStore.tickets"
              :key="ticket.id"
              class="hover:bg-slate-50/90 transition-colors group cursor-pointer"
              @click="$router.push(`/tickets/${ticket.id}`)"
            >
              <!-- Index Number -->
              <td class="py-3.5 px-3 text-center text-xs font-medium text-slate-400 tabular-nums">
                {{ (ticketStore.pagination.page - 1) * ticketStore.pagination.limit + index + 1 }}
              </td>

              <!-- Title & Meta (Hero Column) -->
              <td class="py-3.5 px-4 min-w-[200px]">
                <router-link
                  :to="`/tickets/${ticket.id}`"
                  class="font-semibold text-slate-900 group-hover:text-brand-600 block line-clamp-1 transition-colors leading-snug"
                  @click.stop
                >
                  {{ ticket.title }}
                </router-link>
                <div class="flex items-center gap-1.5 mt-0.5">
                  <span class="text-[11px] font-mono text-slate-400">
                    #{{ ticket.id.substring(0, 8) }}
                  </span>
                  <button
                    type="button"
                    @click.stop="copyTicketId(ticket.id)"
                    class="text-slate-300 hover:text-slate-600 text-[10px] p-0.5 rounded transition"
                    title="Copy UUID"
                  >
                    <Copy class="h-3 w-3" />
                  </button>
                </div>
              </td>

              <!-- Category -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <span class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-slate-100 text-slate-700">
                  {{ ticket.category }}
                </span>
              </td>

              <!-- Requester -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <div class="flex items-center gap-2">
                  <div
                    class="h-6 w-6 rounded-full bg-slate-100 border border-slate-200 flex items-center justify-center text-[10px] font-bold text-slate-700 shrink-0"
                  >
                    {{ getUserInitials(ticket.requester?.name || ticket.requester?.email) }}
                  </div>
                  <span class="text-xs font-medium text-slate-800 max-w-[110px] truncate">
                    {{ ticket.requester?.name || 'Requester' }}
                  </span>
                </div>
              </td>

              <!-- Assignee -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <div v-if="ticket.assignee" class="flex items-center gap-2">
                  <div
                    class="h-6 w-6 rounded-full bg-brand-50 border border-brand-200 flex items-center justify-center text-[10px] font-bold text-brand-700 shrink-0"
                  >
                    {{ getUserInitials(ticket.assignee.name || ticket.assignee.email) }}
                  </div>
                  <span class="text-xs font-medium text-slate-800 max-w-[110px] truncate">
                    {{ ticket.assignee.name }}
                  </span>
                </div>
                <span v-else class="text-xs text-slate-400 italic">
                  Unassigned
                </span>
              </td>

              <!-- Priority -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <BaseBadge :variant="ticket.priority" size="xs">
                  {{ ticket.priority }}
                </BaseBadge>
              </td>

              <!-- Status -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <BaseBadge :variant="ticket.status" dot size="xs">
                  {{ ticket.status }}
                </BaseBadge>
              </td>

              <!-- Created Date -->
              <td class="py-3.5 px-3 text-right text-xs text-slate-500 whitespace-nowrap">
                {{ formatDate(ticket.created_at) }}
              </td>

              <!-- Action View Button -->
              <td class="py-3.5 px-4 text-right whitespace-nowrap">
                <router-link :to="`/tickets/${ticket.id}`" @click.stop>
                  <BaseButton variant="secondary" size="xs">
                    View
                  </BaseButton>
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Modern Pagination Footer -->
      <div
        class="px-5 py-3.5 bg-slate-50/80 border-t border-slate-200/90 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600"
      >
        <div class="flex items-center gap-2">
          <span>Show</span>
          <select
            :value="ticketStore.pagination.limit"
            @change="onChangeLimit($event.target.value)"
            class="rounded-lg border border-slate-300 bg-white py-1 px-2 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-1 focus:ring-brand-500 cursor-pointer"
          >
            <option :value="10">10</option>
            <option :value="25">25</option>
            <option :value="50">50</option>
          </select>
          <span>
            per page &bull; Showing
            <strong class="text-slate-900">{{ recordRangeStart }}</strong>
            to
            <strong class="text-slate-900">{{ recordRangeEnd }}</strong>
            of
            <strong class="text-slate-900">{{ ticketStore.pagination.total }}</strong>
            tickets
          </span>
        </div>

        <div class="flex items-center gap-1">
          <button
            type="button"
            :disabled="ticketStore.pagination.page <= 1 || ticketStore.isLoading"
            @click="changePage(ticketStore.pagination.page - 1)"
            class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            Previous
          </button>
          <div class="px-3 py-1.5 text-xs font-bold text-slate-800">
            Page {{ ticketStore.pagination.page }} of {{ Math.max(ticketStore.pagination.totalPages, 1) }}
          </div>
          <button
            type="button"
            :disabled="ticketStore.pagination.page >= ticketStore.pagination.totalPages || ticketStore.isLoading"
            @click="changePage(ticketStore.pagination.page + 1)"
            class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            Next
          </button>
        </div>
      </div>
    </BaseCard>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Plus, Search, X, RotateCcw, Copy } from 'lucide-vue-next'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import { useTicketStore } from '@/stores/ticket'
import { useAuthStore } from '@/stores/auth'
import userService from '@/services/userService'
import { useNotification } from '@/composables/useNotification'
import { formatDate } from '@/utils/formatters'

const ticketStore = useTicketStore()
const authStore = useAuthStore()
const { success } = useNotification()

// Local filter bindings
const searchInput = ref(ticketStore.searchKeyword)
const selectedStatus = ref(ticketStore.filters.status)
const selectedPriority = ref(ticketStore.filters.priority)
const selectedAssignee = ref(ticketStore.filters.assignee_id)
const selectedSort = ref(`${ticketStore.sortBy}:${ticketStore.sortOrder}`)
const availableStaff = ref([])

// Status Segmented Tabs
const statusTabs = [
  { label: 'All Tickets', value: '' },
  { label: 'Open', value: 'open' },
  { label: 'In Progress', value: 'in_progress' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Closed', value: 'closed' }
]

const priorityOptions = [
  { label: 'All Priorities', value: '' },
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Urgent', value: 'urgent' }
]

const assigneeOptions = computed(() => {
  const options = [
    { label: 'All Assignees', value: '' },
    { label: 'Assigned to Me', value: authStore.user?.id || '' }
  ]
  availableStaff.value.forEach((staff) => {
    if (staff.id !== authStore.user?.id) {
      options.push({
        label: `${staff.name} (${staff.role})`,
        value: staff.id
      })
    }
  })
  return options
})

const sortOptions = [
  { label: 'Newest First', value: 'created_at:desc' },
  { label: 'Oldest First', value: 'created_at:asc' },
  { label: 'Recently Updated', value: 'updated_at:desc' },
  { label: 'Priority Order', value: 'priority:desc' },
  { label: 'Status Order', value: 'status:asc' }
]

const hasActiveFilters = computed(() => {
  return (
    searchInput.value.trim() !== '' ||
    selectedStatus.value !== '' ||
    selectedPriority.value !== '' ||
    selectedAssignee.value !== '' ||
    selectedSort.value !== 'created_at:desc'
  )
})

const recordRangeStart = computed(() => {
  if (ticketStore.pagination.total === 0) return 0
  return (ticketStore.pagination.page - 1) * ticketStore.pagination.limit + 1
})

const recordRangeEnd = computed(() => {
  return Math.min(
    ticketStore.pagination.page * ticketStore.pagination.limit,
    ticketStore.pagination.total
  )
})

function getUserInitials(nameOrEmail) {
  if (!nameOrEmail) return '?'
  return nameOrEmail
    .split(' ')
    .filter(Boolean)
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

function copyTicketId(id) {
  navigator.clipboard?.writeText(id)
  success(`Ticket ID #${id.substring(0, 8)} copied to clipboard!`)
}

// Debounced search handling to avoid excessive API requests
let searchDebounceTimeout = null

function onSearchInput(val) {
  clearTimeout(searchDebounceTimeout)
  searchDebounceTimeout = setTimeout(() => {
    const query = typeof val === 'string' ? val : searchInput.value
    ticketStore.setSearch(query)
  }, 350)
}

function clearSearch() {
  searchInput.value = ''
  ticketStore.setSearch('')
}

function selectStatusTab(statusVal) {
  selectedStatus.value = statusVal
  onFilterChange()
}

function clearPriority() {
  selectedPriority.value = ''
  onFilterChange()
}

function clearAssignee() {
  selectedAssignee.value = ''
  onFilterChange()
}

function onFilterChange() {
  ticketStore.filters.status = selectedStatus.value
  ticketStore.filters.priority = selectedPriority.value
  ticketStore.filters.assignee_id = selectedAssignee.value
  ticketStore.pagination.page = 1
  ticketStore.fetchTickets()
}

function onSortChange() {
  const [field, order] = selectedSort.value.split(':')
  ticketStore.setSort(field, order || 'desc')
}

function handleReset() {
  searchInput.value = ''
  selectedStatus.value = ''
  selectedPriority.value = ''
  selectedAssignee.value = ''
  selectedSort.value = 'created_at:desc'
  ticketStore.resetFilters()
}

async function loadStaffAssignees() {
  if (authStore.isRequester) return
  try {
    const res = await userService.listUsers({ limit: 50 })
    availableStaff.value = (res.data || []).filter(
      (u) => u.role === 'agent' || u.role === 'admin'
    )
  } catch (err) {
    console.error('Failed to load assignees for filter:', err)
  }
}

function changePage(page) {
  ticketStore.setPage(page)
}

function onChangeLimit(newLimit) {
  ticketStore.setLimit(Number(newLimit))
}

function loadTickets() {
  ticketStore.fetchTickets()
}

onMounted(() => {
  ticketStore.fetchTickets()
  loadStaffAssignees()
})
</script>
