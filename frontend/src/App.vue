<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Show, UserButton } from '@clerk/vue'

import ClerkBridge from './components/ClerkBridge.vue'
import { authMode, currentBackendUser, isClerkEnabled, isSignedIn } from './lib/auth'

const route = useRoute()

const authBanner = computed(() => {
  if (!isClerkEnabled.value) {
    return 'Clerk frontend auth is not configured yet, so the app is still in mock-auth-compatible mode.'
  }

  if (route.query.auth === 'admin-required') {
    return 'That route requires an admin session.'
  }

  if (route.query.auth === 'session-error') {
    return 'The browser session loaded, but the API user check failed.'
  }

  return ''
})
</script>

<template>
  <div class="app-shell">
    <ClerkBridge v-if="isClerkEnabled" />

    <header class="topbar">
      <div>
        <p class="eyebrow">Masters Pool</p>
        <h1>Masters Pool Rewrite</h1>
      </div>
      <p class="topbar-copy">
        Learning-focused rebuild of the pool app with a real Go API, a Vue
        frontend, and a cleaner step-by-step architecture.
      </p>
    </header>

    <div class="toolbar">
      <nav class="main-nav" aria-label="Primary">
        <RouterLink to="/" class="nav-link">Home</RouterLink>
        <RouterLink to="/enter" class="nav-link">Enter</RouterLink>
        <RouterLink to="/entries" class="nav-link">Entries</RouterLink>
        <RouterLink to="/admin" class="nav-link">Admin</RouterLink>
        <RouterLink to="/standings" class="nav-link">Standings</RouterLink>
        <RouterLink v-if="authMode === 'clerk' && !isSignedIn" to="/sign-in" class="nav-link">Sign In</RouterLink>
      </nav>

      <div class="auth-controls">
        <Show v-if="isClerkEnabled" when="signed-in">
          <div class="auth-chip">
            <span>{{ currentBackendUser?.user?.email ?? 'Signed in' }}</span>
            <UserButton :show-name="true" />
          </div>
        </Show>
        <Show v-if="isClerkEnabled" when="signed-out">
          <RouterLink to="/sign-in" class="nav-link nav-link-ghost">Sign In</RouterLink>
        </Show>
        <span v-if="!isClerkEnabled" class="badge badge-new">Mock auth compatible</span>
      </div>
    </div>

    <div v-if="authBanner" class="alert alert-error app-alert">
      <p>{{ authBanner }}</p>
    </div>

    <main class="page-grid">
      <RouterView />
    </main>
  </div>
</template>
