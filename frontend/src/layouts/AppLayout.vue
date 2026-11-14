<template>
  <div class="h-screen h-[100dvh] overflow-hidden bg-slate-50 flex font-sans antialiased">
    <!-- Mobile Backdrop Overlay -->
    <transition
      enter-active-class="transition-opacity ease-linear duration-200"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity ease-linear duration-200"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="isSidebarOpen"
        @click="isSidebarOpen = false"
        class="fixed inset-0 z-20 bg-slate-900/60 backdrop-blur-xs lg:hidden"
      ></div>
    </transition>

    <!-- Sidebar Navigation -->
    <AppSidebar
      :is-open="isSidebarOpen"
      @close="isSidebarOpen = false"
    />

    <!-- Main Content Column -->
    <div class="flex-1 flex flex-col min-w-0 min-h-0 h-screen overflow-hidden">
      <!-- Top Navigation Header -->
      <AppHeader @toggle-sidebar="isSidebarOpen = !isSidebarOpen" />

      <!-- Main Content Container -->
      <main class="flex-1 min-h-0 overflow-y-auto p-4 sm:p-6 lg:p-8 max-w-7xl w-full mx-auto">
        <slot>
          <router-view />
        </slot>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'

const isSidebarOpen = ref(false)
const route = useRoute()

// Automatically close sidebar on mobile navigation
watch(
  () => route.path,
  () => {
    isSidebarOpen.value = false
  }
)
</script>
