<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const activeYear = new Date().getFullYear()

const loading = ref(true)
const refreshing = ref(false)
const errorMessage = ref('')
const standings = ref(null)
const refreshedAt = ref('')

let refreshTimerId

const entries = computed(() => standings.value?.entries ?? [])
const updatedAtLabel = computed(() => {
  if (!standings.value?.updated_at) {
    return 'No live result snapshot has been recorded yet.'
  }

  return new Date(standings.value.updated_at).toLocaleString()
})

const leader = computed(() => entries.value[0] ?? null)

onMounted(async () => {
  await loadStandings()

  refreshTimerId = window.setInterval(() => {
    void loadStandings({ silent: true })
  }, 5 * 60 * 1000)
})

onBeforeUnmount(() => {
  if (refreshTimerId) {
    window.clearInterval(refreshTimerId)
  }
})

async function loadStandings(options = {}) {
  const { silent = false } = options

  if (silent) {
    refreshing.value = true
  } else {
    loading.value = true
  }

  errorMessage.value = ''

  try {
    const response = await fetch(`/api/standings/${activeYear}`)
    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load standings.'))
    }

    standings.value = await response.json()
    refreshedAt.value = new Date().toLocaleTimeString()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading standings.'
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function formatMoney(value) {
  return new Intl.NumberFormat('en-CA', {
    style: 'currency',
    currency: 'CAD',
    maximumFractionDigits: 0,
  }).format(Number(value ?? 0))
}

function payoutLabel(pick) {
  if (pick.multiplier > 1) {
    return `${formatMoney(pick.base_payout)} x ${pick.multiplier} = ${formatMoney(pick.total_payout)}`
  }

  return formatMoney(pick.total_payout)
}

async function responseMessage(response, fallback) {
  const text = (await response.text()).trim()
  return text || fallback
}
</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Live Standings</p>
      <h2>Projected pool leaderboard for {{ activeYear }}</h2>
      <p>
        This page now reads the real standings endpoint. Totals are built from
        projected tournament winnings, tie-split payouts, mutt multipliers, and
        any recorded first-round leader bonus.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Leaderboard State</p>
      <p class="status-value">
        {{ leader ? `${leader.display_name} leads` : 'Awaiting result snapshot' }}
      </p>
      <p class="status-meta">
        {{ updatedAtLabel }}
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Standings Feed</p>
        <h3>{{ entries.length }} entries ranked</h3>
      </div>
      <button class="ghost-button" type="button" :disabled="loading || refreshing" @click="loadStandings()">
        {{ refreshing ? 'Refreshing…' : 'Refresh' }}
      </button>
    </div>

    <div v-if="loading" class="empty-state">
      <p>Loading standings...</p>
    </div>

    <div v-else-if="errorMessage" class="alert alert-error">
      <p>{{ errorMessage }}</p>
    </div>

    <div v-else-if="entries.length === 0" class="empty-state">
      <p>
        No ranked entries are available yet. Save entries and add a golfer
        results snapshot through the backend before expecting standings here.
      </p>
    </div>

    <div v-else class="entries-grid standings-grid">
      <article v-for="entry in entries" :key="entry.entry_id" class="group-card entry-card standings-card">
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
    </div>

    <p v-if="refreshedAt" class="helper-copy standings-refresh-meta">
      Page refreshed at {{ refreshedAt }}. Automatic refresh runs every 5 minutes.
    </p>
  </section>
</template>
