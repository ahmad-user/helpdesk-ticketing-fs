<template>
  <div class="space-y-6">
    <!-- 1. Header with Title, Admin Pill, and Primary Action -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-slate-200">
      <div>
        <div class="flex items-center gap-2.5">
          <h1 class="text-xl sm:text-2xl font-extrabold tracking-tight text-slate-900">
            User Directory & Access Control
          </h1>
          <BaseBadge variant="admin" size="xs">
            Administrator Area
          </BaseBadge>
        </div>
        <p class="text-xs sm:text-sm text-slate-500 mt-1">
          Manage system users, review account directories, and configure Role-Based Access Control (RBAC).
        </p>
      </div>

      <div class="flex items-center gap-2.5 shrink-0">
        <BaseButton variant="primary" size="md" @click="openCreateUser">
          <template #icon-left>
            <UserPlus class="h-4 w-4 mr-1.5" />
          </template>
          Add New User
        </BaseButton>
      </div>
    </div>

    <!-- 2. Role Segmented Tabs (Linear / Directory Style) -->
    <div class="flex items-center gap-1.5 overflow-x-auto pb-1 border-b border-slate-200/80 scrollbar-none">
      <button
        v-for="tab in roleTabs"
        :key="tab.value"
        type="button"
        @click="selectedRoleTab = tab.value"
        :class="[
          'px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all duration-150 flex items-center gap-2 cursor-pointer',
          selectedRoleTab === tab.value
            ? 'bg-slate-900 text-white shadow-xs'
            : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
        ]"
      >
        <span>{{ tab.label }}</span>
        <span
          v-if="tab.count !== undefined"
          :class="[
            'text-[10px] px-1.5 py-0.2 rounded-full font-bold',
            selectedRoleTab === tab.value ? 'bg-slate-800 text-slate-200' : 'bg-slate-200/80 text-slate-700'
          ]"
        >
          {{ tab.count }}
        </span>
      </button>
    </div>

    <!-- 3. Search & Toolbar Card -->
    <div class="bg-white rounded-2xl border border-slate-200/90 shadow-card p-4">
      <div class="flex flex-col sm:flex-row items-center justify-between gap-3">
        <div class="w-full sm:max-w-md relative">
          <BaseInput
            v-model="searchInput"
            placeholder="Search accounts by name or email address..."
            @input="onSearchInput"
          >
            <template #prefix>
              <Search class="h-4 w-4 text-slate-400" />
            </template>
            <template #suffix v-if="searchInput">
              <button
                type="button"
                @click="clearSearch"
                class="text-slate-400 hover:text-slate-600 focus:outline-none p-1 rounded"
                aria-label="Clear search"
              >
                <X class="h-4 w-4" />
              </button>
            </template>
          </BaseInput>
        </div>

        <div class="text-xs text-slate-500 font-medium self-end sm:self-center">
          Total: <strong class="text-slate-900">{{ pagination.total }}</strong> registered users
        </div>
      </div>
    </div>

    <!-- 4. Error State -->
    <ErrorState
      v-if="error && !isLoading"
      title="Failed to retrieve users"
      :description="error"
      @retry="loadUsers"
    />

    <!-- 5. Loading Skeleton -->
    <BaseCard v-else-if="isLoading" no-padding class="overflow-hidden">
      <div class="p-6 space-y-4">
        <div v-for="i in 6" :key="i" class="flex items-center justify-between py-2.5 border-b border-slate-100 last:border-0">
          <div class="flex items-center gap-3 flex-1">
            <BaseSkeleton width="2.5rem" height="2.5rem" rounded-full />
            <div class="space-y-1.5 flex-1 max-w-xs">
              <BaseSkeleton width="80%" height="1rem" />
              <BaseSkeleton width="50%" height="0.75rem" />
            </div>
          </div>
          <BaseSkeleton width="6rem" height="1.5rem" rounded-full />
          <BaseSkeleton width="4rem" height="1.5rem" />
        </div>
      </div>
    </BaseCard>

    <!-- 6. Empty State -->
    <EmptyState
      v-else-if="filteredUsers.length === 0"
      title="No registered accounts found"
      :description="
        searchInput || selectedRoleTab
          ? 'No users match your active search or role criteria. Try adjusting your filter.'
          : 'No accounts have been created yet.'
      "
    >
      <template #action>
        <BaseButton
          v-if="searchInput || selectedRoleTab"
          variant="secondary"
          size="sm"
          @click="resetUserFilters"
        >
          Reset Filters
        </BaseButton>
      </template>
    </EmptyState>

    <!-- 7. Modern Users Table (1366px Safe) -->
    <BaseCard v-else no-padding class="overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm divide-y divide-slate-100">
          <thead>
            <tr class="bg-slate-50/80 text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
              <th scope="col" class="w-12 py-3 px-3 text-center">#</th>
              <th scope="col" class="py-3 px-4 min-w-[200px]">User Name & Account</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Email Address</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Current Role</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Status</th>
              <th scope="col" class="py-3 px-3 text-right whitespace-nowrap">Registered</th>
              <th scope="col" class="py-3 px-4 text-right whitespace-nowrap">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 bg-white">
            <tr
              v-for="(user, index) in filteredUsers"
              :key="user.id"
              class="hover:bg-slate-50/90 transition-colors group"
            >
              <!-- Index Number -->
              <td class="py-3.5 px-3 text-center text-xs font-medium text-slate-400 tabular-nums">
                {{ (pagination.page - 1) * pagination.limit + index + 1 }}
              </td>

              <!-- Account Name & Initials -->
              <td class="py-3.5 px-4 min-w-[200px]">
                <div class="flex items-center gap-3">
                  <div
                    :class="[
                      'h-9 w-9 rounded-full flex items-center justify-center text-xs font-bold shrink-0 border',
                      user.role === 'admin'
                        ? 'bg-purple-50 text-purple-700 border-purple-200'
                        : (user.role === 'agent' ? 'bg-indigo-50 text-indigo-700 border-indigo-200' : 'bg-slate-100 text-slate-700 border-slate-200')
                    ]"
                  >
                    {{ getUserInitials(user.name) }}
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-1.5">
                      <span class="font-semibold text-slate-900 truncate block">
                        {{ user.name }}
                      </span>
                      <span
                        v-if="user.id === authStore.user?.id"
                        class="text-[10px] font-bold text-brand-700 bg-brand-50 border border-brand-200/80 px-1.5 py-0.2 rounded"
                      >
                        You
                      </span>
                    </div>
                    <span class="text-[11px] font-mono text-slate-400 block" :title="user.id">
                      #{{ user.id.substring(0, 8) }}
                    </span>
                  </div>
                </div>
              </td>

              <!-- Email -->
              <td class="py-3.5 px-3 text-slate-600 font-mono text-xs whitespace-nowrap">
                <a :href="`mailto:${user.email}`" class="hover:text-brand-600 hover:underline">
                  {{ user.email }}
                </a>
              </td>

              <!-- Role Badge -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <BaseBadge :variant="user.role" dot size="sm">
                  {{ user.role }}
                </BaseBadge>
              </td>

              <!-- Status Badge -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <span class="inline-flex items-center gap-1.5 text-xs font-medium text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200/60">
                  <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                  Active
                </span>
              </td>

              <!-- Registered Date -->
              <td class="py-3.5 px-3 text-right text-xs text-slate-500 whitespace-nowrap">
                {{ formatDate(user.created_at) }}
              </td>

              <!-- Action Menu -->
              <td class="py-3.5 px-4 text-right whitespace-nowrap">
                <div class="flex items-center justify-end gap-1.5">
                  <button
                    type="button"
                    class="p-1.5 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition cursor-pointer"
                    title="Edit User"
                    @click="openEditUser(user)"
                  >
                    <Pencil class="h-4 w-4" />
                  </button>

                  <button
                    type="button"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50 transition cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Delete User"
                    :disabled="user.id === authStore.user?.id"
                    @click="openDeleteUser(user)"
                  >
                    <Trash2 class="h-4 w-4" />
                  </button>
                </div>
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
            :value="pagination.limit"
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
            <strong class="text-slate-900">{{ pagination.total }}</strong>
            users
          </span>
        </div>

        <div class="flex items-center gap-1">
          <button
            type="button"
            :disabled="pagination.page <= 1 || isLoading"
            @click="changePage(pagination.page - 1)"
            class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            Previous
          </button>
          <div class="px-3 py-1.5 text-xs font-bold text-slate-800">
            Page {{ pagination.page }} of {{ Math.max(pagination.totalPages, 1) }}
          </div>
          <button
            type="button"
            :disabled="pagination.page >= pagination.totalPages || isLoading"
            @click="changePage(pagination.page + 1)"
            class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white text-xs font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            Next
          </button>
        </div>
      </div>
    </BaseCard>

    <!-- User Creation / Editing Modal -->
    <Teleport to="body">
      <transition
        enter-active-class="ease-out duration-200 transition"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="ease-in duration-150 transition"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="showUserForm"
          class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-900/60 backdrop-blur-xs p-4 sm:p-6"
          role="dialog"
          aria-modal="true"
          @keydown.esc="!isSavingUser && closeUserForm()"
        >
          <div class="fixed inset-0" @click="!isSavingUser && closeUserForm()"></div>

          <form
            class="relative w-full max-w-lg space-y-4 rounded-2xl bg-white p-6 sm:p-7 shadow-dropdown border border-slate-200/90 z-10"
            @submit.prevent="submitUserForm"
          >
            <!-- Modal Header -->
            <div class="flex items-center justify-between pb-3 border-b border-slate-100">
              <div class="flex items-center gap-2.5">
                <div class="h-9 w-9 rounded-xl bg-brand-50 text-brand-600 flex items-center justify-center">
                  <UserPlus v-if="!editingUser" class="h-5 w-5" />
                  <Pencil v-else class="h-4 w-4" />
                </div>
                <div>
                  <h2 class="text-base font-bold text-slate-900">
                    {{ editingUser ? 'Edit User Account' : 'Add New User' }}
                  </h2>
                  <p class="text-xs text-slate-500">
                    {{ editingUser ? 'Modify profile credentials and access role.' : 'Create credentials and assign an organizational role.' }}
                  </p>
                </div>
              </div>

              <button
                type="button"
                class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
                @click="closeUserForm"
                aria-label="Close modal"
              >
                <X class="h-5 w-5" />
              </button>
            </div>

            <!-- Error banner -->
            <div
              v-if="formError"
              class="p-3 rounded-xl bg-rose-50 border border-rose-200/90 text-xs text-rose-700 flex items-start gap-2"
            >
              <AlertCircle class="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
              <span>{{ formError }}</span>
            </div>

            <!-- Form Inputs -->
            <div class="space-y-3.5">
              <BaseInput
                v-model="userForm.name"
                label="Full Name"
                placeholder="e.g. Alex Morgan"
                required
                :disabled="isSavingUser"
              />

              <BaseInput
                v-model="userForm.email"
                label="Email Address"
                type="email"
                placeholder="name@company.com"
                required
                :disabled="isSavingUser"
              />

              <template v-if="!editingUser">
                <BaseInput
                  v-model="userForm.password"
                  label="Account Password"
                  type="password"
                  placeholder="Minimum 6 characters"
                  required
                  :disabled="isSavingUser"
                  hint="Provide a temporary password for this user's first login."
                />

                <BaseInput
                  v-model="userForm.confirmPassword"
                  label="Confirm Password"
                  type="password"
                  placeholder="Repeat temporary password"
                  required
                  :disabled="isSavingUser"
                />
              </template>

              <BaseSelect
                v-model="userForm.role"
                label="Assigned System Role"
                :options="roleOptions"
                :disabled="isSavingUser"
                required
                hint="Determines administrative, operational, or requester privileges."
              />
            </div>

            <!-- Footer Actions -->
            <div class="flex items-center justify-end gap-2.5 border-t border-slate-100 pt-4">
              <BaseButton
                variant="secondary"
                size="sm"
                type="button"
                :disabled="isSavingUser"
                @click="closeUserForm"
              >
                Cancel
              </BaseButton>

              <BaseButton
                size="sm"
                type="submit"
                variant="primary"
                :loading="isSavingUser"
              >
                {{ editingUser ? 'Save Changes' : 'Create User' }}
              </BaseButton>
            </div>
          </form>
        </div>
      </transition>
    </Teleport>

    <!-- Delete Confirmation Modal -->
    <ConfirmModal
      :is-open="showDeleteModal"
      title="Delete User Account"
      :message="deleteConfirmMessage"
      confirm-text="Delete User"
      variant="danger"
      :loading="isDeletingUser"
      @confirm="executeDeleteUser"
      @cancel="cancelDeleteUser"
    >
      <p v-if="deleteError" class="mt-2 rounded-lg border border-rose-200 bg-rose-50 p-3 text-xs font-medium text-rose-700" role="alert">
        {{ deleteError }}
      </p>
      <p v-else class="mt-2 text-xs font-semibold text-rose-700">
        This action is permanent. Existing tickets or comments will be preserved and may prevent deletion.
      </p>
    </ConfirmModal>

    <!-- Edit Role Confirmation Modal -->
    <ConfirmModal
      :is-open="showEditRoleConfirm"
      title="Confirm Role Change"
      :message="editRoleConfirmMessage"
      confirm-text="Save Changes"
      :variant="editingUser?.role === 'admin' && userForm.role !== 'admin' ? 'danger' : 'primary'"
      :loading="isSavingUser"
      @confirm="saveUser"
      @cancel="showEditRoleConfirm = false"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Search, X, Pencil, Trash2, UserPlus, AlertCircle } from 'lucide-vue-next'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import userService from '@/services/userService'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'
