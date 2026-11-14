<template>
  <div class="text-center py-10 px-6 rounded-xl border border-rose-200 bg-rose-50/40">
    <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-xl bg-rose-100 text-rose-600 mb-3.5 ring-8 ring-rose-50/60">
      <slot name="icon">
        <AlertOctagon class="h-6 w-6" />
      </slot>
    </div>

    <h3 class="text-sm font-semibold text-slate-900">
      {{ title }}
    </h3>
    <p v-if="description" class="mt-1 text-xs text-slate-600 max-w-sm mx-auto leading-relaxed">
      {{ description }}
    </p>

    <div v-if="$slots.action || showRetry" class="mt-5 flex items-center justify-center gap-2.5">
      <slot name="action">
        <BaseButton
          v-if="showRetry"
          type="button"
          variant="danger"
          size="sm"
          @click="$emit('retry')"
        >
          <RotateCcw class="h-3.5 w-3.5 mr-1.5" />
          Try Again
        </BaseButton>
      </slot>
    </div>
  </div>
</template>

<script setup>
import { AlertOctagon, RotateCcw } from 'lucide-vue-next'
import BaseButton from '@/components/common/BaseButton.vue'

defineProps({
  title: {
    type: String,
    default: 'Failed to load content'
  },
  description: {
    type: String,
    default: 'An unexpected error occurred while fetching information. Please try again.'
  },
  showRetry: {
    type: Boolean,
    default: true
  }
})

defineEmits(['retry'])
</script>
