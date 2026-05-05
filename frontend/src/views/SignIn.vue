<script setup>
import { computed } from 'vue'
import { SignIn } from '@clerk/vue'

import { isClerkEnabled } from '../lib/auth'

const isConfigured = computed(() => isClerkEnabled.value)
</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Authentication</p>
      <h2>Sign in to manage your entry</h2>
      <p>
        Clerk is the real browser auth flow for this app. Use it to reach
        protected entry and admin routes without relying on mock auth.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Current Mode</p>
      <p class="status-value">{{ isConfigured ? 'Clerk enabled' : 'Clerk not configured' }}</p>
      <p class="status-meta">
        {{ isConfigured
          ? 'This page is live once your publishable key is present.'
          : 'Add a Clerk publishable key to enable the real browser sign-in flow.' }}
      </p>
    </div>
  </section>

  <section class="panel">
    <div v-if="!isConfigured" class="empty-state">
      <p>
        Clerk is not configured in the frontend yet. Add
        `VITE_CLERK_PUBLISHABLE_KEY` to the frontend environment to enable this
        screen.
      </p>
    </div>

    <div v-else class="clerk-panel">
      <SignIn />
    </div>
  </section>
</template>
