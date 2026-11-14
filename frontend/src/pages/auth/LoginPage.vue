<template>
  <div>
    <!-- Form Heading -->
    <div class="auth-page-heading mb-6">
      <h2 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900">
        Sign in to your account
      </h2>
      <p class="mt-1 text-xs sm:text-sm text-slate-500">
        Enter your enterprise credentials to access your support workspace.
      </p>
    </div>

    <!-- Server / Authentication Error Alert -->
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
          <p class="font-semibold text-rose-900">Sign In Failed</p>
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

    <!-- Login Form -->
    <form @submit.prevent="handleLogin" class="auth-form space-y-4" novalidate>
      <BaseInput
        id="email"
        label="Email address"
        type="email"
        placeholder="name@company.com"
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
        placeholder="Enter your password"
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

      <div class="flex items-center justify-between pt-0.5">
        <label class="flex items-center gap-2 cursor-pointer select-none">
          <input
            id="remember-me"
            v-model="rememberMe"
            type="checkbox"
            class="h-4 w-4 rounded border-slate-300 text-brand-600 focus:ring-brand-500 cursor-pointer"
          />
          <span class="text-xs text-slate-600 font-medium">Remember my session</span>
        </label>
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
          Sign In
        </BaseButton>
      </div>

      <div class="text-center pt-1">
        <p class="text-xs text-slate-500">
          Need an account?
          <router-link to="/register" class="font-semibold text-brand-600 hover:text-brand-500 hover:underline">
            Register as Requester
          </router-link>
        </p>
      </div>

      <!-- Quick Demo Account Credentials Picker -->
      <div class="auth-demo mt-6 pt-5 border-t border-slate-100">
        <div class="flex items-center justify-between mb-2.5">
          <p class="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
            Fast Demo Login
          </p>
          <span class="text-[10px] text-slate-400">Click to autofill</span>
        </div>
        <div class="grid grid-cols-3 gap-2">
          <button
            type="button"
            @click="fillDemo('admin@example.com', 'Admin123!')"
            class="px-2.5 py-2 text-xs font-medium rounded-lg border border-slate-200 bg-slate-50 hover:bg-brand-50 hover:border-brand-200 hover:text-brand-700 text-slate-700 transition-all text-center flex flex-col items-center gap-0.5"
          >
            <span class="font-semibold text-slate-900">Admin</span>
            <span class="text-[10px] text-slate-400">Full Access</span>
          </button>
          <button
            type="button"
            @click="fillDemo('agent@example.com', 'Agent123!')"
            class="px-2.5 py-2 text-xs font-medium rounded-lg border border-slate-200 bg-slate-50 hover:bg-indigo-50 hover:border-indigo-200 hover:text-indigo-700 text-slate-700 transition-all text-center flex flex-col items-center gap-0.5"
          >
            <span class="font-semibold text-slate-900">Agent</span>
            <span class="text-[10px] text-slate-400">Support Staff</span>
          </button>
          <button
            type="button"
            @click="fillDemo('requester@example.com', 'Requester123!')"
            class="px-2.5 py-2 text-xs font-medium rounded-lg border border-slate-200 bg-slate-50 hover:bg-teal-50 hover:border-teal-200 hover:text-teal-700 text-slate-700 transition-all text-center flex flex-col items-center gap-0.5"
          >
            <span class="font-semibold text-slate-900">Requester</span>
            <span class="text-[10px] text-slate-400">Ticket Creator</span>
          </button>
        </div>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Mail, Lock, Eye, EyeOff, AlertCircle, X } from 'lucide-vue-next'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotification } from '@/composables/useNotification'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { success } = useNotification()

const showPassword = ref(false)
const isSubmitting = ref(false)
const rememberMe = ref(true)
const errorMessage = ref('')

const form = reactive({
  email: '',
  password: ''
})

const errors = reactive({
  email: '',
  password: ''
})

function fillDemo(email, password) {
  form.email = email
  form.password = password
  errors.email = ''
  errors.password = ''
  errorMessage.value = ''
}

function validateForm() {
  let isValid = true
  errors.email = ''
  errors.password = ''

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
  }

  return isValid
}

async function handleLogin() {
  if (isSubmitting.value) return
  errorMessage.value = ''

  if (!validateForm()) {
    return
  }

  isSubmitting.value = true

  try {
    const authData = await authStore.login({
      email: form.email.trim(),
      password: form.password
    })

    success(`Welcome back, ${authData.user.name || 'User'}!`)

    // Redirect to originally requested protected route or fallback to dashboard
    const redirectPath = route.query.redirect || '/dashboard'
    router.push(redirectPath)
  } catch (err) {
    errorMessage.value = err.message || 'Invalid credentials. Please verify your email and password.'

    if (err.fieldErrors) {
      if (err.fieldErrors.email) errors.email = err.fieldErrors.email
      if (err.fieldErrors.password) errors.password = err.fieldErrors.password
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>
