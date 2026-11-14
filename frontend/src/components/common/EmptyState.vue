<template>
  <div class="text-center py-12 px-6 rounded-xl border border-dashed border-slate-200 bg-white/60">
    <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-xl bg-slate-100 text-slate-400 mb-3.5 ring-8 ring-slate-50">
      <slot name="icon">
        <Inbox class="h-6 w-6 text-slate-400" />
      </slot>
    </div>

    <h3 class="text-sm font-semibold text-slate-900">
      {{ title }}
    </h3>
    <p v-if="description" class="mt-1 text-xs text-slate-500 max-w-sm mx-auto leading-relaxed">
      {{ description }}
    </p>

    <div v-if="$slots.action" class="mt-5 flex items-center justify-center gap-2.5">
      <slot name="action" />
    </div>
  </div>
</template>

<script setup>
import { Inbox } from 'lucide-vue-next'

defineProps({
  title: {
    type: String,
    default: 'No items found'
  },
  description: {
    type: String,
    default: 'Get started by creating a new entry or adjusting your filters.'
  }
})
</script>
