<template>
  <div class="space-y-6">
    <!-- 1. Executive Hero Banner -->
    <div
      class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-slate-900 to-indigo-950 p-6 sm:p-8 text-white border border-slate-800 shadow-sm"
    >
      <!-- Decorative Ambient Glow -->
      <div class="absolute -top-24 -right-24 h-64 w-64 rounded-full bg-brand-500/15 blur-3xl pointer-events-none"></div>
      <div
        class="absolute -bottom-24 -left-24 h-64 w-64 rounded-full bg-indigo-500/10 blur-3xl pointer-events-none"
      ></div>

      <div class="relative z-10 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
        <div class="space-y-2 max-w-2xl">
          <!-- Status Pill & Role Badge -->
          <div class="flex items-center gap-2.5 flex-wrap">
            <span
              class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/20"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
              Live Support Operational
            </span>
            <BaseBadge :variant="authStore.role" size="xs">
              {{ authStore.role }}
            </BaseBadge>
          </div>

          <!-- Greeting Heading -->
          <h1 class="text-2xl sm:text-3xl font-extrabold tracking-tight text-white leading-tight">
            {{ dashboardTitle }}
          </h1>

          <!-- Role-Targeted Context Subtitle -->
          <p class="text-xs sm:text-sm text-slate-300 leading-relaxed">
            Welcome back, <span class="font-semibold text-white">{{ authStore.userName || "User" }}</span
            >.
            <span v-if="authStore.isRequester">
              Track active inquiries, verify resolutions, or open a new service request with our engineering desk.
            </span>
            <span v-else-if="authStore.isAgent">
              Review assigned support queues, triage incoming issues, and maintain SLA response metrics.
            </span>
            <span v-else-if="authStore.isAdmin">
              Platform administration, enterprise queue oversight, user permissions, and SLA analytics.
            </span>
          </p>
        </div>

        <!-- Action Ribbon inside Hero -->
        <div class="flex items-center gap-3 shrink-0 flex-wrap">
          <router-link v-if="authStore.isAdmin" to="/users">
            <BaseButton variant="secondary" size="md">
              <template #icon-left>
                <Users class="h-4 w-4 mr-1.5 text-slate-600" />
              </template>
              Manage Users
            </BaseButton>
          </router-link>

          <router-link v-if="authStore.isAgent" to="/tickets">
            <BaseButton variant="primary" size="md">
              <template #icon-left>
                <TicketCheck class="h-4 w-4 mr-1.5" />
              </template>
              Browse All Tickets
            </BaseButton>
          </router-link>
        </div>
      </div>
    </div>

    <!-- 2. Statistics & KPI Metrics Grid with Visual Distribution Bars -->
    <!-- ADMIN METRICS (5 Key Metrics from real backend APIs) -->
    <div v-if="authStore.isAdmin" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5 sm:gap-4">
      <!-- Total Tickets -->
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Total Tickets</p>
          <div class="p-2 bg-brand-50 text-brand-600 rounded-lg">
            <TicketIcon class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.total }}</span>
        </div>
        <div class="mt-2.5 w-full bg-slate-100 rounded-full h-1 overflow-hidden">
          <div class="bg-brand-600 h-full rounded-full w-full"></div>
        </div>
        <p class="mt-1.5 text-[10px] text-slate-400 font-medium">All system records</p>
      </BaseCard>

      <!-- Open Tickets -->
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-sky-600 uppercase tracking-wider">Open</p>
          <div class="p-2 bg-sky-50 text-sky-600 rounded-lg">
            <Clock class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.open }}</span>
        </div>
        <div class="mt-2.5 w-full bg-slate-100 rounded-full h-1 overflow-hidden">
          <div
            class="bg-sky-500 h-full rounded-full transition-all duration-500"
            :style="{ width: `${openPercent}%` }"
          ></div>
        </div>
        <p class="mt-1.5 text-[10px] text-slate-500 font-medium flex items-center justify-between">
          <span>Needs triage</span>
          <span class="text-sky-600 font-bold">{{ openPercent }}%</span>
        </p>
      </BaseCard>

      <!-- In Progress -->
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-amber-600 uppercase tracking-wider">In Progress</p>
          <div class="p-2 bg-amber-50 text-amber-600 rounded-lg">
            <AlertCircle class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.inProgress }}</span>
        </div>
        <div class="mt-2.5 w-full bg-slate-100 rounded-full h-1 overflow-hidden">
          <div
            class="bg-amber-500 h-full rounded-full transition-all duration-500"
            :style="{ width: `${inProgressPercent}%` }"
          ></div>
        </div>
        <p class="mt-1.5 text-[10px] text-slate-500 font-medium flex items-center justify-between">
          <span>Under investigation</span>
          <span class="text-amber-600 font-bold">{{ inProgressPercent }}%</span>
        </p>
      </BaseCard>

      <!-- Resolved -->
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-emerald-600 uppercase tracking-wider">Resolved</p>
          <div class="p-2 bg-emerald-50 text-emerald-600 rounded-lg">
            <CheckCircle2 class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.resolved }}</span>
        </div>
        <div class="mt-2.5 w-full bg-slate-100 rounded-full h-1 overflow-hidden">
          <div
            class="bg-emerald-500 h-full rounded-full transition-all duration-500"
            :style="{ width: `${resolvedPercent}%` }"
          ></div>
        </div>
        <p class="mt-1.5 text-[10px] text-slate-500 font-medium flex items-center justify-between">
          <span>Awaiting close</span>
          <span class="text-emerald-600 font-bold">{{ resolvedPercent }}%</span>
        </p>
      </BaseCard>

      <!-- Closed -->
      <BaseCard no-padding class="p-4 sm:p-5 col-span-2 sm:col-span-1">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">Closed</p>
          <div class="p-2 bg-slate-100 text-slate-600 rounded-lg">
            <Archive class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.closed }}</span>
        </div>
        <div class="mt-2.5 w-full bg-slate-100 rounded-full h-1 overflow-hidden">
          <div
            class="bg-slate-400 h-full rounded-full transition-all duration-500"
            :style="{ width: `${closedPercent}%` }"
          ></div>
        </div>
        <p class="mt-1.5 text-[10px] text-slate-500 font-medium flex items-center justify-between">
          <span>Archived history</span>
          <span class="text-slate-600 font-bold">{{ closedPercent }}%</span>
        </p>
      </BaseCard>
    </div>

    <!-- AGENT METRICS (4 Key Operational Metrics) -->
    <div v-else-if="authStore.isAgent" class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 sm:gap-4">
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-brand-600 uppercase tracking-wider">Assigned to Me</p>
          <div class="p-2 bg-brand-50 text-brand-600 rounded-lg">
            <UserCheck class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.assignedToMe }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Your active workload</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-sky-600 uppercase tracking-wider">Open Tickets</p>
          <div class="p-2 bg-sky-50 text-sky-600 rounded-lg">
            <Clock class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.open }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Unresolved queue</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-amber-600 uppercase tracking-wider">In Progress</p>
          <div class="p-2 bg-amber-50 text-amber-600 rounded-lg">
            <AlertCircle class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.inProgress }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Being diagnosed</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-emerald-600 uppercase tracking-wider">Resolved</p>
          <div class="p-2 bg-emerald-50 text-emerald-600 rounded-lg">
            <CheckCircle2 class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.resolved }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Successfully completed</p>
      </BaseCard>
    </div>

    <!-- REQUESTER METRICS (4 Key Customer Metrics) -->
    <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-3.5 sm:gap-4">
      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-slate-500 uppercase tracking-wider">My Total Tickets</p>
          <div class="p-2 bg-brand-50 text-brand-600 rounded-lg">
            <TicketIcon class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.total }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Submitted by your account</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-sky-600 uppercase tracking-wider">Open</p>
          <div class="p-2 bg-sky-50 text-sky-600 rounded-lg">
            <Clock class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.open }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Queued for support</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-amber-600 uppercase tracking-wider">In Progress</p>
          <div class="p-2 bg-amber-50 text-amber-600 rounded-lg">
            <AlertCircle class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.inProgress }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Under active investigation</p>
      </BaseCard>

      <BaseCard no-padding class="p-4 sm:p-5">
        <div class="flex items-center justify-between">
          <p class="text-[11px] font-semibold text-emerald-600 uppercase tracking-wider">Resolved</p>
          <div class="p-2 bg-emerald-50 text-emerald-600 rounded-lg">
            <CheckCircle2 class="h-4 w-4" />
          </div>
        </div>
        <div class="mt-3 text-2xl sm:text-3xl font-extrabold text-slate-900 tabular-nums">
          <span v-if="isLoadingMetrics" class="text-slate-300 animate-pulse">--</span>
          <span v-else>{{ stats.resolved }}</span>
        </div>
        <p class="mt-2 text-[11px] text-slate-500 font-medium">Ready for your verification</p>
      </BaseCard>
    </div>

    <!-- 3. PRIMARY SECTION: Recent Tickets Table (Hero Width - Fully 1366px Laptop Responsive!) -->
    <BaseCard no-padding class="overflow-hidden">
      <!-- Card Header Ribbon -->
      <div
        class="p-5 sm:p-6 border-b border-slate-100 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4"
      >
        <div>
          <div class="flex items-center gap-2">
            <h2 class="text-base font-bold text-slate-900">
              {{
                authStore.isAdmin
                  ? "Recent Ticket Activity"
                  : authStore.isRequester
                    ? "My Recent Tickets"
                    : "Recent Support Queue"
              }}
            </h2>
            <span
              class="px-2 py-0.5 rounded-full text-[11px] font-semibold bg-brand-50 text-brand-700 border border-brand-200/60"
            >
              {{ recentTotal }} Total
            </span>
          </div>
          <p class="text-xs text-slate-500 mt-0.5">
            {{
              authStore.isAdmin
                ? "Real-time overview of ticket submissions and workflow updates across the organization."
                : "Latest support requests logged in the helpdesk system."
            }}
          </p>
        </div>

        <router-link
          to="/tickets"
          class="inline-flex items-center gap-1.5 text-xs font-semibold text-brand-600 hover:text-brand-700 transition-colors"
        >
          <span>View All Tickets</span>
          <ChevronRight class="h-4 w-4" />
        </router-link>
      </div>

      <!-- Loading State Skeleton -->
      <div v-if="isLoadingRecentTickets" class="p-6 space-y-4">
        <div
          v-for="i in 5"
          :key="i"
          class="flex items-center justify-between py-2 border-b border-slate-100 last:border-0"
        >
          <div class="space-y-1.5 flex-1 pr-6">
            <BaseSkeleton width="60%" height="1rem" />
            <BaseSkeleton width="30%" height="0.75rem" />
          </div>
          <BaseSkeleton width="5rem" height="1.5rem" rounded-full />
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="recentTickets.length === 0" class="p-8">
        <EmptyState
          title="No recent tickets found"
          :description="
            authStore.isAgent
              ? 'There are currently no tickets logged in the queue.'
              : 'Submit your first support ticket to track resolution progress.'
          "
        >
          <template #action>
            <router-link v-if="authStore.isRequester || authStore.isAdmin" to="/tickets/create">
              <BaseButton variant="primary" size="sm">
                <template #icon-left>
                  <Plus class="h-4 w-4 mr-1.5" />
                </template>
                Create Support Ticket
              </BaseButton>
            </router-link>
          </template>
        </EmptyState>
      </div>

      <!-- Modern Table View (Fits effortlessly within 1366px viewport without horizontal cutoff!) -->
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-sm divide-y divide-slate-100">
          <thead>
            <tr class="bg-slate-50/80 text-[11px] font-semibold text-slate-500 uppercase tracking-wider">
              <th v-if="authStore.isAdmin" scope="col" class="w-12 py-3 px-4 text-center">#</th>
              <th scope="col" class="py-3 px-4 sm:px-6">Ticket Title & ID</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Category</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Priority</th>
              <th scope="col" class="py-3 px-3 whitespace-nowrap">Status</th>
              <th scope="col" class="py-3 px-4 text-right whitespace-nowrap">Created</th>
              <th scope="col" class="py-3 px-4 sm:px-6 text-right whitespace-nowrap">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 bg-white">
            <tr
              v-for="(ticket, index) in recentTickets"
              :key="ticket.id"
              class="hover:bg-slate-50/90 transition-colors group cursor-pointer"
              @click="$router.push(`/tickets/${ticket.id}`)"
            >
              <!-- Index Number for Admin -->
              <td
                v-if="authStore.isAdmin"
                class="py-3.5 px-4 text-center text-xs font-medium text-slate-400 tabular-nums"
              >
                {{ (recentPage - 1) * recentPageSize + index + 1 }}
              </td>

              <!-- Title & Meta -->
              <td class="py-3.5 px-4 sm:px-6">
                <div class="min-w-0 max-w-md">
                  <router-link
                    :to="`/tickets/${ticket.id}`"
                    class="font-semibold text-slate-900 group-hover:text-brand-600 transition-colors block truncate"
                    @click.stop
                  >
                    {{ ticket.title }}
                  </router-link>
                  <div class="flex items-center gap-2 mt-0.5 text-[11px] text-slate-400">
                    <span class="font-mono text-slate-500 font-medium">#{{ ticket.id.substring(0, 8) }}</span>
                    <span v-if="ticket.requester?.name" class="truncate"> &bull; by {{ ticket.requester.name }} </span>
                  </div>
                </div>
              </td>

              <!-- Category -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <span
                  class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-medium bg-slate-100 text-slate-700"
                >
                  {{ ticket.category }}
                </span>
              </td>

              <!-- Priority -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <BaseBadge :variant="ticket.priority" size="xs">
                  {{ ticket.priority }}
                </BaseBadge>
              </td>

              <!-- Status -->
              <td class="py-3.5 px-3 whitespace-nowrap">
                <BaseBadge :variant="ticket.status" dot size="xs">
                  {{ ticket.status }}
                </BaseBadge>
              </td>

              <!-- Created Date -->
              <td class="py-3.5 px-4 text-right text-xs text-slate-500 whitespace-nowrap">
                {{ formatDate(ticket.created_at) }}
              </td>

              <!-- View Button -->
              <td class="py-3.5 px-4 sm:px-6 text-right whitespace-nowrap">
                <router-link :to="`/tickets/${ticket.id}`" @click.stop>
                  <BaseButton variant="secondary" size="xs"> View </BaseButton>
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Integrated Pagination Bar -->
      <div
        v-if="recentTotal > recentPageSize"
        class="p-4 bg-slate-50/70 border-t border-slate-100 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 text-xs text-slate-600"
      >
        <p class="text-xs text-slate-500">
          Showing <span class="font-semibold text-slate-900">{{ recentRangeStart }}</span> to
          <span class="font-semibold text-slate-900">{{ recentRangeEnd }}</span> of
          <span class="font-semibold text-slate-900">{{ recentTotal }}</span> tickets
        </p>

        <div class="flex items-center gap-1 self-center sm:self-auto">
          <button
            type="button"
            class="inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="recentPage <= 1 || isLoadingRecentTickets"
            @click="loadRecentTickets(recentPage - 1)"
          >
            <ChevronLeft class="h-3.5 w-3.5" />
            <span>Previous</span>
          </button>

          <template v-for="pageItem in recentPageItems" :key="pageItem">
            <span v-if="typeof pageItem === 'string'" class="px-1 text-slate-400">…</span>
            <button
              v-else
              type="button"
              class="h-8 min-w-8 rounded-lg px-2 text-xs font-semibold transition"
              :class="
                pageItem === recentPage
                  ? 'bg-brand-600 text-white shadow-xs'
                  : 'border border-slate-200 bg-white text-slate-700 hover:bg-slate-50'
              "
              :disabled="isLoadingRecentTickets"
              @click="loadRecentTickets(pageItem)"
            >
              {{ pageItem }}
            </button>
          </template>

          <button
            type="button"
            class="inline-flex h-8 items-center gap-1 rounded-lg border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
            :disabled="recentPage >= recentTotalPages || isLoadingRecentTickets"
            @click="loadRecentTickets(recentPage + 1)"
          >
            <span>Next</span>
            <ChevronRight class="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </BaseCard>

    <!-- 4. BOTTOM GRID: Operational Intelligence & Quick Actions -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <!-- Card A: Quick Action Shortcuts -->
      <BaseCard title="Quick Shortcuts" subtitle="Fast operational tasks">
        <div class="space-y-2">
          <!-- Create Ticket Action -->
          <router-link
            v-if="authStore.isRequester || authStore.isAdmin"
            to="/tickets/create"
            class="flex items-center justify-between p-3 rounded-xl border border-slate-200/80 bg-white hover:bg-brand-50/50 hover:border-brand-200 text-xs transition group"
          >
            <div class="flex items-center gap-2.5">
              <div class="p-1.5 rounded-lg bg-brand-50 text-brand-600 group-hover:bg-brand-100 transition">
                <Plus class="h-4 w-4" />
              </div>
              <div>
                <p class="font-semibold text-slate-900 group-hover:text-brand-900">Submit New Ticket</p>
                <p class="text-[11px] text-slate-400">Log an issue or service request</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 text-slate-400 group-hover:text-brand-600" />
          </router-link>

          <!-- Browse Tickets Action -->
          <router-link
            to="/tickets"
            class="flex items-center justify-between p-3 rounded-xl border border-slate-200/80 bg-white hover:bg-slate-50 hover:border-slate-300 text-xs transition group"
          >
            <div class="flex items-center gap-2.5">
              <div class="p-1.5 rounded-lg bg-slate-100 text-slate-600 group-hover:bg-slate-200 transition">
                <TicketCheck class="h-4 w-4" />
              </div>
              <div>
                <p class="font-semibold text-slate-900">Browse Ticket Directory</p>
                <p class="text-[11px] text-slate-400">Search, filter & sort all records</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 text-slate-400 group-hover:text-slate-600" />
          </router-link>

          <!-- User Management (Admin) -->
          <router-link
            v-if="authStore.isAdmin"
            to="/users"
            class="flex items-center justify-between p-3 rounded-xl border border-slate-200/80 bg-white hover:bg-purple-50/50 hover:border-purple-200 text-xs transition group"
          >
            <div class="flex items-center gap-2.5">
              <div class="p-1.5 rounded-lg bg-purple-50 text-purple-600 group-hover:bg-purple-100 transition">
                <Users class="h-4 w-4" />
              </div>
              <div>
                <p class="font-semibold text-slate-900 group-hover:text-purple-900">User Management</p>
                <p class="text-[11px] text-slate-400">Manage accounts & assign roles</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 text-slate-400 group-hover:text-purple-600" />
          </router-link>
        </div>
      </BaseCard>

      <!-- Card B: Queue Health & SLA Policy -->
      <BaseCard v-if="authStore.isAdmin" title="Enterprise System Health" subtitle="Operational metrics & security">
        <div class="space-y-3 text-xs">
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <span class="text-slate-500">Registered Accounts</span>
            <span class="font-bold text-slate-900">{{ totalUsersCount }} users</span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <span class="text-slate-500">Active Pipeline</span>
            <span class="font-bold text-brand-600">{{ stats.open + stats.inProgress }} tickets</span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-slate-100">
            <span class="text-slate-500">Resolution Rate</span>
            <span class="font-bold text-emerald-600">
              {{ stats.total > 0 ? Math.round((stats.resolved / stats.total) * 100) : 0 }}% completed
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500">Security & RBAC</span>
            <span class="font-semibold text-emerald-600 flex items-center gap-1">
              <CheckCircle2 class="h-3.5 w-3.5" />
              Active
            </span>
          </div>
        </div>
      </BaseCard>

      <!-- Agent Queue Guidance -->
      <BaseCard v-else-if="authStore.isAgent" title="Agent Operational SLA" subtitle="Standard operating procedures">
        <div class="space-y-2.5 text-xs text-slate-600">
          <div class="p-2.5 rounded-lg bg-sky-50/70 border border-sky-100 flex items-start gap-2">
            <Clock class="h-4 w-4 text-sky-600 mt-0.5 shrink-0" />
            <div>
              <p class="font-semibold text-sky-900">Open Tickets Triage</p>
              <p class="text-[11px] text-sky-700 leading-snug">
                Claim unassigned issues or set In Progress within 2 hours of submission.
              </p>
            </div>
          </div>
          <div class="p-2.5 rounded-lg bg-amber-50/70 border border-amber-100 flex items-start gap-2">
            <AlertCircle class="h-4 w-4 text-amber-600 mt-0.5 shrink-0" />
            <div>
              <p class="font-semibold text-amber-900">Internal Diagnostics</p>
              <p class="text-[11px] text-amber-700 leading-snug">
                Always log technical reproduction steps as staff-only notes.
              </p>
            </div>
          </div>
        </div>
      </BaseCard>

      <!-- Requester Guidance -->
      <BaseCard v-else title="Support Resolution Flow" subtitle="What to expect after submitting">
        <ol class="space-y-2 text-xs text-slate-600 list-decimal list-inside leading-relaxed">
          <li><span class="font-semibold text-slate-900">Submit inquiry:</span> Detail steps & error codes.</li>
          <li>
            <span class="font-semibold text-slate-900">Investigation:</span> Agent investigates & updates progress.
          </li>
          <li>
            <span class="font-semibold text-slate-900">Verify & Close:</span> Reopen if unresolved, or close once
            satisfied.
          </li>
        </ol>
      </BaseCard>

      <!-- Card C: Queue Status Distribution -->
      <BaseCard title="Queue Composition" subtitle="Visual breakdown of total tickets">
        <div class="space-y-3 text-xs">
          <!-- Stacked Bar -->
          <div class="h-2.5 w-full bg-slate-100 rounded-full overflow-hidden flex">
            <div class="bg-sky-500 h-full" :style="{ width: `${openPercent}%` }" title="Open"></div>
            <div class="bg-amber-500 h-full" :style="{ width: `${inProgressPercent}%` }" title="In Progress"></div>
            <div class="bg-emerald-500 h-full" :style="{ width: `${resolvedPercent}%` }" title="Resolved"></div>
            <div class="bg-slate-400 h-full" :style="{ width: `${closedPercent}%` }" title="Closed"></div>
          </div>

          <!-- Legend -->
          <div class="grid grid-cols-2 gap-2 text-[11px]">
            <div class="flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-sky-500 shrink-0"></span>
              <span class="text-slate-600">Open ({{ stats.open }})</span>
            </div>
            <div class="flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-amber-500 shrink-0"></span>
              <span class="text-slate-600">In Progress ({{ stats.inProgress }})</span>
            </div>
            <div class="flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-emerald-500 shrink-0"></span>
              <span class="text-slate-600">Resolved ({{ stats.resolved }})</span>
            </div>
            <div class="flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-slate-400 shrink-0"></span>
              <span class="text-slate-600">Closed ({{ stats.closed }})</span>
            </div>
          </div>
        </div>
      </BaseCard>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import {
  Ticket as TicketIcon,
  Clock,
  AlertCircle,
  CheckCircle2,
  Archive,
  UserCheck,
  Users,
  Plus,
  TicketCheck,
  ChevronLeft,
  ChevronRight,
} from "lucide-vue-next";
import BaseCard from "@/components/common/BaseCard.vue";
import BaseBadge from "@/components/common/BaseBadge.vue";
import BaseButton from "@/components/common/BaseButton.vue";
import BaseSkeleton from "@/components/common/BaseSkeleton.vue";
import EmptyState from "@/components/common/EmptyState.vue";
import { useAuthStore } from "@/stores/auth";
import ticketService from "@/services/ticketService";
import userService from "@/services/userService";
import { formatDate } from "@/utils/formatters";