import { formatDate } from '@/utils/formatters'

const authStore = useAuthStore()
const { success } = useNotification()

const roleOptions = [
  { label: 'Requester (Submit & view own tickets)', value: 'requester' },
  { label: 'Agent (Operational support staff)', value: 'agent' },
  { label: 'Administrator (Full platform permissions)', value: 'admin' }
]

const users = ref([])
const isLoading = ref(true)
const error = ref(null)
const searchInput = ref('')
const selectedRoleTab = ref('')

const roleTabs = computed(() => {
  return [
    { label: 'All Users', value: '' },
    { label: 'Administrators', value: 'admin' },
    { label: 'Support Agents', value: 'agent' },
    { label: 'Requesters', value: 'requester' }
  ]
})

const filteredUsers = computed(() => {
  if (!selectedRoleTab.value) return users.value
  return users.value.filter((u) => u.role === selectedRoleTab.value)
})

const pagination = reactive({
  page: 1,
  limit: 10,
  total: 0,
  totalPages: 1
})

const showUserForm = ref(false)
const editingUser = ref(null)
const isSavingUser = ref(false)
const formError = ref('')
const userForm = reactive({ name: '', email: '', password: '', confirmPassword: '', role: 'requester' })

const showDeleteModal = ref(false)
const deleteTarget = ref(null)
const deleteError = ref('')
const isDeletingUser = ref(false)

