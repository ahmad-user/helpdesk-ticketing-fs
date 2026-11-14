<template>
  <div class="w-full">
    <div v-if="label || $slots.labelRight" class="flex items-center justify-between mb-1.5">
      <label
        v-if="label"
        :for="selectId"
        class="block text-xs font-semibold uppercase tracking-wider text-slate-700"
      >
        {{ label }}
        <span v-if="required" class="text-rose-500 font-bold ml-0.5" aria-hidden="true">*</span>
      </label>
      <slot name="labelRight" />
    </div>

    <div class="relative rounded-lg shadow-xs">
      <select
        :id="selectId"
        :value="modelValue"
        :disabled="disabled"
        :required="required"
        @change="handleChange"
        @blur="$emit('blur', $event)"
        :class="[
          'block w-full appearance-none rounded-lg border text-sm py-2 pl-3 pr-9 transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-offset-0 disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed bg-white cursor-pointer',
          error
            ? 'border-rose-300 text-rose-900 focus:border-rose-500 focus:ring-rose-200'
            : 'border-slate-300/90 text-slate-900 hover:border-slate-400 focus:border-brand-500 focus:ring-brand-100'
        ]"
      >
        <option v-if="placeholder" value="" disabled :selected="!modelValue">
          {{ placeholder }}
        </option>
        <option
          v-for="opt in formattedOptions"
          :key="opt.value"
          :value="opt.value"
        >
          {{ opt.label }}
        </option>
      </select>

      <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2.5 text-slate-400">
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
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
  options: {
    type: Array,
    default: () => []
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

const emit = defineEmits(['update:modelValue', 'change', 'blur'])

function handleChange(event) {
  const value = event.target.value
  emit('update:modelValue', value)
  emit('change', value)
}

const selectId = computed(() => props.id || `select-${Math.random().toString(36).substring(2, 9)}`)

const formattedOptions = computed(() => {
  return props.options.map((opt) => {
    if (typeof opt === 'object' && opt !== null) {
      return {
        label: opt.label ?? opt.name ?? String(opt.value),
        value: opt.value ?? opt.id
      }
    }
    return {
      label: String(opt),
      value: opt
    }
  })
})
</script>
