<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import EntryCard from '../components/EntryCard.vue'
import SideLeaderboard from '../components/SideLeaderboard.vue'
import { fetchActiveConfig } from '../lib/tournament'

const activeYear = ref('')

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
const topEntries = computed(() => entries.value.slice(0, 5))
const totalProjectedPayout = computed(() =>
  entries.value.reduce((total, entry) => total + Number(entry.total_payout ?? 0), 0),
)
const averageProjectedPayout = computed(() => {
  if (entries.value.length === 0) {
    return 0
  }

  return Math.round(totalProjectedPayout.value / entries.value.length)
})

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
    if (!activeYear.value) {
      const config = await fetchActiveConfig()
      activeYear.value = config.year ?? ''
    }

    const response = await fetch(`/api/standings/${activeYear.value}`)
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
      <h2>Projected pool leaderboard for {{ activeYear || 'the active year' }}</h2>
      <p>
        This page now reads the real standings endpoint. Totals are built from
        projected tournament winnings, tie-split payouts, mutt multipliers, and
        any recorded first-round leader bonus.
      </p>

      <div v-if="entries.length" class="standings-hero-stats">
        <div class="group-card standings-stat-card">
          <span class="status-label">Entries</span>
          <strong>{{ entries.length }}</strong>
        </div>
        <div class="group-card standings-stat-card">
          <span class="status-label">Results</span>
          <strong>{{ standings?.result_count ?? 0 }}</strong>
        </div>
        <div class="group-card standings-stat-card">
          <span class="status-label">Average Projected</span>
          <strong>{{ formatMoney(averageProjectedPayout) }}</strong>
        </div>
      </div>
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
      <div class="inline-actions">
        <button class="ghost-button" type="button" @click="loadStandings()">Try again</button>
      </div>
    </div>

    <div v-else-if="entries.length === 0" class="empty-state">
      <p>
        No ranked entries are available yet. Save entries and add a golfer
        results snapshot through the backend before expecting standings here.
      </p>
    </div>

    <div v-else class="standings-layout">
      <SideLeaderboard
        :entries="topEntries"
        :updated-at-label="updatedAtLabel"
        :result-count="standings?.result_count ?? 0"
        :format-money="formatMoney"
      />

      <div class="entries-grid standings-grid">
        <EntryCard
          v-for="entry in entries"
          :key="entry.entry_id"
          :entry="entry"
          :auto-expand="entry.rank === 1"
          :format-money="formatMoney"
          :payout-label="payoutLabel"
        />
      </div>
    </div>

    <p v-if="refreshedAt" class="helper-copy standings-refresh-meta">
      Page refreshed at {{ refreshedAt }}. Automatic refresh runs every 5 minutes.
    </p>
  </section>
</template>
