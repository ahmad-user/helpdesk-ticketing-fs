<template>
  <div class="space-y-6 max-w-4xl mx-auto">
    <!-- Header with Back Button -->
    <div class="flex items-center justify-between pb-4 border-b border-slate-200/90">
      <div class="flex items-center gap-3">
        <router-link
          to="/tickets"
          class="p-2 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 transition-colors"
          title="Back to tickets list"
        >
          <ArrowLeft class="h-5 w-5" />
        </router-link>
        <div>
          <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900">
            Create Support Ticket
          </h1>
          <p class="text-xs sm:text-sm text-slate-500 mt-0.5">
            Submit a new support inquiry or technical issue to the helpdesk team.
          </p>
        </div>
      </div>
    </div>

    <!-- Error Banner -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="transform -translate-y-2 opacity-0"
      enter-to-class="transform translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="errorMessage"
        class="p-4 rounded-xl bg-rose-50 border border-rose-200/90 flex items-start gap-3 text-xs text-rose-700"
      >
        <AlertCircle class="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
        <div class="flex-1 min-w-0">
          <p class="font-semibold text-rose-900">Unable to submit ticket</p>
          <p class="mt-0.5 leading-relaxed">{{ errorMessage }}</p>
        </div>
        <button
          @click="errorMessage = ''"
          type="button"
          class="text-rose-400 hover:text-rose-600 p-0.5"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
    </transition>

    <!-- Create Form Split Layout -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Main Form (2 cols) -->
      <BaseCard no-padding class="lg:col-span-2 p-5 sm:p-6">
        <form @submit.prevent="handleSubmit" class="space-y-5" novalidate>
          <!-- Title -->
          <BaseInput
            id="title"
            label="Ticket Subject"
            placeholder="Brief summary of the issue (e.g. Cannot connect to VPN)"
            v-model="form.title"
            :error="errors.title"
            :disabled="isSubmitting"
            required
            hint="Summarize the issue clearly in a few words (3 - 255 characters)"
          />

          <!-- Category & Priority Row -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <!-- Category -->
            <BaseSelect
              id="category"
              label="Issue Category"
              placeholder="Select Category"
              v-model="form.category"
              :options="categoryOptions"
              :error="errors.category"
              :disabled="isSubmitting"
              required
            />

            <!-- Priority -->
            <div v-if="authStore.isRequester" class="w-full">
              <label class="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-1.5">
                Priority
              </label>
              <div class="p-2.5 rounded-lg border border-slate-200 bg-slate-50 flex items-center justify-between text-xs">
                <div class="flex items-center gap-2">
                  <BaseBadge variant="medium" dot size="xs">Medium</BaseBadge>
                  <span class="text-slate-500 text-[11px]">Standard SLA</span>
                </div>
                <span class="text-[10px] text-slate-400 font-medium">Auto-assigned</span>
              </div>
              <p class="mt-1.5 text-[11px] text-slate-400">
                Support agents review and adjust urgency based on team impact.
              </p>
            </div>

            <!-- Priority Selector for Agent / Admin -->
            <BaseSelect
              v-else
              id="priority"
              label="Priority Level"
              v-model="form.priority"
              :options="priorityOptions"
              :error="errors.priority"
              :disabled="isSubmitting"
              required
            />
          </div>

          <!-- Description -->
          <BaseTextarea
            id="description"
            label="Detailed Description"
            placeholder="Please describe the issue in detail, including steps to reproduce, error codes, and troubleshooting already attempted..."
            v-model="form.description"
            :rows="6"
            :error="errors.description"
            :disabled="isSubmitting"
            required
            hint="Provide sufficient detail to help support engineers resolve your ticket quickly."
          />

          <!-- Form Action Buttons -->
          <div class="pt-4 border-t border-slate-100 flex items-center justify-between">
            <span class="text-xs text-slate-400">
              <span class="text-rose-500 font-bold">*</span> Required fields
            </span>

            <div class="flex items-center gap-2.5">
              <router-link to="/tickets">
                <BaseButton variant="secondary" size="md" :disabled="isSubmitting">
                  Cancel
                </BaseButton>
              </router-link>

              <BaseButton
                type="submit"
                variant="primary"
                size="md"
                :loading="isSubmitting"
                :disabled="isSubmitting"
              >
                Submit Ticket
              </BaseButton>
            </div>
          </div>
        </form>
      </BaseCard>

      <!-- Sidebar Guidance Card (1 col) -->
      <div class="space-y-4">
        <BaseCard title="Submission Guide" subtitle="Helpful tips for rapid resolution">
          <div class="space-y-3.5 text-xs text-slate-600 leading-relaxed">
            <div class="flex items-start gap-2.5">
              <div class="h-5 w-5 rounded-full bg-brand-50 text-brand-600 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
                1
              </div>
              <div>
                <p class="font-semibold text-slate-900">Be Specific</p>
                <p class="text-slate-500 mt-0.5">Include what you were doing when the issue occurred and what was expected.</p>
              </div>
            </div>

            <div class="flex items-start gap-2.5">
              <div class="h-5 w-5 rounded-full bg-brand-50 text-brand-600 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
                2
              </div>
              <div>
                <p class="font-semibold text-slate-900">Error Messages</p>
                <p class="text-slate-500 mt-0.5">Copy and paste the exact wording or error code received.</p>
              </div>
            </div>

            <div class="flex items-start gap-2.5">
              <div class="h-5 w-5 rounded-full bg-brand-50 text-brand-600 flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5">
                3
              </div>
              <div>
                <p class="font-semibold text-slate-900">Check Updates</p>
                <p class="text-slate-500 mt-0.5">You can post follow-up comments and track status in real-time once submitted.</p>
              </div>
            </div>
          </div>
        </BaseCard>

        <!-- Category Reference Card -->
        <BaseCard title="Common Categories" subtitle="Select the most appropriate">
          <div class="space-y-2 text-xs">
            <div class="p-2 rounded-lg bg-slate-50 border border-slate-100">
              <span class="font-semibold text-slate-900">Network:</span>
              <span class="text-slate-500 ml-1">VPN, Wi-Fi, Office LAN connectivity</span>
            </div>
            <div class="p-2 rounded-lg bg-slate-50 border border-slate-100">
              <span class="font-semibold text-slate-900">Hardware:</span>
              <span class="text-slate-500 ml-1">Laptops, monitors, peripherals</span>
            </div>
            <div class="p-2 rounded-lg bg-slate-50 border border-slate-100">
              <span class="font-semibold text-slate-900">Software:</span>
              <span class="text-slate-500 ml-1">Email, SaaS tools, installed applications</span>
            </div>
            <div class="p-2 rounded-lg bg-slate-50 border border-slate-100">
              <span class="font-semibold text-slate-900">Access:</span>
              <span class="text-slate-500 ml-1">Password resets, permissions, 2FA</span>
            </div>
          </div>
        </BaseCard>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ArrowLeft, AlertCircle, X } from 'lucide-vue-next'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextarea from '@/components/common/BaseTextarea.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import { useTicketStore } from '@/stores/ticket'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const ticketStore = useTicketStore()
