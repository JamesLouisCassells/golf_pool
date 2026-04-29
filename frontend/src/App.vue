<script setup>
import EnterView from './views/Enter.vue'
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <div>
        <p class="eyebrow">Masters Pool</p>
        <h1>Entry Builder</h1>
      </div>
      <p class="topbar-copy">
        First frontend integration against the live backend entry workflow.
      </p>
    </header>

    <main class="page-grid">
      <EnterView />
    </main>
  </div>
</template>
