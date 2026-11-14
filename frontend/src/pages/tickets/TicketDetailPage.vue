<template>
  <div class="space-y-6">
    <!-- Top Action & Navigation Ribbon -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-3 border-b border-slate-200">
      <div class="flex items-center gap-3">
        <router-link
          to="/tickets"
          class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl border border-slate-200 bg-white text-xs font-semibold text-slate-700 hover:bg-slate-50 hover:text-slate-900 transition shadow-xs"
        >
          <ArrowLeft class="h-3.5 w-3.5" />
          <span>Back to tickets</span>
        </router-link>

        <span v-if="ticket" class="text-xs font-mono font-medium text-slate-500 bg-slate-100 px-2 py-1 rounded-lg flex items-center gap-1 border border-slate-200/80">
          #{{ ticket.id.substring(0, 8) }}
          <button
            type="button"
            @click="copyTicketId(ticket.id)"
            class="text-slate-400 hover:text-slate-700 p-0.5 ml-0.5"
            title="Copy full UUID"
          >
            <Copy class="h-3 w-3" />
          </button>
        </span>
      </div>

      <!-- Action Button: Role-based status transition -->
      <div v-if="ticket" class="flex items-center gap-2 flex-wrap">
        <BaseButton
          v-for="target in availableTransitions"
          :key="target.newStatus"
          :variant="target.buttonVariant"
          size="sm"
          :loading="isUpdatingStatus && pendingStatus === target.newStatus"
          :disabled="isUpdatingStatus"
          @click="promptStatusTransition(target)"
        >
          <component :is="target.icon" class="h-3.5 w-3.5 mr-1.5" />
          {{ target.label }}
        </BaseButton>

        <div v-if="ticket.status === 'closed'" class="px-3 py-1.5 rounded-xl bg-slate-100 border border-slate-200 text-xs text-slate-600 flex items-center gap-1.5 font-semibold">
          <CheckCircle2 class="h-4 w-4 text-slate-500" />
          <span>Ticket Closed</span>
        </div>
      </div>
    </div>

    <!-- Loading Skeleton State -->
    <div v-if="isLoadingTicket" class="space-y-6">
      <BaseCard no-padding class="p-6 space-y-4">
        <BaseSkeleton width="50%" height="2rem" />
        <BaseSkeleton width="30%" height="1rem" />
      </BaseCard>
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        <div class="lg:col-span-8 space-y-4">
          <BaseCard no-padding class="p-6 space-y-3">
            <BaseSkeleton width="100%" height="8rem" />
          </BaseCard>
        </div>
        <div class="lg:col-span-4 space-y-4">
          <BaseCard no-padding class="p-5 space-y-3">
            <BaseSkeleton width="100%" height="12rem" />
          </BaseCard>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <ErrorState
      v-else-if="ticketError"
      title="Unable to load ticket details"
      :description="ticketError"
      @retry="loadTicketData"
    />

    <!-- Main Two-Column Professional Layout -->
    <div v-else-if="ticket" class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- MAIN PRIMARY STREAM (Left: 8 cols) -->
      <div class="lg:col-span-8 space-y-6">
        <!-- 1. Ticket Heading & Original Description Card -->
        <div class="bg-white rounded-2xl border border-slate-200/90 shadow-card p-6 sm:p-7 space-y-5">
          <!-- Badges & Meta -->
          <div class="flex items-center gap-2 flex-wrap">
            <BaseBadge :variant="ticket.status" dot size="sm">
              {{ ticket.status }}
            </BaseBadge>
            <BaseBadge :variant="ticket.priority" size="sm">
              {{ ticket.priority }}
            </BaseBadge>
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700">
              {{ ticket.category }}
            </span>
            <span class="text-xs text-slate-400 ml-auto">
              Opened {{ formatDate(ticket.created_at) }}
            </span>
          </div>

          <!-- Hero Ticket Title -->
          <h1 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight leading-snug">
            {{ ticket.title }}
          </h1>

          <!-- Initial Requester Note Box -->
          <div class="pt-4 border-t border-slate-100">
            <div class="flex items-center gap-3 mb-3">
              <div class="h-8 w-8 rounded-full bg-slate-100 border border-slate-200 flex items-center justify-center text-xs font-bold text-slate-700 shrink-0">
                {{ getUserInitials(ticket.requester?.name || ticket.requester?.email) }}
              </div>
              <div class="min-w-0">
                <p class="text-xs font-bold text-slate-900">
                  {{ ticket.requester?.name || 'Requester' }}
                  <span class="font-normal text-slate-400 text-[11px] ml-1">({{ ticket.requester?.email }})</span>
                </p>
                <p class="text-[11px] text-slate-400">Original issue description</p>
              </div>
            </div>

            <div class="text-sm text-slate-800 leading-relaxed bg-slate-50/70 p-4 rounded-xl border border-slate-200/70 whitespace-pre-line font-normal">
              {{ ticket.description }}
            </div>
          </div>
        </div>

        <!-- 2. Conversation & Activity Stream -->
        <div class="bg-white rounded-2xl border border-slate-200/90 shadow-card overflow-hidden">
          <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <MessageSquare class="h-4 w-4 text-brand-600" />
              <h2 class="text-sm font-bold text-slate-900">
                Activity Stream & Discussion
              </h2>
              <span class="px-2 py-0.5 rounded-full text-[11px] font-bold bg-slate-100 text-slate-700">
                {{ comments.length }}
              </span>
            </div>
            <span class="text-xs text-slate-400 font-medium">
              {{ authStore.isRequester ? 'Public replies only' : 'Public & internal timeline' }}
            </span>
          </div>

          <!-- Spinner for comments -->
          <div v-if="isLoadingComments" class="p-8 text-center">
            <BaseSpinner text="Retrieving discussion history..." size="md" />
          </div>

          <!-- Empty comments -->
          <div v-else-if="comments.length === 0" class="p-8">
            <EmptyState
              title="No messages in this stream yet"
              description="Post a message below to start collaborating on this ticket."
            />
          </div>

          <!-- Visual Timeline Rail -->
          <div v-else class="p-6 space-y-6">
            <div
              v-for="comment in comments"
              :key="comment.id"
              :class="[
                'relative pl-6 sm:pl-8 before:absolute before:left-2 sm:before:left-3 before:top-3 before:bottom-0 before:w-0.5 last:before:hidden',
                comment.is_internal ? 'before:bg-amber-200' : 'before:bg-slate-200'
              ]"
            >
              <!-- Timeline Node Marker -->
              <div
                :class="[
                  'absolute left-0 sm:left-1 top-1 h-5 w-5 rounded-full flex items-center justify-center ring-4 ring-white shrink-0',
                  comment.is_internal ? 'bg-amber-500 text-white' : 'bg-brand-600 text-white'
                ]"
              >
                <Lock v-if="comment.is_internal" class="h-2.5 w-2.5" />
                <MessageSquare v-else class="h-2.5 w-2.5" />
              </div>

              <!-- Message Container -->
              <div
                :class="[
                  'rounded-xl p-4 sm:p-5 border transition-all duration-150',
                  comment.is_internal
                    ? 'bg-amber-50/80 border-amber-200 ring-1 ring-amber-300/40'
                    : 'bg-white border-slate-200/90 shadow-xs'
                ]"
              >
                <!-- Author Header -->
                <div class="flex items-center justify-between mb-3 gap-2 flex-wrap">
                  <div class="flex items-center gap-2.5">
                    <div
                      :class="[
                        'h-7 w-7 rounded-full flex items-center justify-center text-xs font-bold shrink-0',
                        comment.is_internal
                          ? 'bg-amber-200 text-amber-900'
                          : 'bg-brand-100 text-brand-800'
                      ]"
                    >
                      {{ getUserInitials(comment.author?.name || comment.author?.email) }}
                    </div>
                    <div>
                      <div class="flex items-center gap-2">
                        <span class="text-xs font-bold text-slate-900">
                          {{ comment.author?.name || 'Support Staff' }}
                        </span>
                        <BaseBadge :variant="comment.author?.role" size="xs">
                          {{ comment.author?.role }}
                        </BaseBadge>
                      </div>
                      <p class="text-[10px] text-slate-400">
                        {{ comment.author?.email }}
                      </p>
                    </div>
                  </div>

                  <div class="flex items-center gap-2">
                    <span
                      v-if="comment.is_internal"
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-extrabold uppercase tracking-wider bg-amber-100 text-amber-900 border border-amber-300/80"
                    >
                      <Lock class="h-3 w-3" />
                      Staff Private Note
                    </span>
                    <span class="text-[11px] text-slate-400 whitespace-nowrap">
                      {{ formatDate(comment.created_at) }}
                    </span>
                  </div>
                </div>

                <!-- Message Body -->
                <div class="text-xs sm:text-sm text-slate-800 leading-relaxed whitespace-pre-line pl-0.5">
                  {{ comment.body }}
                </div>
              </div>
            </div>
          </div>

          <!-- 3. Modern Reply Composer -->
          <div class="p-6 bg-slate-50/80 border-t border-slate-200/90">
            <form @submit.prevent="submitComment" class="space-y-3.5">
              <!-- Staff Tab Switcher (Public Reply vs Internal Note) -->
              <div v-if="!authStore.isRequester" class="flex items-center justify-between flex-wrap gap-2">
                <div class="inline-flex p-1 rounded-xl bg-slate-200/70 border border-slate-200 text-xs font-semibold">
                  <button
                    type="button"
                    @click="isInternalNote = false"
                    :class="[
                      'px-3 py-1.5 rounded-lg transition-all duration-150 flex items-center gap-1.5 cursor-pointer',
                      !isInternalNote
                        ? 'bg-white text-slate-900 shadow-xs'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                  >
                    <MessageSquare class="h-3.5 w-3.5 text-brand-600" />
                    <span>Public Reply</span>
                  </button>
                  <button
                    type="button"
                    @click="isInternalNote = true"
                    :class="[
                      'px-3 py-1.5 rounded-lg transition-all duration-150 flex items-center gap-1.5 cursor-pointer',
                      isInternalNote
                        ? 'bg-amber-500 text-white shadow-xs font-bold'
                        : 'text-slate-600 hover:text-slate-900'
                    ]"
                  >
                    <Lock class="h-3.5 w-3.5" />
                    <span>Internal Note 🔒</span>
                  </button>
                </div>

                <span v-if="isInternalNote" class="text-[11px] font-semibold text-amber-700 bg-amber-50 px-2.5 py-1 rounded-lg border border-amber-200">
                  Staff Only — Invisible to customer
                </span>
              </div>

              <div v-else class="flex items-center justify-between">
                <h3 class="text-xs font-bold uppercase tracking-wider text-slate-700">
                  Send Response to Support Team
                </h3>
              </div>

              <!-- Composer Textarea -->
              <div
                :class="[
                  'rounded-xl border transition-colors',
                  isInternalNote
                    ? 'border-amber-300 ring-2 ring-amber-100 bg-amber-50/20'
                    : 'border-slate-300 focus-within:border-brand-500 focus-within:ring-2 focus-within:ring-brand-100'
                ]"
              >
                <textarea
                  v-model="newCommentBody"
                  :placeholder="
                    isInternalNote
                      ? 'Write private staff notes, diagnosis logs, or internal comments...'
                      : 'Write your message to the customer or support team...'
                  "
                  rows="4"
                  :disabled="isSubmittingComment || ticket.status === 'closed'"
                  class="block w-full rounded-xl border-0 bg-transparent p-3.5 text-sm text-slate-900 placeholder-slate-400 focus:outline-none resize-y"
                ></textarea>
              </div>

              <!-- Error notification -->
              <p v-if="commentError" class="text-xs text-rose-600 font-medium">
                {{ commentError }}
              </p>

              <!-- Footer Buttons -->
              <div class="flex items-center justify-between pt-1">
                <span v-if="ticket.status === 'closed'" class="text-xs text-slate-400 italic">
                  This ticket is closed. Reopen the ticket to post new messages.
                </span>
                <span v-else class="text-[11px] text-slate-400">
                  Markdown formatting supported.
                </span>

                <BaseButton
                  type="submit"
                  :variant="isInternalNote ? 'warning' : 'primary'"
                  size="sm"
                  :loading="isSubmittingComment"
                  :disabled="isSubmittingComment || !newCommentBody.trim() || ticket.status === 'closed'"
                >
                  <Send class="h-3.5 w-3.5 mr-1.5" />
                  {{ isInternalNote ? 'Save Internal Note' : 'Send Reply' }}
                </BaseButton>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- CONTEXT INSPECTOR SIDEBAR (Right: 4 cols) -->
      <div class="lg:col-span-4 space-y-5">
        <!-- Card 1: Ticket Management Controls (Staff Only) -->
        <BaseCard
          v-if="!authStore.isRequester"
          title="Staff Controls"
          subtitle="Priority & staff assignment"
        >
          <div class="space-y-4">
            <!-- Priority Selector -->
            <div>
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                Priority Level
              </label>
              <div class="flex items-center gap-2">
                <div class="flex-1">
                  <BaseSelect
                    v-model="selectedPriority"
                    :options="priorityOptions"
                    :disabled="isUpdatingPriority || ticket.status === 'closed'"
                    @change="updatePriority"
                  />
                </div>
                <span v-if="isUpdatingPriority" class="text-xs text-brand-600 font-semibold animate-pulse">
                  Saving...
                </span>
              </div>
            </div>

            <!-- Assignee Controls -->
            <div class="pt-3 border-t border-slate-100">
              <label class="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1.5">
                Assigned Support Staff
              </label>

              <!-- Admin Full Staff Select -->
              <div v-if="authStore.isAdmin" class="space-y-2">
                <BaseSelect
                  v-model="selectedAssigneeId"
                  placeholder="Unassigned"
                  :options="assigneeOptions"
                  :disabled="isUpdatingAssignee || ticket.status === 'closed'"
                  @change="updateAssignee"
                />
                <div class="flex items-center justify-between text-xs">
                  <button
                    v-if="ticket.assignee_id !== authStore.user?.id"
                    type="button"
                    @click="assignToSelf"
                    :disabled="isUpdatingAssignee || ticket.status === 'closed'"
                    class="text-brand-600 hover:text-brand-700 font-semibold transition cursor-pointer"
                  >
                    Assign to myself
                  </button>
                  <span v-if="isUpdatingAssignee" class="text-brand-600 font-semibold ml-auto animate-pulse">
                    Updating...
                  </span>
                </div>
              </div>

              <!-- Agent Quick Assign -->
              <div v-else class="space-y-2">
                <div v-if="ticket.assignee" class="p-2.5 rounded-xl bg-slate-50 border border-slate-200 text-xs flex items-center justify-between">
                  <div>
                    <p class="font-bold text-slate-900">{{ ticket.assignee.name }}</p>
                    <p class="text-[11px] text-slate-500">{{ ticket.assignee.email }}</p>
                  </div>
                  <BaseBadge :variant="ticket.assignee.role" size="xs">
                    {{ ticket.assignee.role }}
                  </BaseBadge>
                </div>
                <div v-else class="p-2.5 rounded-xl bg-slate-50 border border-dashed border-slate-300 text-xs text-slate-500 text-center">
                  Ticket is currently unassigned
                </div>

                <BaseButton
                  v-if="ticket.assignee_id !== authStore.user?.id"
                  variant="secondary"
                  full-width
                  size="xs"
                  :loading="isUpdatingAssignee"
                  :disabled="isUpdatingAssignee || ticket.status === 'closed'"
                  @click="assignToSelf"
                >
                  <UserCheck class="h-3.5 w-3.5 mr-1.5 text-brand-600" />
                  Assign to Me
                </BaseButton>
              </div>
            </div>
          </div>
        </BaseCard>

        <!-- Card 2: Requester Profile -->
        <BaseCard title="Requester Profile" subtitle="Contact & account details">
          <div class="space-y-3.5 text-xs">
            <div class="flex items-center gap-3">
              <div class="h-10 w-10 rounded-full bg-slate-100 border border-slate-200 flex items-center justify-center text-xs font-bold text-slate-700 shrink-0">
                {{ getUserInitials(ticket.requester?.name || ticket.requester?.email) }}
              </div>
              <div class="min-w-0 flex-1">
                <p class="font-bold text-slate-900 truncate">
                  {{ ticket.requester?.name || 'Requester' }}
                </p>
                <p class="text-slate-500 truncate mt-0.5">
                  {{ ticket.requester?.email }}
                </p>
              </div>
            </div>

            <div class="pt-2.5 border-t border-slate-100 flex items-center justify-between">
              <span class="text-slate-400">Account Role</span>
              <BaseBadge :variant="ticket.requester?.role || 'requester'" size="xs">
                {{ ticket.requester?.role || 'requester' }}
              </BaseBadge>
            </div>
          </div>
        </BaseCard>

        <!-- Card 3: Ticket Metadata Properties -->
        <BaseCard title="System Properties" subtitle="Tracking metadata">
          <dl class="space-y-2.5 text-xs">
            <div class="flex items-center justify-between pb-2 border-b border-slate-100">
              <dt class="text-slate-500">Category</dt>
              <dd class="font-semibold text-slate-900">{{ ticket.category }}</dd>
            </div>

            <div class="flex items-center justify-between pb-2 border-b border-slate-100">
              <dt class="text-slate-500">Priority</dt>
              <dd>
                <BaseBadge :variant="ticket.priority" size="xs">
                  {{ ticket.priority }}
                </BaseBadge>
              </dd>
            </div>

            <div class="flex items-center justify-between pb-2 border-b border-slate-100">
              <dt class="text-slate-500">Status</dt>
              <dd>
                <BaseBadge :variant="ticket.status" dot size="xs">
                  {{ ticket.status }}
                </BaseBadge>
              </dd>
            </div>

            <div class="flex items-center justify-between pb-2 border-b border-slate-100">
              <dt class="text-slate-500">Created</dt>
              <dd class="text-slate-700 font-medium">{{ formatDate(ticket.created_at) }}</dd>
            </div>

            <div class="flex items-center justify-between">
              <dt class="text-slate-500">Last Updated</dt>
              <dd class="text-slate-700 font-medium">{{ formatDate(ticket.updated_at) }}</dd>
            </div>
          </dl>
        </BaseCard>
      </div>
    </div>

    <!-- Status Transition Confirmation Modal -->
    <ConfirmModal
      :is-open="showConfirmModal"
      :title="confirmModalData.title"
      :message="confirmModalData.message"
      :confirm-text="confirmModalData.confirmText"
      :variant="confirmModalData.variant"
      :loading="isUpdatingStatus"
      @confirm="executeStatusTransition"
      @cancel="showConfirmModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import {
  ArrowLeft,
  Copy,
  CheckCircle2,
  Play,
  RotateCcw,
  XCircle,
  MessageSquare,
  Lock,
  Send,
  UserCheck
} from 'lucide-vue-next'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseSpinner from '@/components/common/BaseSpinner.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ErrorState from '@/components/common/ErrorState.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'
import ticketService from '@/services/ticketService'
import commentService from '@/services/commentService'
import userService from '@/services/userService'
import { formatDate } from '@/utils/formatters'

