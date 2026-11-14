<template>
  <aside
    :class="[
      'fixed inset-y-0 left-0 z-30 h-screen h-[100dvh] w-64 min-h-0 flex flex-col bg-slate-900 text-slate-300 transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-auto shrink-0 border-r border-slate-800/80',
      isOpen ? 'translate-x-0 shadow-2xl' : '-translate-x-full',
    ]"
  >
    <!-- Brand / Workspace Logo -->
    <div class="flex shrink-0 items-center justify-between h-16 px-4 bg-slate-950/70 border-b border-slate-800/80">
      <router-link to="/dashboard" class="flex items-center gap-3 group">
        <div
          class="h-9 w-9 rounded-xl bg-brand-600 flex items-center justify-center text-white shadow-md shadow-brand-500/25 group-hover:bg-brand-500 transition-colors"
        >
          <Ticket class="h-5 w-5" />
        </div>
        <div>
          <div class="flex items-center gap-1.5">
            <span class="text-sm font-bold text-white tracking-tight">Helpdesk</span>
            <span
              class="text-[10px] font-semibold uppercase px-1.5 py-0.2 bg-brand-500/20 text-brand-300 rounded border border-brand-500/30"
              >Pro</span
            >
          </div>
          <span class="text-[11px] block text-slate-400 font-medium">Ticketing Platform</span>
        </div>
      </router-link>

      <!-- Mobile Close Trigger -->
      <button
        @click="$emit('close')"
        type="button"
        class="lg:hidden p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors focus:outline-none"
        aria-label="Close navigation"
      >
        <X class="h-5 w-5" />
      </button>
    </div>

    <!-- Navigation Menu Items -->
    <nav class="app-sidebar-nav flex-1 min-h-0 overflow-y-auto overscroll-contain px-3 py-3 space-y-4">
      <!-- Section: Overview -->
      <div>
        <p class="px-3 mb-1.5 text-[9px] font-bold text-slate-400 uppercase tracking-[0.12em]">Main Workspace</p>
        <div class="space-y-1">
          <router-link to="/dashboard" v-slot="{ isActive, href, navigate }" custom>
            <a
              :href="href"
              @click="navigate"
              :class="[
                'flex min-h-9 items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-all duration-150',
                isActive
                  ? 'bg-brand-600 text-white font-semibold shadow-xs shadow-brand-500/25'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white',
              ]"
            >
              <LayoutDashboard class="h-4 w-4 shrink-0" />
              <span>Dashboard</span>
            </a>
          </router-link>
        </div>
      </div>

      <!-- Section: Tickets -->
      <div>
        <p class="px-3 mb-1.5 text-[9px] font-bold text-slate-400 uppercase tracking-[0.12em]">Ticket Operations</p>
        <div class="space-y-1">
          <router-link to="/tickets" v-slot="{ href, navigate }" custom>
            <a
              :href="href"
              @click="navigate"
              :class="[
                'flex min-h-9 items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-all duration-150',
                $route.path === '/tickets' || $route.name === 'ticket-detail'
                  ? 'bg-brand-600 text-white font-semibold shadow-xs shadow-brand-500/25'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white',
              ]"
            >
              <TicketCheck class="h-4 w-4 shrink-0" />
              <span>{{ authStore.isRequester ? "My Tickets" : "All Tickets" }}</span>
            </a>
          </router-link>

          <!-- Create Ticket (Requester & Admin) -->
        </div>
      </div>

      <!-- Section: Administration (Admin Role Only) -->
      <div v-if="authStore.isAdmin">
        <p class="px-3 mb-1.5 text-[9px] font-bold text-slate-400 uppercase tracking-[0.12em]">Administration</p>
        <div class="space-y-1">
          <router-link to="/users" v-slot="{ isActive, href, navigate }" custom>
            <a
              :href="href"
              @click="navigate"
              :class="[
                'flex min-h-9 items-center gap-2.5 px-3 py-2 rounded-lg text-xs font-medium transition-all duration-150',
                isActive
                  ? 'bg-brand-600 text-white font-semibold shadow-xs shadow-brand-500/25'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white',
              ]"
            >
              <Users class="h-4 w-4 shrink-0" />
              <span>User Management</span>
            </a>
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Sidebar Footer / User Profile Card -->
    <div class="shrink-0 p-2.5 border-t border-slate-800/80 bg-slate-950/60">
      <div v-if="authStore.isAuthenticated" class="flex items-center justify-between gap-2.5">
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="relative shrink-0">
            <div
              class="h-8 w-8 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-xs font-bold text-brand-300"
            >
              {{ userInitials }}
            </div>
            <span
              class="absolute bottom-0 right-0 h-2 w-2 rounded-full bg-emerald-500 ring-2 ring-slate-900"
              title="Active"
            ></span>
          </div>

          <div class="min-w-0 flex-1">
            <p class="text-xs font-semibold text-white truncate">
              {{ authStore.userName || "Account" }}
            </p>
            <div class="flex items-center gap-1.5 mt-0.5">
              <BaseBadge :variant="authStore.role" size="xs">
                {{ authStore.role }}
              </BaseBadge>
            </div>
          </div>
        </div>

        <button
          @click="authStore.logout"
          type="button"
          class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors shrink-0"
          title="Sign Out"
        >
          <LogOut class="h-4 w-4" />
        </button>
      </div>

      <div v-else class="text-center py-1">
        <router-link to="/login" class="text-xs font-medium text-brand-400 hover:text-brand-300">
          Sign in to your account &rarr;
        </router-link>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { computed } from "vue";
import { Ticket, LayoutDashboard, TicketCheck, PlusCircle, Users, LogOut, X } from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";
import BaseBadge from "@/components/common/BaseBadge.vue";

defineProps({
  isOpen: {
    type: Boolean,
    default: false,
  },
});

defineEmits(["close"]);

const authStore = useAuthStore();

const userInitials = computed(() => {
  if (!authStore.userName) return "U";
  return authStore.userName
    .split(" ")
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join("")
    .toUpperCase();
});
</script>
