<template>
  <div>
    <!-- Form Heading -->
    <div class="auth-page-heading mb-6">
      <h2 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900">
        Create an account
      </h2>
      <p class="mt-1 text-xs sm:text-sm text-slate-500">
        Register to submit and track support issues with our engineering team.
      </p>
    </div>

    <!-- Server / Error Alert -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="transform -translate-y-2 opacity-0"
      enter-to-class="transform translate-y-0 opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="errorMessage"
        class="mb-5 p-3.5 rounded-xl bg-rose-50 border border-rose-200/90 flex items-start gap-2.5 text-xs text-rose-700"
      >
        <AlertCircle class="h-4 w-4 shrink-0 text-rose-600 mt-0.5" />
        <div class="flex-1 min-w-0">
          <p class="font-semibold text-rose-900">Registration Failed</p>
          <p class="mt-0.5 leading-relaxed">{{ errorMessage }}</p>
        </div>
        <button
          @click="errorMessage = ''"
          type="button"
          class="text-rose-400 hover:text-rose-600 p-0.5 rounded transition-colors"
          aria-label="Dismiss alert"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
    </transition>

    <!-- Register Form -->
    <form @submit.prevent="handleRegister" class="auth-form space-y-4" novalidate>
      <BaseInput
        id="name"
        label="Full Name"
        type="text"
        placeholder="Jane Doe"
        v-model="form.name"
        :error="errors.name"
        :disabled="isSubmitting"
        required
      >
        <template #prefix>
          <User class="h-4 w-4 text-slate-400" />
        </template>
      </BaseInput>

      <BaseInput
        id="email"
        label="Work Email"
        type="email"
        placeholder="jane@company.com"
        v-model="form.email"
        :error="errors.email"
        :disabled="isSubmitting"
        required
      >
        <template #prefix>
          <Mail class="h-4 w-4 text-slate-400" />
        </template>
      </BaseInput>

      <BaseInput
        id="password"
        label="Password"
        :type="showPassword ? 'text' : 'password'"
        placeholder="Minimum 6 characters"
        v-model="form.password"
        :error="errors.password"
        :disabled="isSubmitting"
        required
      >
        <template #prefix>
          <Lock class="h-4 w-4 text-slate-400" />
        </template>
        <template #suffix>
          <button
            type="button"
            @click="showPassword = !showPassword"
            class="text-slate-400 hover:text-slate-600 focus:outline-none p-1 transition-colors"
            :aria-label="showPassword ? 'Hide password' : 'Show password'"
          >
            <EyeOff v-if="showPassword" class="h-4 w-4" />
            <Eye v-else class="h-4 w-4" />
          </button>
        </template>
      </BaseInput>

      <BaseInput
        id="password_confirmation"
        label="Confirm Password"
        :type="showPassword ? 'text' : 'password'"
        placeholder="Repeat your password"
        v-model="form.passwordConfirmation"
        :error="errors.passwordConfirmation"
        :disabled="isSubmitting"
        required
      >
        <template #prefix>
          <Lock class="h-4 w-4 text-slate-400" />
        </template>
      </BaseInput>

      <!-- Account Policy Note -->
      <div class="auth-policy p-3 rounded-xl bg-slate-50 border border-slate-200/90 flex items-start gap-2.5">
        <Info class="h-4 w-4 text-slate-400 mt-0.5 shrink-0" />
        <p class="text-[11px] text-slate-500 leading-relaxed">
          Public registrations are assigned the <span class="font-semibold text-slate-700">Requester</span> role. Contact a system administrator to request elevated Agent or Admin permissions.
        </p>
      </div>

      <div class="pt-1">
        <BaseButton
          type="submit"
          variant="primary"
          full-width
          class="!min-h-10"
          :loading="isSubmitting"
          :disabled="isSubmitting"
          size="md"
        >
          Create Account
        </BaseButton>
      </div>

      <div class="text-center pt-2">
        <p class="text-xs text-slate-500">
          Already have an account?
          <router-link to="/login" class="font-semibold text-brand-600 hover:text-brand-500 hover:underline">
            Sign in here
          </router-link>
        </p>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { User, Mail, Lock, Eye, EyeOff, AlertCircle, Info, X } from 'lucide-vue-next'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const authStore = useAuthStore()
const { success } = useNotification()

const showPassword = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref('')

const form = reactive({
  name: '',
  email: '',
  password: '',
  passwordConfirmation: ''
})

const errors = reactive({
  name: '',
  email: '',
  password: '',
  passwordConfirmation: ''
})

function validateForm() {
  let isValid = true
  errors.name = ''
  errors.email = ''
  errors.password = ''
  errors.passwordConfirmation = ''

  if (!form.name.trim()) {
    errors.name = 'Full name is required'
    isValid = false
  } else if (form.name.trim().length < 2) {
    errors.name = 'Name must be at least 2 characters'
    isValid = false
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!form.email.trim()) {
    errors.email = 'Email address is required'
    isValid = false
  } else if (!emailRegex.test(form.email.trim())) {
    errors.email = 'Please enter a valid email address'
    isValid = false
  }

  if (!form.password) {
    errors.password = 'Password is required'
    isValid = false
  } else if (form.password.length < 6) {
    errors.password = 'Password must be at least 6 characters'
    isValid = false
  }

  if (!form.passwordConfirmation) {
    errors.passwordConfirmation = 'Please confirm your password'
    isValid = false
  } else if (form.password !== form.passwordConfirmation) {
    errors.passwordConfirmation = 'Passwords do not match'
    isValid = false
  }

  return isValid
}

async function handleRegister() {
  if (isSubmitting.value) return
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isSubmitting.value = true

  try {
    // Strictly call authStore.register({ name, email, password })
    await authStore.register({
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password
    })

    success('Account registered successfully! Please sign in.')
    router.push('/login')
  } catch (err) {
    errorMessage.value = err.message || 'Registration failed'

    if (err.fieldErrors) {
      if (err.fieldErrors.name) errors.name = err.fieldErrors.name
      if (err.fieldErrors.email) errors.email = err.fieldErrors.email
      if (err.fieldErrors.password) errors.password = err.fieldErrors.password
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>
