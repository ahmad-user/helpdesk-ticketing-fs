<template>
  <div
    :class="[
      'animate-pulse bg-slate-200/80 rounded',
      roundedFull ? 'rounded-full' : '',
      customClass
    ]"
    :style="customStyle"
    aria-hidden="true"
  ></div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  width: {
    type: String,
    default: '100%'
  },
  height: {
    type: String,
    default: '1rem'
  },
  roundedFull: {
    type: Boolean,
    default: false
  },
  customClass: {
    type: String,
    default: ''
  }
})

const customStyle = computed(() => {
  return {
    width: props.width,
    height: props.height
  }
})
</script>