const authStore = useAuthStore();

const isLoadingMetrics = ref(true);
const isLoadingRecentTickets = ref(true);
const recentTickets = ref([]);
const recentPage = ref(1);
const recentPageSize = 5;
const recentTotal = ref(0);
const totalUsersCount = ref(0);

const recentTotalPages = computed(() => Math.ceil(recentTotal.value / recentPageSize));
const recentRangeStart = computed(() => (recentTotal.value ? (recentPage.value - 1) * recentPageSize + 1 : 0));
const recentRangeEnd = computed(() => Math.min(recentPage.value * recentPageSize, recentTotal.value));
const recentPageItems = computed(() => {
  const totalPages = recentTotalPages.value;
  if (totalPages <= 5) return Array.from({ length: totalPages }, (_, index) => index + 1);

  const start = Math.max(1, Math.min(recentPage.value - 1, totalPages - 2));
  const end = Math.min(totalPages, start + 2);
  const items = [];
  if (start > 1) items.push(1);
  if (start > 2) items.push("leading-ellipsis");
  for (let page = start; page <= end; page += 1) items.push(page);
  if (end < totalPages - 1) items.push("trailing-ellipsis");
  if (end < totalPages) items.push(totalPages);
  return items;
});

const stats = reactive({
  total: 0,
  open: 0,
  inProgress: 0,
  resolved: 0,
  closed: 0,
  assignedToMe: 0,
});

