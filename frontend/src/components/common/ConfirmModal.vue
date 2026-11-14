<template>
  <Teleport to="body">
    <transition
      enter-active-class="ease-out duration-200 transition"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="ease-in duration-150 transition"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isOpen"
        class="fixed inset-0 z-50 overflow-y-auto flex items-center justify-center p-4 sm:p-6"
        role="dialog"
        aria-modal="true"
        @keydown.esc="!loading && $emit('cancel')"
      >
        <!-- Backdrop -->
        <div
          class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
          @click="!loading && $emit('cancel')"
        ></div>

        <!-- Modal Box -->
        <div class="relative w-full max-w-md bg-white rounded-2xl shadow-dropdown border border-slate-200/90 p-6 z-10 space-y-4">
          <div class="flex items-start gap-3.5">
            <div
              :class="[
                'h-10 w-10 rounded-xl flex items-center justify-center shrink-0 ring-4',
                variant === 'danger' ? 'bg-rose-50 text-rose-600 ring-rose-50' : '',
                variant === 'warning' ? 'bg-amber-50 text-amber-600 ring-amber-50' : '',
                variant === 'success' ? 'bg-emerald-50 text-emerald-600 ring-emerald-50' : '',
                variant === 'primary' ? 'bg-brand-50 text-brand-600 ring-brand-50' : ''
              ]"
            >
              <slot name="icon">
                <AlertTriangle v-if="variant === 'warning' || variant === 'danger'" class="h-5 w-5" />
                <CheckCircle2 v-else-if="variant === 'success'" class="h-5 w-5" />
                <Info v-else class="h-5 w-5" />
              </slot>
            </div>

            <div class="flex-1 min-w-0">
              <h3 class="text-base font-semibold text-slate-900">
                {{ title }}
              </h3>
              <p class="mt-1 text-xs text-slate-500 leading-relaxed">
                {{ message }}
              </p>
              <div v-if="$slots.default" class="mt-3">
                <slot />
              </div>
            </div>
          </div>

          <div class="pt-3 border-t border-slate-100 flex items-center justify-end gap-2.5">
            <BaseButton
              variant="secondary"
              size="sm"
              :disabled="loading"
              @click="$emit('cancel')"
            >
              {{ cancelText }}
            </BaseButton>

            <BaseButton
              :variant="confirmButtonVariant"
              size="sm"
              :loading="loading"
              :disabled="loading"
              @click="$emit('confirm')"
            >
              {{ confirmText }}
            </BaseButton>
          </div>
        </div>
      </div>
    </transition>
  </Teleport>
</template>

<script setup>
import { computed } from 'vue'
import { AlertTriangle, Info, CheckCircle2 } from 'lucide-vue-next'
import BaseButton from '@/components/common/BaseButton.vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Confirm Action'
  },
  message: {
    type: String,
    default: 'Are you sure you want to proceed?'
  },
  confirmText: {
    type: String,
    default: 'Confirm'
  },
  cancelText: {
    type: String,
    default: 'Cancel'
  },
  variant: {
    type: String,
    default: 'primary', // 'primary', 'warning', 'danger', 'success'
    validator: (v) => ['primary', 'warning', 'danger', 'success'].includes(v)
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['confirm', 'cancel'])

const confirmButtonVariant = computed(() => {
  if (props.variant === 'danger') return 'danger'
  if (props.variant === 'warning') return 'warning'
  if (props.variant === 'success') return 'success'
  return 'primary'
})
</script>
