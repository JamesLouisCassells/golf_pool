<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Current Frontend State</p>
      <h2>The pool is running through real routes now</h2>
      <p>
        The app now runs on the real Go API with Clerk-backed protected routes,
        live standings, admin tournament controls, and a routed Vue shell that
        behaves like an actual product instead of a scaffold.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Ready To Use</p>
      <p class="status-value">Entries, standings, and admin flow</p>
      <p class="status-meta">
        The main gap now is polish and deployment follow-through, not core app
        structure.
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Working Areas</p>
        <h3>What each route is for right now</h3>
      </div>
      <span class="badge badge-existing">Live frontend surface</span>
    </div>

    <div class="home-grid">
      <article class="group-card">
        <p class="card-step">1</p>
        <h4>Standings</h4>
        <p>
          Read the live leaderboard feed with expandable entry detail and
          side-panel summary.
        </p>
      </article>

      <article class="group-card">
        <p class="card-step">2</p>
        <h4>Entry flow</h4>
        <p>
          Create or edit a personal ticket through the real protected API route
          before the deadline closes.
        </p>
      </article>

      <article class="group-card">
        <p class="card-step">3</p>
        <h4>Admin controls</h4>
        <p>
          Operate the active tournament year, standings refresh, entry cleanup,
          and locking controls.
        </p>
      </article>
    </div>
  </section>
</template>