const openPercent = computed(() => (stats.total > 0 ? Math.round((stats.open / stats.total) * 100) : 0));
const inProgressPercent = computed(() => (stats.total > 0 ? Math.round((stats.inProgress / stats.total) * 100) : 0));
const resolvedPercent = computed(() => (stats.total > 0 ? Math.round((stats.resolved / stats.total) * 100) : 0));
const closedPercent = computed(() => (stats.total > 0 ? Math.round((stats.closed / stats.total) * 100) : 0));

const dashboardTitle = computed(() => {
  if (authStore.isAdmin) return "Administrator Overview";
  if (authStore.isAgent) return "Agent Support Desk";
  return "Customer Support Portal";
});

async function loadRecentTickets(page = 1, force = false) {
  if (isLoadingRecentTickets.value && !force) return null;
  const targetPage = Math.max(1, Math.min(page, recentTotalPages.value || 1));
  if (targetPage === recentPage.value && recentTickets.value.length > 0 && !force) return;

  isLoadingRecentTickets.value = true;
  try {
    const response = await ticketService.listTickets({
      page: targetPage,
      limit: recentPageSize,
      sort_by: "created_at",
      order: "desc",
    });
    recentTickets.value = response.data || [];
    recentPage.value = Number(response.pagination?.page) || targetPage;
    recentTotal.value = Number(response.pagination?.total) || 0;
    return response;
  } catch (err) {
    console.error("Failed to load recent dashboard tickets:", err);
    return null;
  } finally {
    isLoadingRecentTickets.value = false;
  }
}