const route = useRoute()
const authStore = useAuthStore()
const { success, error: notifyError } = useNotification()

// State
const ticket = ref(null)
const comments = ref([])
const isLoadingTicket = ref(true)
const isLoadingComments = ref(false)
const isLoadingAssignees = ref(false)
const ticketError = ref(null)

// Action & Mutation States
const isUpdatingStatus = ref(false)
const isUpdatingPriority = ref(false)
const isUpdatingAssignee = ref(false)
const isSubmittingComment = ref(false)

const selectedPriority = ref('')
const selectedAssigneeId = ref('')
const availableAssignees = ref([])

// Comment form
const newCommentBody = ref('')
const isInternalNote = ref(false)
const commentError = ref('')

// Confirmation modal state
const showConfirmModal = ref(false)
const pendingStatus = ref('')
const confirmModalData = ref({
  title: 'Confirm Transition',
  message: '',
  confirmText: 'Confirm',
  variant: 'primary'
})

const priorityOptions = [
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
  { label: 'Urgent', value: 'urgent' }
]

const assigneeOptions = computed(() => {
  const options = [{ label: 'Unassigned', value: '' }]
  availableAssignees.value.forEach((u) => {
    options.push({
      label: `${u.name} (${u.role})`,
      value: u.id
    })
  })
  return options
})

