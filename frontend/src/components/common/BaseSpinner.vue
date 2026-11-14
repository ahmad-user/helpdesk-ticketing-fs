<template>
  <div :class="['flex flex-col items-center justify-center', centered ? 'py-12 w-full' : '']">
    <svg
      :class="['animate-spin text-brand-600', sizeClasses]"
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        class="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        stroke-width="4"
      ></circle>
      <path
        class="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8v8H4z"
      ></path>
    </svg>
    <p v-if="text" class="mt-3 text-sm text-slate-500 font-medium">
      {{ text }}
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  size: {
    type: String,
    default: 'md', // sm, md, lg, xl
    validator: (v) => ['sm', 'md', 'lg', 'xl'].includes(v)
  },
  text: {
    type: String,
    default: ''
  },
  centered: {
    type: Boolean,
    default: true
  }
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'h-4 w-4'
    case 'lg':
      return 'h-8 w-8'
    case 'xl':
      return 'h-12 w-12'
    case 'md':
    default:
      return 'h-6 w-6'
  }
})
</script>