async function loadDashboardData() {
  isLoadingMetrics.value = true;
  try {
    const recentRes = await loadRecentTickets(1, true);
    stats.total = Number(recentRes?.pagination?.total) || 0;

    const requests = [
      ticketService.listTickets({ status: "open", limit: 1 }),
      ticketService.listTickets({ status: "in_progress", limit: 1 }),
      ticketService.listTickets({ status: "resolved", limit: 1 }),
      ticketService.listTickets({ status: "closed", limit: 1 }),
    ];

    if (authStore.isAgent && authStore.user?.id) {
      requests.push(ticketService.listTickets({ assignee_id: authStore.user.id, limit: 1 }));
    }

    if (authStore.isAdmin) {
      requests.push(userService.listUsers({ limit: 1 }));
    }

    const results = await Promise.allSettled(requests);

    if (results[0].status === "fulfilled") {
      stats.open = Number(results[0].value.pagination?.total) || 0;
    }
    if (results[1].status === "fulfilled") {
      stats.inProgress = Number(results[1].value.pagination?.total) || 0;
    }
    if (results[2].status === "fulfilled") {
      stats.resolved = Number(results[2].value.pagination?.total) || 0;
    }
    if (results[3].status === "fulfilled") {
      stats.closed = Number(results[3].value.pagination?.total) || 0;
    }

    if (authStore.isAgent && results[4] && results[4].status === "fulfilled") {
      stats.assignedToMe = Number(results[4].value.pagination?.total) || 0;
    }

    if (authStore.isAdmin && results[4] && results[4].status === "fulfilled") {
      totalUsersCount.value = Number(results[4].value.pagination?.total) || 0;
    }
  } catch (err) {
    console.error("Failed to load dashboard metrics:", err);
  } finally {
    isLoadingMetrics.value = false;
  }
}

onMounted(() => {
  loadDashboardData();
});
</script>