/**
 * Strict state machine transition rules from backend specifications:
 * - Open -> In Progress: Agent, Admin
 * - Open -> Closed: Requester, Admin
 * - In Progress -> Resolved: Agent, Admin
 * - Resolved -> Closed: Requester, Agent, Admin
 * - Resolved -> Open: Requester (Reopen)
 */
const availableTransitions = computed(() => {
  if (!ticket.value) return []
  const transitions = []
  const currentStatus = ticket.value.status
  const role = authStore.role

  if (currentStatus === 'open') {
    if (role === 'agent' || role === 'admin') {
      transitions.push({
        newStatus: 'in_progress',
        label: 'Start Progress',
        buttonVariant: 'primary',
        icon: Play,
        confirmTitle: 'Start Working on Ticket',
        confirmMsg: 'Mark this ticket as In Progress? This signals to the customer that an engineer is actively diagnosing the issue.'
      })
    }
    if (role === 'requester' || role === 'admin') {
      transitions.push({
        newStatus: 'closed',
        label: 'Close Ticket',
        buttonVariant: 'secondary',
        icon: XCircle,
        confirmTitle: 'Close Ticket',
        confirmMsg: 'Are you sure you want to close this ticket now?'
      })
    }
  } else if (currentStatus === 'in_progress') {
    if (role === 'agent' || role === 'admin') {
      transitions.push({
        newStatus: 'resolved',
        label: 'Mark as Resolved',
        buttonVariant: 'success',
        icon: CheckCircle2,
        confirmTitle: 'Resolve Support Ticket',
        confirmMsg: 'Mark this ticket as Resolved? The requester will be notified to verify and close.'
      })
    }
  } else if (currentStatus === 'resolved') {
    if (role === 'requester') {
      transitions.push({
        newStatus: 'open',
        label: 'Reopen Ticket',
        buttonVariant: 'outline',
        icon: RotateCcw,
        confirmTitle: 'Reopen Ticket',
        confirmMsg: 'If the issue is still persisting, you can reopen this ticket for further investigation.'
      })
    }
    if (role === 'requester' || role === 'agent' || role === 'admin') {
      transitions.push({
        newStatus: 'closed',
        label: 'Confirm Closure',
        buttonVariant: 'secondary',
        icon: CheckCircle2,
        confirmTitle: 'Confirm Closure',
        confirmMsg: 'Are you sure you want to permanently close this resolved ticket?'
      })
    }
  }

  return transitions
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
  success(`Ticket UUID copied to clipboard!`)
}

function promptStatusTransition(target) {
  pendingStatus.value = target.newStatus
  confirmModalData.value = {
    title: target.confirmTitle,
    message: target.confirmMsg,
    confirmText: target.label,
    variant: target.newStatus === 'closed' ? 'warning' : (target.newStatus === 'resolved' ? 'success' : 'primary')
  }
  showConfirmModal.value = true
}

async function executeStatusTransition() {
  if (!pendingStatus.value) return
  isUpdatingStatus.value = true

  try {
    const res = await ticketService.updateTicket(ticket.value.id, {
      status: pendingStatus.value
    })
    ticket.value = res.data
    success(`Ticket status successfully updated to "${res.data.status}"!`)
    showConfirmModal.value = false
  } catch (err) {
    notifyError(err.message || 'Failed to update ticket status')
  } finally {
    isUpdatingStatus.value = false
    pendingStatus.value = ''
  }
}

async function updatePriority() {
  if (!selectedPriority.value || selectedPriority.value === ticket.value.priority) return
  isUpdatingPriority.value = true

  try {
    const res = await ticketService.updateTicket(ticket.value.id, {
      priority: selectedPriority.value
    })
    ticket.value = res.data
    success(`Ticket priority changed to "${res.data.priority}"`)
  } catch (err) {
    notifyError(err.message || 'Failed to update priority')
    selectedPriority.value = ticket.value.priority
  } finally {
    isUpdatingPriority.value = false
  }
}

async function updateAssignee() {
  if (!selectedAssigneeId.value || selectedAssigneeId.value === ticket.value.assignee_id) return
  isUpdatingAssignee.value = true

  try {
    const res = await ticketService.updateTicket(ticket.value.id, {
      assignee_id: selectedAssigneeId.value
    })
    ticket.value = res.data
    success('Ticket successfully assigned!')
  } catch (err) {
    notifyError(err.message || 'Failed to assign ticket')
    selectedAssigneeId.value = ticket.value.assignee_id || ''
  } finally {
    isUpdatingAssignee.value = false
  }
}

async function assignToSelf() {
  if (!authStore.user?.id) return
  isUpdatingAssignee.value = true

  try {
    const res = await ticketService.updateTicket(ticket.value.id, {
      assignee_id: authStore.user.id
    })
    ticket.value = res.data
    selectedAssigneeId.value = authStore.user.id
    success('Ticket assigned to you!')
  } catch (err) {
    notifyError(err.message || 'Failed to assign ticket to yourself')
  } finally {
    isUpdatingAssignee.value = false
  }
}

async function loadAssignees() {
  if (!authStore.isAdmin) return
  isLoadingAssignees.value = true

  try {
    const res = await userService.listUsers({ limit: 100 })
    availableAssignees.value = (res.data || []).filter(
      (u) => u.role === 'agent' || u.role === 'admin'
    )
  } catch (err) {
    console.error('Could not load assignees list:', err)
  } finally {
    isLoadingAssignees.value = false
  }
}

async function loadTicketData() {
  const ticketId = route.params.id
  if (!ticketId) return

  isLoadingTicket.value = true
  ticketError.value = null

  try {
    const res = await ticketService.getTicket(ticketId)
    ticket.value = res.data
    selectedPriority.value = res.data.priority
    selectedAssigneeId.value = res.data.assignee_id || ''

    loadAssignees()
    loadComments(ticketId)
  } catch (err) {
    ticketError.value = err.message || 'Ticket not found or access denied.'
  } finally {
    isLoadingTicket.value = false
  }
}

async function loadComments(ticketId) {
  isLoadingComments.value = true
  try {
    const res = await commentService.listComments(ticketId)
    comments.value = res.data || []
  } catch (err) {
    console.error('Failed to load comments:', err)
  } finally {
    isLoadingComments.value = false
  }
}

async function submitComment() {
  commentError.value = ''
  const trimmed = newCommentBody.value.trim()
  if (!trimmed) {
    commentError.value = 'Comment text cannot be empty'
    return
  }

  isSubmittingComment.value = true
  try {
    const payload = {
      body: trimmed,
      is_internal: !authStore.isRequester && isInternalNote.value
    }

    const res = await commentService.createComment(ticket.value.id, payload)
    comments.value.push(res.data)
    newCommentBody.value = ''
    isInternalNote.value = false
    success(payload.is_internal ? 'Internal staff note added.' : 'Reply posted.')
  } catch (err) {
    commentError.value = err.message || 'Failed to post message'
    notifyError(commentError.value)
  } finally {
    isSubmittingComment.value = false
  }
}

onMounted(() => {
  loadTicketData()
})
</script>
