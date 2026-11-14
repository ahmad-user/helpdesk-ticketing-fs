<template>
  <header class="h-16 bg-white border-b border-slate-200/90 px-4 sm:px-6 flex items-center justify-between sticky top-0 z-20 shadow-xs">
    <!-- Left side: Mobile Toggle & Page Title / Breadcrumbs -->
    <div class="flex items-center gap-3 min-w-0">
      <button
        @click="$emit('toggle-sidebar')"
        type="button"
        class="lg:hidden p-2 rounded-lg text-slate-500 hover:text-slate-800 hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-brand-500 transition-colors shrink-0"
        aria-label="Toggle navigation menu"
      >
        <Menu class="h-5 w-5" />
      </button>

      <nav class="flex items-center gap-1.5 text-xs text-slate-500 min-w-0" aria-label="Breadcrumb">
        <template v-for="(crumb, idx) in breadcrumbs" :key="idx">
          <span v-if="idx > 0" class="text-slate-300 shrink-0">/</span>
          <router-link
            v-if="crumb.to"
            :to="crumb.to"
            class="text-slate-500 hover:text-slate-900 transition-colors truncate max-w-[140px]"
          >
            {{ crumb.label }}
          </router-link>
          <span
            v-else
            :class="idx === breadcrumbs.length - 1 ? 'font-semibold text-slate-900 truncate max-w-[220px]' : 'text-slate-400 hidden sm:inline truncate max-w-[120px]'"
          >
            {{ crumb.label }}
          </span>
        </template>
      </nav>
    </div>

    <!-- Right side: Profile Menu Dropdown (or Unauthenticated Links) -->
    <div class="flex items-center gap-3 shrink-0">
      <!-- Authenticated Unified Profile Dropdown -->
      <UserProfileDropdown v-if="authStore.isAuthenticated" />

      <!-- Unauthenticated Fallback -->
      <div v-else class="flex items-center gap-2 pl-2">
        <router-link
          to="/login"
          class="text-xs font-medium text-slate-600 hover:text-brand-600 px-3 py-1.5 rounded-lg hover:bg-slate-50 transition-colors"
        >
          Sign In
        </router-link>
        <router-link
          to="/register"
          class="text-xs font-medium text-white bg-brand-600 hover:bg-brand-700 px-3 py-1.5 rounded-lg shadow-xs transition-colors"
        >
          Register
        </router-link>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Menu } from 'lucide-vue-next'
import { useAuthStore } from '@/stores/auth'
import UserProfileDropdown from '@/components/layout/UserProfileDropdown.vue'

defineEmits(['toggle-sidebar'])

const route = useRoute()
const authStore = useAuthStore()

const currentTitle = computed(() => {
  return route.meta?.title || 'Helpdesk System'
})

const breadcrumbs = computed(() => {
  const path = route.path
  if (path === '/dashboard') {
    return [
      { label: 'Workspace' },
      { label: 'Dashboard' }
    ]
  }
  if (path === '/tickets') {
    return [
      { label: 'Workspace' },
      { label: authStore.isRequester ? 'My Tickets' : 'All Tickets' }
    ]
  }
  if (path === '/tickets/create') {
    return [
      { label: 'Workspace' },
      { label: 'Tickets', to: '/tickets' },
      { label: 'Create Ticket' }
    ]
  }
  if (route.name === 'ticket-detail') {
    return [
      { label: 'Workspace' },
      { label: 'Tickets', to: '/tickets' },
      { label: `Ticket #${String(route.params.id || '').substring(0, 8)}` }
    ]
  }
  if (path === '/users') {
    return [
      { label: 'Workspace' },
      { label: 'Administration' },
      { label: 'User Management' }
    ]
  }
  return [
    { label: 'Workspace' },
    { label: currentTitle.value }
  ]
})
</script>
