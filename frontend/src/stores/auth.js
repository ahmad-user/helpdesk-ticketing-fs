import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authService from '@/services/authService'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('auth_token') || null)
  const isLoading = ref(false)
  const isInitialLoading = ref(true)
  const isInitialized = ref(false)
  const error = ref(null)

  // Computed state
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const role = computed(() => user.value?.role || null)
  const isAdmin = computed(() => role.value === 'admin')
  const isAgent = computed(() => role.value === 'agent')
  const isRequester = computed(() => role.value === 'requester')
  const userName = computed(() => user.value?.name || '')
  const userEmail = computed(() => user.value?.email || '')

  /**
   * Initialize and restore session on page reload via GET /api/me
   */
  async function initAuth() {
    if (isInitialized.value) {
      return isAuthenticated.value
    }

    const savedToken = localStorage.getItem('auth_token')
    if (!savedToken) {
      token.value = null
      user.value = null
      isInitialLoading.value = false
      isInitialized.value = true
      return false
    }

    token.value = savedToken
    isInitialLoading.value = true

    try {
      const response = await authService.getMe()
      // Backend envelope: { success: true, message: "...", data: UserResponse }
      user.value = response.data
      isInitialized.value = true
      return true
    } catch (err) {
      // Clear token if invalid or expired
      clearSession()
      isInitialized.value = true
      return false
    } finally {
      isInitialLoading.value = false
    }
  }

  /**
   * Login user via POST /api/auth/login
   * @param {Object} credentials - { email, password }
   */
  async function login(credentials) {
    isLoading.value = true
    error.value = null

    try {
      const response = await authService.login(credentials)
      // Backend envelope: { success: true, message: "...", data: { token, user } }
      const authData = response.data
      token.value = authData.token
      user.value = authData.user

      localStorage.setItem('auth_token', authData.token)
      isInitialized.value = true

      return authData
    } catch (err) {
      error.value = err.message || 'Login failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Register user via POST /api/auth/register
   * @param {Object} userData - { name, email, password }
   */
  async function register(userData) {
    isLoading.value = true
    error.value = null

    try {
      const response = await authService.register(userData)
      return response.data
    } catch (err) {
      error.value = err.message || 'Registration failed'
      throw err
    } finally {
      isLoading.value = false
    }
  }

  /**
   * Refresh current user profile via GET /api/me
   */
  async function fetchProfile() {
    if (!token.value) return null
    try {
      const response = await authService.getMe()
      user.value = response.data
      return user.value
    } catch (err) {
      clearSession()
      throw err
    }
  }

  /**
   * Clear all session data from reactive state and localStorage
   */
  function clearSession() {
    user.value = null
    token.value = null
    localStorage.removeItem('auth_token')
  }

  /**
   * Logout user and redirect to login page
   */
  function logout() {
    clearSession()
    router.push('/login')
  }

  return {
    user,
    token,
    isLoading,
    isInitialLoading,
    isInitialized,
    error,
    isAuthenticated,
    role,
    isAdmin,
    isAgent,
    isRequester,
    userName,
    userEmail,
    initAuth,
    login,
    register,
    fetchProfile,
    clearSession,
    logout
  }
})

export default useAuthStore