const showEditRoleConfirm = ref(false)
const editRoleConfirmMessage = computed(() =>
  `Change ${editingUser.value?.name}'s role from "${editingUser.value?.role}" to "${userForm.role}"?`
)
const deleteConfirmMessage = computed(() =>
  deleteTarget.value ? `Permanently delete ${deleteTarget.value.name} (${deleteTarget.value.email})?` : ''
)

const recordRangeStart = computed(() => {
  if (pagination.total === 0) return 0
  return (pagination.page - 1) * pagination.limit + 1
})

const recordRangeEnd = computed(() => {
  return Math.min(pagination.page * pagination.limit, pagination.total)
})

function getUserInitials(name) {
  if (!name) return 'U'
  return name
    .split(' ')
    .filter(Boolean)
    .map((p) => p[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

// Debounced search
let searchTimeout = null
function onSearchInput(val) {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadUsers()
  }, 350)
}

function clearSearch() {
  searchInput.value = ''
  pagination.page = 1
  loadUsers()
}

function resetUserFilters() {
  searchInput.value = ''
  selectedRoleTab.value = ''
  pagination.page = 1
  loadUsers()
}

function changePage(page) {
  if (page < 1 || (pagination.totalPages > 0 && page > pagination.totalPages)) return
  pagination.page = page
  loadUsers()
}

function onChangeLimit(newLimit) {
  pagination.limit = Number(newLimit)
  pagination.page = 1
  loadUsers()
}

async function loadUsers() {
  isLoading.value = true
  error.value = null

  try {
    const res = await userService.listUsers({
      page: pagination.page,
      limit: pagination.limit,
      search: searchInput.value.trim()
    })

    users.value = res.data || []
    if (res.pagination) {
      pagination.total = Number(res.pagination.total) || 0
      pagination.page = res.pagination.page
      pagination.limit = res.pagination.limit
      pagination.totalPages = res.pagination.total_pages || 1
    }
  } catch (err) {
    error.value = err.message || 'Failed to retrieve registered users from backend.'
  } finally {
    isLoading.value = false
  }
}

function openCreateUser() {
  editingUser.value = null
  Object.assign(userForm, { name: '', email: '', password: '', confirmPassword: '', role: 'requester' })
  formError.value = ''
  showUserForm.value = true
}

function openEditUser(user) {
  editingUser.value = user
  Object.assign(userForm, { name: user.name, email: user.email, password: '', confirmPassword: '', role: user.role })
  formError.value = ''
  showUserForm.value = true
}

function closeUserForm() {
  if (isSavingUser.value) return
  showUserForm.value = false
  showEditRoleConfirm.value = false
}

function openDeleteUser(user) {
  deleteTarget.value = user
  deleteError.value = ''
  showDeleteModal.value = true
}

function cancelDeleteUser() {
  showDeleteModal.value = false
  deleteError.value = ''
}

function validUserForm() {
  if (userForm.name.trim().length < 2) return 'Full name must contain at least 2 characters.'
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(userForm.email.trim())) return 'Enter a valid email address.'
  if (!['requester', 'agent', 'admin'].includes(userForm.role)) return 'Select a valid role.'
  if (!editingUser.value && userForm.password.length < 6) return 'Password must contain at least 6 characters.'
  if (!editingUser.value && userForm.password !== userForm.confirmPassword) return 'Passwords do not match.'
  return ''
}

