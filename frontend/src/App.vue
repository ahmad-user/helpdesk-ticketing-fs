<template>
  <component :is="layoutComponent">
    <router-view />
  </component>

  <!-- Global notification toasts -->
  <div
    aria-live="assertive"
    class="fixed inset-0 pointer-events-none flex items-end sm:items-start px-4 py-6 sm:p-6 z-50 justify-end"
  >
    <div class="w-full flex flex-col items-center space-y-2.5 sm:items-end max-w-sm">
      <transition-group
        enter-active-class="transform ease-out duration-300 transition"
        enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-4 scale-95"
        enter-to-class="translate-y-0 opacity-100 sm:translate-x-0 scale-100"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-for="notification in notifications"
          :key="notification.id"
          class="w-full bg-white shadow-dropdown rounded-xl pointer-events-auto border border-slate-200/90 p-3.5 overflow-hidden flex items-start gap-3 backdrop-blur-xs"
        >
          <div
            :class="[
              'h-7 w-7 rounded-lg flex items-center justify-center shrink-0 text-white text-xs shadow-xs',
              notification.type === 'success' ? 'bg-emerald-500' : '',
              notification.type === 'error' ? 'bg-rose-500' : '',
              notification.type === 'info' ? 'bg-brand-600' : ''
            ]"
          >
            <CheckCircle2 v-if="notification.type === 'success'" class="h-4 w-4" />
            <AlertCircle v-else-if="notification.type === 'error'" class="h-4 w-4" />
            <Info v-else class="h-4 w-4" />
          </div>

          <div class="flex-1 min-w-0 pt-0.5">
            <p class="text-xs font-semibold text-slate-900 capitalize">
              {{ notification.type === 'error' ? 'Action Failed' : (notification.type === 'success' ? 'Success' : 'Notice') }}
            </p>
            <p class="text-xs text-slate-600 mt-0.5 leading-relaxed break-words">
              {{ notification.message }}
            </p>
          </div>

          <button
            @click="removeNotification(notification.id)"
            type="button"
            class="text-slate-400 hover:text-slate-600 p-1 rounded-md hover:bg-slate-100 transition-colors focus:outline-none"
            aria-label="Dismiss notification"
          >
            <X class="h-3.5 w-3.5" />
          </button>
        </div>
      </transition-group>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { CheckCircle2, AlertCircle, Info, X } from 'lucide-vue-next'
import AppLayout from '@/layouts/AppLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { useNotification } from '@/composables/useNotification'

const route = useRoute()
const { notifications, removeNotification } = useNotification()

const layoutComponent = computed(() => {
  return route.meta?.layout === 'auth' ? AuthLayout : AppLayout
})
</script>
