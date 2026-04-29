<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Current Frontend Phase</p>
      <h2>App shell is in place</h2>
      <p>
        The project now has a real navigation spine. Entry creation and editing
        live behind the <strong>/enter</strong> route, and the remaining pages
        can be added without another frontend rewrite.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">What Works</p>
      <p class="status-value">Router + entry workflow</p>
      <p class="status-meta">
        This is the point where the app stops being a single mounted demo page
        and starts behaving like a real frontend.
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Next Slices</p>
        <h3>Where the frontend goes from here</h3>
      </div>
      <span class="badge badge-new">Intentional placeholders</span>
    </div>

    <div class="home-grid">
      <article class="group-card">
        <p class="card-step">1</p>
        <h4>Standings integration</h4>
        <p>
          Wire the future standings route to live data once the backend endpoint
          exists.
        </p>
      </article>

      <article class="group-card">
        <p class="card-step">2</p>
        <h4>Browser auth state</h4>
        <p>
          Replace mock assumptions with real Clerk state when the auth frontend
          work is ready.
        </p>
      </article>

      <article class="group-card">
        <p class="card-step">3</p>
        <h4>Shared page structure</h4>
        <p>
          Add entries, admin, and navigation polish on top of the router spine
          instead of hardwiring more pages into the root component.
        </p>
      </article>
    </div>
  </section>
</template>
