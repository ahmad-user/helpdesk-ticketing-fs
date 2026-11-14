<template>
  <div class="w-full">
    <div v-if="label || $slots.labelRight" class="flex items-center justify-between mb-1.5">
      <label
        v-if="label"
        :for="inputId"
        class="block text-xs font-semibold uppercase tracking-wider text-slate-700"
      >
        {{ label }}
        <span v-if="required" class="text-rose-500 font-bold ml-0.5" aria-hidden="true">*</span>
      </label>
      <slot name="labelRight" />
    </div>

    <div class="relative rounded-lg shadow-xs">
      <div
        v-if="$slots.prefix"
        class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-400"
      >
        <slot name="prefix" />
      </div>

      <input
        :id="inputId"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :required="required"
        @input="handleInput"
        @blur="$emit('blur', $event)"
        :class="[
          'block w-full rounded-lg border text-sm py-2 px-3 transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed',
          $slots.prefix ? 'pl-9' : '',
          $slots.suffix ? 'pr-9' : '',
          error
            ? 'border-rose-300 text-rose-900 placeholder-rose-300 focus:border-rose-500 focus:ring-rose-200'
            : 'border-slate-300/90 text-slate-900 placeholder-slate-400 hover:border-slate-400 focus:border-brand-500 focus:ring-brand-100'
        ]"
      />

      <div
        v-if="$slots.suffix"
        class="absolute inset-y-0 right-0 pr-3 flex items-center"
      >
        <slot name="suffix" />
      </div>
    </div>

    <p v-if="error" class="mt-1.5 text-xs text-rose-600 flex items-center gap-1 font-medium">
      <svg class="h-3.5 w-3.5 shrink-0" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
      </svg>
      <span>{{ error }}</span>
    </p>
    <p v-else-if="hint" class="mt-1.5 text-xs text-slate-500">
      {{ hint }}
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  id: {
    type: String,
    default: ''
  },
  type: {
    type: String,
    default: 'text'
  },
  placeholder: {
    type: String,
    default: ''
  },
  disabled: {
    type: Boolean,
    default: false
  },
  required: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  },
  hint: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'input', 'blur'])

function handleInput(event) {
  const value = event.target.value
  emit('update:modelValue', value)
  emit('input', value)
}

const inputId = computed(() => props.id || `input-${Math.random().toString(36).substring(2, 9)}`)
</script>
