import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/auth/LoginPage.vue'),
    meta: {
      layout: 'auth',
      guestOnly: true,
      title: 'Sign In'
    }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/pages/auth/RegisterPage.vue'),
    meta: {
      layout: 'auth',
      guestOnly: true,
      title: 'Register Account'
    }
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/pages/dashboard/DashboardPage.vue'),
    meta: {
      layout: 'app',
      requiresAuth: true,
      title: 'Dashboard'
    }
  },
  {
    path: '/tickets',
    name: 'tickets',
    component: () => import('@/pages/tickets/TicketListPage.vue'),
    meta: {
      layout: 'app',
      requiresAuth: true,
      title: 'Support Tickets'
    }
  },
  {
    path: '/tickets/create',
    name: 'ticket-create',
    component: () => import('@/pages/tickets/TicketCreatePage.vue'),
    meta: {
      layout: 'app',
      requiresAuth: true,
      title: 'Create Ticket'
    }
  },
  {
    path: '/tickets/:id',
    name: 'ticket-detail',
    component: () => import('@/pages/tickets/TicketDetailPage.vue'),
    meta: {
      layout: 'app',
      requiresAuth: true,
      title: 'Ticket Details'
    }
  },
  {
    path: '/users',
    name: 'users',
    component: () => import('@/pages/users/UserManagementPage.vue'),
    meta: {
      layout: 'app',
      requiresAuth: true,
      roles: ['admin'],
      title: 'User Management'
    }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/pages/NotFoundPage.vue'),
    meta: {
      layout: 'app',
      title: '404 Not Found'
    }
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

// Navigation Guards: Session Restoration & Role-Based Authorization
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Restore and validate session before routing if not yet initialized
  if (!authStore.isInitialized) {
    await authStore.initAuth()
  }

  const isAuthenticated = authStore.isAuthenticated

  // 1. Check if route requires authentication
  if (to.matched.some((record) => record.meta.requiresAuth)) {
    if (!isAuthenticated) {
      // Preserve intended destination path in query param
      return next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
    }

    // 2. Role-based route guard (e.g. admin-only for /users)
    const requiredRoles = to.meta.roles
    if (requiredRoles && Array.isArray(requiredRoles)) {
      if (!requiredRoles.includes(authStore.role)) {
        // Forbidden: Redirect non-privileged user to dashboard
        return next({ path: '/dashboard' })
      }
    }

    return next()
  }

  // 3. Guest-only routes (login, register) should redirect authenticated users
  if (to.matched.some((record) => record.meta.guestOnly)) {
    if (isAuthenticated) {
      return next({ path: '/dashboard' })
    }
    return next()
  }

  // 4. Default: allow navigation
  next()
})

// Update document title dynamically
router.afterEach((to) => {
  const baseTitle = 'Helpdesk System'
  document.title = to.meta.title ? `${to.meta.title} | ${baseTitle}` : baseTitle
})

export default router
