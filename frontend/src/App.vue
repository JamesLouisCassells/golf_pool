<template>
  <div class="app-shell">
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

    <nav class="main-nav" aria-label="Primary">
      <RouterLink to="/" class="nav-link">Home</RouterLink>
      <RouterLink to="/enter" class="nav-link">Enter</RouterLink>
      <RouterLink to="/standings" class="nav-link">Standings</RouterLink>
    </nav>

    <main class="page-grid">
      <RouterView />
    </main>
  </div>
</template>
