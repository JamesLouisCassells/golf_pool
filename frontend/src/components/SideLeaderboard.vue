<script setup>
defineProps({
  entries: {
    type: Array,
    required: true,
  },
  updatedAtLabel: {
    type: String,
    required: true,
  },
  resultCount: {
    type: Number,
    required: true,
  },
  formatMoney: {
    type: Function,
    required: true,
  },
})
</script>

<template>
  <aside class="side-leaderboard">
    <div class="side-leaderboard-header">
      <p class="kicker">Leaderboard Pulse</p>
      <h4>Top field positions</h4>
    </div>

    <ol class="side-leaderboard-list">
      <li v-for="entry in entries.slice(0, 5)" :key="entry.entry_id" class="side-leaderboard-row">
        <div>
          <p class="side-leaderboard-rank">#{{ entry.rank }}</p>
          <strong>{{ entry.display_name }}</strong>
        </div>
        <span>{{ formatMoney(entry.total_payout) }}</span>
      </li>
    </ol>

    <dl class="side-leaderboard-meta">
      <div>
        <dt>Live golfers</dt>
        <dd>{{ resultCount }}</dd>
      </div>
      <div>
        <dt>Last update</dt>
        <dd>{{ updatedAtLabel }}</dd>
      </div>
    </dl>
  </aside>
</template>