const authStore = useAuthStore()
const { success } = useNotification()

const isSubmitting = ref(false)
const errorMessage = ref('')

const form = reactive({
  title: '',
  description: '',
  category: '',
  priority: 'medium'
})

const errors = reactive({
  title: '',
  description: '',
  category: '',
  priority: ''
})

const categoryOptions = [
  { label: 'Network & Connectivity', value: 'Network' },
  { label: 'Hardware & Workstation', value: 'Hardware' },
  { label: 'Software & Application', value: 'Software' },
  { label: 'Infrastructure & Cloud', value: 'Infrastructure' },
  { label: 'Access Control & Security', value: 'Access' }
]

const priorityOptions = [
  { label: 'Low - Minor inquiry or low impact', value: 'low' },
  { label: 'Medium - Standard business operations', value: 'medium' },
  { label: 'High - Degraded performance affecting work', value: 'high' },
  { label: 'Urgent - Critical service blocker / outage', value: 'urgent' }
]

function validateForm() {
  let isValid = true
  errors.title = ''
  errors.description = ''
  errors.category = ''
  errors.priority = ''

  if (!form.title.trim()) {
    errors.title = 'Ticket title is required'
    isValid = false
  } else if (form.title.trim().length < 3) {
    errors.title = 'Title must be at least 3 characters'
    isValid = false
  } else if (form.title.trim().length > 255) {
    errors.title = 'Title must not exceed 255 characters'
    isValid = false
  }

  if (!form.category.trim()) {
    errors.category = 'Category is required'
    isValid = false
  }

  const effectivePriority = authStore.isRequester ? 'medium' : form.priority
  if (!effectivePriority) {
    errors.priority = 'Priority is required'
    isValid = false
  }

  if (!form.description.trim()) {
    errors.description = 'Detailed description is required'
    isValid = false
  } else if (form.description.trim().length < 5) {
    errors.description = 'Description must be at least 5 characters'
    isValid = false
  }

  return isValid
}

async function handleSubmit() {
  if (isSubmitting.value) return
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isSubmitting.value = true

  try {
    const payload = {
      title: form.title.trim(),
      description: form.description.trim(),
      category: form.category.trim(),
      priority: authStore.isRequester ? 'medium' : form.priority
    }

    const createdTicket = await ticketStore.createTicket(payload)

    success('Support ticket submitted successfully!')

    if (createdTicket?.id) {
      router.push(`/tickets/${createdTicket.id}`)
    } else {
      router.push('/tickets')
    }
  } catch (err) {
    errorMessage.value = err.message || 'Failed to submit ticket'

    if (err.fieldErrors) {
      if (err.fieldErrors.title) errors.title = err.fieldErrors.title
      if (err.fieldErrors.description) errors.description = err.fieldErrors.description
      if (err.fieldErrors.category) errors.category = err.fieldErrors.category
      if (err.fieldErrors.priority) errors.priority = err.fieldErrors.priority
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>
