<script setup>
defineProps({
  entry: {
    type: Object,
    required: true,
  },
  formatMoney: {
    type: Function,
    required: true,
  },
  payoutLabel: {
    type: Function,
    required: true,
  },
})
</script>

<template>
  <article class="group-card entry-card standings-card">
    <div class="entry-card-head">
      <div>
        <p class="card-step">Rank {{ entry.rank }}</p>
        <h4>{{ entry.display_name }}</h4>
      </div>
      <div class="standings-card-total">
        <span class="badge" :class="entry.in_overs ? 'badge-new' : 'badge-existing'">
          {{ entry.in_overs ? 'In overs' : 'Standard' }}
        </span>
        <p>{{ formatMoney(entry.total_payout) }}</p>
      </div>
    </div>

    <div v-if="entry.frl_bonus" class="alert alert-success standings-frl">
      <p>FRL bonus applied: {{ formatMoney(entry.frl_bonus) }}</p>
    </div>

    <dl class="pick-list standings-pick-list">
      <div v-for="pick in entry.picks" :key="`${entry.entry_id}-${pick.group_name}`">
        <dt>
          <span>{{ pick.group_name }}</span>
          <strong>{{ pick.golfer_name }}</strong>
        </dt>
        <dd>
          <span>{{ pick.position || 'No live position yet' }}</span>
          <span>{{ payoutLabel(pick) }}</span>
        </dd>
      </div>
    </dl>
  </article>
</template>
