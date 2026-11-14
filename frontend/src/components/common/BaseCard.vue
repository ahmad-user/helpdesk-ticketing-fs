<template>
  <div
    :class="[
      'bg-white rounded-xl border border-slate-200/90 shadow-card overflow-hidden transition-all duration-200',
      hoverable ? 'hover:shadow-card-hover hover:border-slate-300' : ''
    ]"
  >
    <div
      v-if="$slots.header || title"
      class="px-5 py-4 border-b border-slate-100 flex items-center justify-between gap-4"
    >
      <slot name="header">
        <div class="min-w-0">
          <h3 v-if="title" class="text-sm font-semibold text-slate-900 truncate">
            {{ title }}
          </h3>
          <p v-if="subtitle" class="mt-0.5 text-xs text-slate-500 truncate">
            {{ subtitle }}
          </p>
        </div>
        <div v-if="$slots.headerAction" class="shrink-0 flex items-center gap-2">
          <slot name="headerAction" />
        </div>
      </slot>
    </div>

    <div :class="[noPadding ? '' : 'p-5']">
      <slot />
    </div>

    <div
      v-if="$slots.footer"
      class="px-5 py-3.5 bg-slate-50/80 border-t border-slate-100 flex items-center justify-between"
    >
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: ''
  },
  subtitle: {
    type: String,
    default: ''
  },
  noPadding: {
    type: Boolean,
    default: false
  },
  hoverable: {
    type: Boolean,
    default: false
  }
})
</script>