async function submitUserForm() {
  formError.value = validUserForm()
  if (formError.value) return
  if (editingUser.value && editingUser.value.role !== userForm.role) {
    showEditRoleConfirm.value = true
    return
  }
  await saveUser()
}

async function saveUser() {
  isSavingUser.value = true
  formError.value = ''
  try {
    if (editingUser.value) {
      if (editingUser.value.role !== userForm.role) await userService.updateUserRole(editingUser.value.id, userForm.role)
      const res = await userService.updateUser(editingUser.value.id, { name: userForm.name.trim(), email: userForm.email.trim() })
      if (editingUser.value.id === authStore.user?.id) await authStore.fetchProfile()
      success(res.message || 'User updated successfully.')
    } else {
      const res = await userService.createUser({ name: userForm.name.trim(), email: userForm.email.trim(), password: userForm.password, role: userForm.role })
      success(res.message || 'User created successfully.')
    }
    showEditRoleConfirm.value = false
    showUserForm.value = false
    await loadUsers()
  } catch (err) {
    formError.value = err.message || 'Unable to save user.'
    if (editingUser.value) await loadUsers()
  } finally {
    isSavingUser.value = false
  }
}

async function executeDeleteUser() {
  if (!deleteTarget.value || deleteTarget.value.id === authStore.user?.id) return
  isDeletingUser.value = true
  deleteError.value = ''
  try {
    const res = await userService.deleteUser(deleteTarget.value.id)
    success(res.message || 'User deleted successfully.')
    showDeleteModal.value = false
    deleteTarget.value = null
    deleteError.value = ''
    if (users.value.length === 1 && pagination.page > 1) pagination.page--
    await loadUsers()
  } catch (err) {
    const backendMessage = err.raw?.response?.data?.message || err.message
    const isHistoryConflict = err.status === 409 && /ticket|comment|history/i.test(backendMessage || '')
    deleteError.value = isHistoryConflict
      ? 'This account cannot be deleted because it has existing tickets or comments. The history has been preserved.'
      : backendMessage || 'Unable to delete user.'
  } finally {
    isDeletingUser.value = false
  }
}

onMounted(() => {
  loadUsers()
})
</script>
