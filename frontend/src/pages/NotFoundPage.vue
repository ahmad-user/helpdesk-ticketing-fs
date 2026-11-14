<template>
  <div class="min-h-[65vh] flex flex-col items-center justify-center text-center px-4">
    <div class="h-20 w-20 rounded-2xl bg-brand-50 text-brand-600 flex items-center justify-center mb-5 ring-8 ring-brand-50/50">
      <FileQuestion class="h-10 w-10 text-brand-600" />
    </div>

    <span class="text-xs font-bold uppercase tracking-widest text-brand-600 mb-2">404 Error</span>
    <h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900">
      Page Not Found
    </h1>
    <p class="mt-2 text-xs sm:text-sm text-slate-500 max-w-sm leading-relaxed">
      The support resource or URL you requested could not be located or may have been relocated.
    </p>

    <div class="mt-6 flex items-center gap-3">
      <router-link to="/dashboard">
        <BaseButton variant="primary" size="md">
          <template #icon-left>
            <LayoutDashboard class="h-4 w-4 mr-1.5" />
          </template>
          Return to Dashboard
        </BaseButton>
      </router-link>
      <router-link to="/tickets">
        <BaseButton variant="secondary" size="md">
          View Tickets
        </BaseButton>
      </router-link>
    </div>
  </div>
</template>

<script setup>
import { FileQuestion, LayoutDashboard } from 'lucide-vue-next'
import BaseButton from '@/components/common/BaseButton.vue'
</script>
