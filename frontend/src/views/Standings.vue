<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Standings Route</p>
      <h2>Frontend placeholder, backend not wired yet</h2>
      <p>
        This route exists now so the app structure is real, but the standings
        page will stay lightweight until the backend standings endpoint and golf
        scoring logic are ready.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Dependency</p>
      <p class="status-value">`GET /api/standings/:year`</p>
      <p class="status-meta">
        Once that endpoint exists, this route becomes the next clean data
        integration slice.
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="empty-state">
      <p>
        Standings UI is intentionally deferred until payout logic and live golf
        data exist in the backend. That keeps the page shape honest instead of
        mocking a leaderboard too early.
      </p>
    </div>
  </section>
</template>
