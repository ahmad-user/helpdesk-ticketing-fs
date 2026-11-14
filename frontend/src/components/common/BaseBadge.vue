<template>
  <span
    :class="[
      'inline-flex items-center font-medium rounded-full tracking-normal select-none',
      sizeClasses,
      colorConfig.classes
    ]"
  >
    <span
      v-if="dot"
      :class="['mr-1.5 h-1.5 w-1.5 rounded-full shrink-0', colorConfig.dot]"
      aria-hidden="true"
    />
    <slot>{{ displayLabel }}</slot>
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'default', // 'open', 'in_progress', 'resolved', 'closed', 'low', 'medium', 'high', 'urgent', 'requester', 'agent', 'admin', 'default', 'success', 'warning', 'danger', 'info'
  },
  size: {
    type: String,
    default: 'sm', // 'xs', 'sm', 'md'
    validator: (v) => ['xs', 'sm', 'md'].includes(v)
  },
  dot: {
    type: Boolean,
    default: false
  },
  label: {
    type: String,
    default: ''
  }
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'xs':
      return 'text-[11px] px-2 py-0.5'
    case 'md':
      return 'text-xs px-3 py-1'
    case 'sm':
    default:
      return 'text-xs px-2.5 py-0.5'
  }
})

const displayLabel = computed(() => {
  if (props.label) return props.label
  const mapping = {
    open: 'Open',
    in_progress: 'In Progress',
    resolved: 'Resolved',
    closed: 'Closed',
    low: 'Low',
    medium: 'Medium',
    high: 'High',
    urgent: 'Urgent',
    requester: 'Requester',
    agent: 'Agent',
    admin: 'Administrator'
  }
  return mapping[props.variant] || props.variant
})

const colorConfig = computed(() => {
  switch (props.variant) {
    // Ticket Statuses
    case 'open':
      return {
        classes: 'bg-sky-50 text-sky-700 ring-1 ring-inset ring-sky-600/20',
        dot: 'bg-sky-500'
      }
    case 'in_progress':
      return {
        classes: 'bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-600/20',
        dot: 'bg-amber-500 animate-pulse'
      }
    case 'resolved':
      return {
        classes: 'bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/20',
        dot: 'bg-emerald-500'
      }
    case 'closed':
      return {
        classes: 'bg-slate-100 text-slate-600 ring-1 ring-inset ring-slate-500/20',
        dot: 'bg-slate-400'
      }

    // Priorities
    case 'low':
      return {
        classes: 'bg-slate-100 text-slate-600 ring-1 ring-inset ring-slate-500/20',
        dot: 'bg-slate-400'
      }
    case 'medium':
      return {
        classes: 'bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20',
        dot: 'bg-blue-500'
      }
    case 'high':
      return {
        classes: 'bg-orange-50 text-orange-700 ring-1 ring-inset ring-orange-600/20',
        dot: 'bg-orange-500'
      }
    case 'urgent':
      return {
        classes: 'bg-rose-50 text-rose-700 ring-1 ring-inset ring-rose-600/20 font-semibold',
        dot: 'bg-rose-500'
      }

    // Roles
    case 'admin':
      return {
        classes: 'bg-purple-50 text-purple-700 ring-1 ring-inset ring-purple-600/20 font-medium',
        dot: 'bg-purple-500'
      }
    case 'agent':
      return {
        classes: 'bg-indigo-50 text-indigo-700 ring-1 ring-inset ring-indigo-600/20 font-medium',
        dot: 'bg-indigo-500'
      }
    case 'requester':
      return {
        classes: 'bg-teal-50 text-teal-700 ring-1 ring-inset ring-teal-600/20 font-medium',
        dot: 'bg-teal-500'
      }

    // Generic states
    case 'success':
      return {
        classes: 'bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/20',
        dot: 'bg-emerald-500'
      }
    case 'warning':
      return {
        classes: 'bg-amber-50 text-amber-700 ring-1 ring-inset ring-amber-600/20',
        dot: 'bg-amber-500'
      }
    case 'danger':
      return {
        classes: 'bg-rose-50 text-rose-700 ring-1 ring-inset ring-rose-600/20',
        dot: 'bg-rose-500'
      }
    case 'info':
      return {
        classes: 'bg-blue-50 text-blue-700 ring-1 ring-inset ring-blue-600/20',
        dot: 'bg-blue-500'
      }
    case 'default':
    default:
      return {
        classes: 'bg-slate-100 text-slate-700 ring-1 ring-inset ring-slate-500/10',
        dot: 'bg-slate-400'
      }
  }
})
</script>
