<template>
  <div class="relative" ref="dropdownRef">
    <!-- Trigger Capsule Button -->
    <button
      type="button"
      @click="toggleDropdown"
      :aria-expanded="isOpen"
      aria-haspopup="true"
      class="flex items-center gap-2 p-1 sm:px-2.5 sm:py-1.5 rounded-xl border border-slate-200/80 bg-white hover:bg-slate-50 hover:border-slate-300 transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-brand-500/20 group cursor-pointer shadow-xs"
    >
      <!-- Avatar Circle -->
      <div class="h-8 w-8 rounded-full bg-brand-600 text-white flex items-center justify-center text-xs font-bold shrink-0 shadow-xs ring-2 ring-brand-100">
        {{ userInitials }}
      </div>

      <!-- User Name and Role (Desktop/Tablet) -->
      <div class="hidden sm:flex flex-col text-left min-w-0 pr-0.5">
        <span class="text-xs font-semibold text-slate-800 leading-tight truncate max-w-[140px]">
          {{ authStore.userName || 'User' }}
        </span>
        <div class="flex items-center mt-0.5">
          <BaseBadge :variant="authStore.role" size="xs">
            {{ authStore.role }}
          </BaseBadge>
        </div>
      </div>

      <!-- Chevron Down Indicator -->
      <ChevronDown
        :class="[
          'h-3.5 w-3.5 text-slate-400 group-hover:text-slate-600 transition-transform duration-200 shrink-0',
          isOpen ? 'rotate-180 text-brand-600' : ''
        ]"
      />
    </button>

    <!-- Dropdown Popup Menu -->
    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="transform scale-95 opacity-0 -translate-y-1"
      enter-to-class="transform scale-100 opacity-100 translate-y-0"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="transform scale-100 opacity-100 translate-y-0"
      leave-to-class="transform scale-95 opacity-0 -translate-y-1"
    >
      <div
        v-if="isOpen"
        class="absolute right-0 top-full mt-2 w-72 bg-white rounded-2xl shadow-dropdown border border-slate-200/90 py-2 z-50 focus:outline-none"
        role="menu"
        aria-orientation="vertical"
        tabindex="-1"
      >
        <!-- Profile Header Section -->
        <div class="px-4 py-3 border-b border-slate-100 flex items-start gap-3">
          <div class="h-10 w-10 rounded-full bg-brand-600 text-white flex items-center justify-center text-sm font-bold shrink-0 ring-4 ring-brand-50 shadow-xs">
            {{ userInitials }}
          </div>

          <div class="min-w-0 flex-1">
            <p class="text-sm font-bold text-slate-900 truncate">
              {{ authStore.userName || 'Account User' }}
            </p>
            <p class="text-xs text-slate-500 font-mono truncate mt-0.5" :title="authStore.userEmail">
              {{ authStore.userEmail }}
            </p>
            <div class="mt-1.5 flex items-center">
              <BaseBadge :variant="authStore.role" dot size="xs">
                {{ authStore.role }}
              </BaseBadge>
            </div>
          </div>
        </div>

        <!-- Menu Action Items -->
        <div class="py-1">
          <!-- 1. My Profile Action -->
          <button
            type="button"
            @click="openProfileModal"
            class="w-full flex items-center gap-3 px-4 py-2.5 text-xs text-slate-700 hover:bg-slate-50 transition-colors group text-left cursor-pointer"
            role="menuitem"
          >
            <div class="p-1.5 rounded-lg bg-slate-100 text-slate-500 group-hover:bg-brand-50 group-hover:text-brand-600 transition-colors">
              <User class="h-4 w-4" />
            </div>
            <div>
              <p class="font-semibold text-slate-800 group-hover:text-slate-900">
                My Profile
              </p>
              <p class="text-[11px] text-slate-400">
                View account details and permissions
              </p>
            </div>
          </button>
        </div>

        <!-- Divider -->
        <div class="border-t border-slate-100 my-1"></div>

        <!-- 2. Sign Out Action -->
        <div class="py-1">
          <button
            type="button"
            @click="promptSignOut"
            class="w-full flex items-center gap-3 px-4 py-2.5 text-xs text-rose-600 hover:bg-rose-50 transition-colors group text-left cursor-pointer"
            role="menuitem"
          >
            <div class="p-1.5 rounded-lg bg-rose-100/60 text-rose-600 group-hover:bg-rose-100 transition-colors">
              <LogOut class="h-4 w-4" />
            </div>
            <div>
              <p class="font-semibold text-rose-700">
                Sign Out
              </p>
              <p class="text-[11px] text-rose-400">
                End current active session
              </p>
            </div>
          </button>
        </div>
      </div>
    </transition>

    <!-- My Profile Modal (Read-Only) -->
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
          v-if="showProfileModal"
          class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4 sm:p-6"
          role="dialog"
          aria-modal="true"
          @keydown.esc="showProfileModal = false"
        >
          <!-- Backdrop -->
          <div
            class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
            @click="showProfileModal = false"
          ></div>

          <!-- Modal Window -->
          <div class="relative w-full max-w-md bg-white rounded-2xl shadow-dropdown border border-slate-200/90 p-6 z-10 space-y-5">
            <!-- Modal Header -->
            <div class="flex items-center justify-between pb-3 border-b border-slate-100">
              <div class="flex items-center gap-2.5">
                <div class="h-9 w-9 rounded-xl bg-brand-50 text-brand-600 flex items-center justify-center">
                  <User class="h-5 w-5" />
                </div>
                <div>
                  <h3 class="text-base font-bold text-slate-900">
                    Account Profile
                  </h3>
                  <p class="text-xs text-slate-500">
                    Your authenticated credentials & organization role
                  </p>
                </div>
              </div>

              <button
                type="button"
                @click="showProfileModal = false"
                class="p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
                aria-label="Close modal"
              >
                <X class="h-5 w-5" />
              </button>
            </div>

            <!-- Profile Summary Card -->
            <div class="flex items-center gap-4 p-4 rounded-xl bg-slate-50 border border-slate-100">
              <div class="h-14 w-14 rounded-full bg-brand-600 text-white flex items-center justify-center text-lg font-bold shadow-xs shrink-0 ring-4 ring-white">
                {{ userInitials }}
              </div>
              <div class="min-w-0 flex-1">
                <h4 class="text-sm font-bold text-slate-900 truncate">
                  {{ authStore.userName || 'Account User' }}
                </h4>
                <p class="text-xs text-slate-500 truncate mt-0.5">
                  {{ authStore.userEmail }}
                </p>
                <div class="mt-2 flex items-center gap-2">
                  <BaseBadge :variant="authStore.role" dot size="xs">
                    {{ authStore.role }}
                  </BaseBadge>
                  <span class="text-[11px] text-slate-400">Authenticated</span>
                </div>
              </div>
            </div>

            <!-- Profile Attributes Table -->
            <div class="space-y-3 text-xs">
              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Account ID</span>
                <div class="flex items-center gap-1.5 font-mono text-slate-700">
                  <span>#{{ authStore.user?.id ? authStore.user.id.substring(0, 12) + '...' : '-' }}</span>
                  <button
                    v-if="authStore.user?.id"
                    type="button"
                    @click="copyId(authStore.user.id)"
                    class="text-slate-400 hover:text-slate-600 p-0.5"
                    title="Copy UUID"
                  >
                    <Copy class="h-3 w-3" />
                  </button>
                </div>
              </div>

              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Access Role</span>
                <span class="font-semibold text-slate-900 capitalize">{{ authStore.role }}</span>
              </div>

              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Email Address</span>
                <span class="font-medium text-slate-800">{{ authStore.userEmail }}</span>
              </div>

              <div class="flex items-center justify-between py-2 border-b border-slate-100">
                <span class="text-slate-500">Registered Since</span>
                <span class="text-slate-700 font-medium">{{ formatDate(authStore.user?.created_at) }}</span>
              </div>
            </div>

            <!-- Security Notice -->
            <div class="p-3 rounded-xl bg-slate-50 border border-slate-200/80 flex items-start gap-2.5 text-[11px] text-slate-500 leading-relaxed">
              <ShieldCheck class="h-4 w-4 text-emerald-500 shrink-0 mt-0.5" />
              <span>
                Account credentials and permissions are synchronized with the enterprise authentication server.
              </span>
            </div>

            <!-- Modal Action Footer -->
            <div class="pt-2 flex justify-end">
              <BaseButton
                variant="secondary"
                size="sm"
                @click="showProfileModal = false"
              >
                Close
              </BaseButton>
            </div>
          </div>
        </div>
      </transition>
    </Teleport>

    <!-- Sign Out Confirmation Modal -->
    <ConfirmModal
      :is-open="showSignOutModal"
      title="Sign Out"
      message="Are you sure you want to end your current session? You will need to log in again to access the dashboard."
      confirm-text="Sign Out"
      cancel-text="Cancel"
      variant="danger"
      @confirm="executeSignOut"
      @cancel="showSignOutModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  ChevronDown,
  User,
  LogOut,
  X,
  Copy,
  ShieldCheck
} from 'lucide-vue-next'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import ConfirmModal from '@/components/common/ConfirmModal.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'
import { formatDate } from '@/utils/formatters'

const authStore = useAuthStore()
const { success } = useNotification()

const isOpen = ref(false)
const showProfileModal = ref(false)
const showSignOutModal = ref(false)
const dropdownRef = ref(null)

const userInitials = computed(() => {
  if (!authStore.userName) return 'U'
  return authStore.userName
    .split(' ')
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
})

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function openProfileModal() {
  isOpen.value = false
  showProfileModal.value = true
}

function promptSignOut() {
  isOpen.value = false
  showSignOutModal.value = true
}

function executeSignOut() {
  showSignOutModal.value = false
  authStore.logout()
}

function copyId(id) {
  navigator.clipboard?.writeText(id)
  success('User ID copied to clipboard!')
}

// Click outside handler
function handleClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isOpen.value = false
  }
}

// Keydown Escape handler
function handleKeydown(event) {
  if (event.key === 'Escape') {
    isOpen.value = false
    showProfileModal.value = false
    showSignOutModal.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})
</script>
