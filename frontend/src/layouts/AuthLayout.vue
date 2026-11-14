<template>
  <div class="auth-shell min-h-[100svh] bg-slate-50 flex flex-col lg:h-[100svh] lg:min-h-[680px] lg:flex-row antialiased">
    <!-- Left Hero Showcase Panel (Desktop & Laptop) -->
    <div
      class="auth-brand-panel hidden lg:flex lg:h-full lg:w-1/2 xl:w-5/12 bg-slate-950 text-white p-8 xl:p-10 flex-col justify-between relative overflow-hidden border-r border-slate-800/80 shrink-0"
    >
      <!-- Background Ambient Glow & Grid -->
      <div
        class="absolute -top-32 -left-32 w-96 h-96 bg-brand-600/20 rounded-full blur-3xl pointer-events-none"
      ></div>
      <div
        class="absolute -bottom-32 -right-32 w-96 h-96 bg-indigo-500/20 rounded-full blur-3xl pointer-events-none"
      ></div>
      <div
        class="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-slate-900/60 via-slate-950 to-slate-950 pointer-events-none"
      ></div>

      <!-- Top Branding -->
      <div class="relative z-10">
        <router-link to="/login" class="inline-flex items-center gap-3 group">
          <div
            class="h-10 w-10 rounded-xl bg-brand-600 flex items-center justify-center text-white shadow-lg shadow-brand-500/30 group-hover:bg-brand-500 transition-colors"
          >
            <Ticket class="h-5 w-5" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <span class="text-lg font-bold tracking-tight text-white leading-none">Helpdesk</span>
              <span
                class="text-[10px] font-bold uppercase px-1.5 py-0.5 bg-brand-500/20 text-brand-300 rounded border border-brand-500/30"
              >
                Enterprise
              </span>
            </div>
            <span class="text-xs text-slate-400 font-medium">Ticketing Platform</span>
          </div>
        </router-link>
      </div>

      <!-- Middle Value Proposition -->
      <div class="auth-brand-content relative z-10 my-auto py-6 space-y-6">
        <div class="space-y-3 max-w-md">
          <div
            class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-slate-800/80 border border-slate-700/80 text-xs font-medium text-brand-300"
          >
            <span class="h-2 w-2 rounded-full bg-brand-400 animate-pulse"></span>
            <span>Version 2.0 Enterprise SaaS</span>
          </div>
          <h1 class="text-3xl xl:text-4xl font-extrabold tracking-tight text-white leading-tight">
            High-velocity support & incident resolution.
          </h1>
          <p class="text-sm text-slate-400 leading-relaxed">
            Unify ticketing queues, automate SLA compliance, and empower support engineers with
            private diagnostics in a single workspace.
          </p>
        </div>

        <!-- Capability Checklist -->
        <div class="space-y-3 max-w-md">
          <div class="flex items-start gap-3">
            <div
              class="h-6 w-6 rounded-lg bg-brand-500/10 border border-brand-500/20 text-brand-400 flex items-center justify-center shrink-0 mt-0.5"
            >
              <CheckCircle2 class="h-3.5 w-3.5" />
            </div>
            <div>
              <p class="text-xs font-semibold text-white">Real-Time Operational Triage</p>
              <p class="text-[11px] text-slate-400">
                Live queue status distribution across Open, In Progress, and Resolved states.
              </p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div
              class="h-6 w-6 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center shrink-0 mt-0.5"
            >
              <ShieldCheck class="h-3.5 w-3.5" />
            </div>
            <div>
              <p class="text-xs font-semibold text-white">Strict RBAC & Staff Notes</p>
              <p class="text-[11px] text-slate-400">
                End-to-end security separating customer conversations from confidential internal
                notes.
              </p>
            </div>
          </div>

          <div class="flex items-start gap-3">
            <div
              class="h-6 w-6 rounded-lg bg-sky-500/10 border border-sky-500/20 text-sky-400 flex items-center justify-center shrink-0 mt-0.5"
            >
              <Sparkles class="h-3.5 w-3.5" />
            </div>
            <div>
              <p class="text-xs font-semibold text-white">Enterprise SLA Compliance</p>
              <p class="text-[11px] text-slate-400">
                Full audit trails with timestamped transitions and assignee ownership.
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom Trust Badge -->
      <div class="relative z-10 pt-4 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
        <span class="flex items-center gap-1.5 font-medium text-slate-300">
          <span class="h-2 w-2 rounded-full bg-emerald-400"></span>
          99.98% SLA Availability
        </span>
        <span class="text-[11px] text-slate-500">ISO 27001 Certified Security</span>
      </div>
    </div>

    <!-- Right Content / Form Panel -->
    <div
      class="auth-form-panel flex-1 flex flex-col justify-between px-5 py-4 sm:px-8 lg:h-full lg:min-h-0 lg:px-8 lg:py-4 xl:px-12 overflow-y-auto min-h-[100svh] bg-slate-50"
    >
      <!-- Mobile Logo Header (Only visible < lg) -->
      <div class="lg:hidden flex items-center justify-between pb-4 border-b border-slate-200">
        <router-link to="/login" class="inline-flex items-center gap-2.5">
          <div
            class="h-9 w-9 rounded-xl bg-brand-600 flex items-center justify-center text-white shadow-md shadow-brand-500/20"
          >
            <Ticket class="h-5 w-5" />
          </div>
          <div>
            <span class="text-base font-bold tracking-tight text-slate-900 leading-none">Helpdesk</span>
            <span class="text-[10px] font-bold uppercase ml-1.5 px-1.5 py-0.2 bg-brand-100 text-brand-700 rounded">
              Enterprise
            </span>
          </div>
        </router-link>
        <span class="text-xs text-slate-500 font-medium">SaaS Platform</span>
      </div>

      <!-- Main Form Wrapper Container -->
      <div class="auth-card-wrap my-auto w-full max-w-md mx-auto py-2 lg:py-1">
        <div class="auth-card bg-white rounded-2xl border border-slate-200/90 shadow-card p-5 sm:p-6">
          <slot>
            <router-view />
          </slot>
        </div>
      </div>

      <!-- Footer Assurance -->
      <div class="auth-panel-footer pt-2 text-center space-y-1 text-xs text-slate-500">
        <p class="flex items-center justify-center gap-1.5 text-[11px]">
          <ShieldCheck class="h-3.5 w-3.5 text-emerald-600" />
          <span>Encrypted with Enterprise JWT and Role-Based Authorization</span>
        </p>
        <p class="text-[11px] text-slate-400">
          &copy; 2026 PT Simple Journey Indonesia. All rights reserved.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Ticket, ShieldCheck, CheckCircle2, Sparkles } from 'lucide-vue-next'
</script>

<style scoped>
@media (min-width: 1024px) and (max-height: 820px) {
  .auth-shell {
    height: 100svh;
  }

  .auth-brand-panel {
    padding-block: 1.25rem;
    padding-inline: 2rem;
  }

  .auth-brand-content {
    padding-block: 1rem;
    row-gap: 1rem;
  }

  .auth-form-panel {
    padding-block: 0.75rem;
  }

  .auth-card-wrap {
    padding-block: 0;
  }

  .auth-card {
    padding: 1.25rem;
  }

  .auth-panel-footer {
    padding-top: 0.375rem;
  }

  :deep(.auth-page-heading) {
    margin-bottom: 0.875rem;
  }

  :deep(.auth-form) {
    row-gap: 0.625rem;
  }

  :deep(.auth-form > :not([hidden]) ~ :not([hidden])) {
    margin-top: 0.625rem;
  }

  :deep(.auth-demo) {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
  }

  :deep(.auth-policy) {
    padding: 0.5rem 0.75rem;
  }
}

@media (max-width: 1023px) {
  .auth-form-panel {
    min-height: 100svh;
  }
}
</style>
