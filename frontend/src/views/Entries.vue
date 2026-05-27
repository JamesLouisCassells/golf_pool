<script setup>
import { computed, onMounted, ref } from 'vue'

const loading = ref(true)
const errorMessage = ref('')
const entries = ref([])
const entriesVisible = ref(true)
const inOversCount = computed(() => entries.value.filter((entry) => entry.in_overs).length)

const helperTitle = computed(() => {
  if (!entriesVisible.value) {
    return 'Entries stay private until the tournament starts'
  }

  return `${entries.length} submissions loaded`
})

onMounted(async () => {
  await loadEntries()
})

async function loadEntries() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch('/api/entries')
    if (response.status === 403) {
      entriesVisible.value = false
      errorMessage.value = ''
      return
    }

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load entries.'))
    }

    entriesVisible.value = true
    entries.value = await response.json()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading entries.'
  } finally {
    loading.value = false
  }
}

async function responseMessage(response, fallback) {
  const text = (await response.text()).trim()
  return text || fallback
}
</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Entries Route</p>
      <h2>Public picks after tournament start</h2>
      <p>
        This page consumes the new public entries endpoint. It intentionally
        stays simple for now: once the tournament has started, the community can
        see who submitted and what each ticket looks like.
      </p>

      <div v-if="entriesVisible && entries.length" class="standings-hero-stats">
        <div class="group-card standings-stat-card">
          <span class="status-label">Entries</span>
          <strong>{{ entries.length }}</strong>
        </div>
        <div class="group-card standings-stat-card">
          <span class="status-label">In overs</span>
          <strong>{{ inOversCount }}</strong>
        </div>
        <div class="group-card standings-stat-card">
          <span class="status-label">Standard</span>
          <strong>{{ entries.length - inOversCount }}</strong>
        </div>
      </div>
    </div>

    <div class="status-card">
      <p class="status-label">Data Source</p>
      <p class="status-value">`GET /api/entries`</p>
      <p class="status-meta">
        {{ entriesVisible ? 'The backend reveals this once the tournament starts.' : 'The backend is still holding the field private.' }}
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Active Year Entries</p>
        <h3>{{ helperTitle }}</h3>
      </div>
      <button class="ghost-button" type="button" :disabled="loading" @click="loadEntries">
        Reload
      </button>
    </div>

    <div v-if="loading" class="empty-state">
      <p>Loading entries...</p>
    </div>

    <div v-else-if="errorMessage" class="alert alert-error">
      <p>{{ errorMessage }}</p>
      <div class="inline-actions">
        <button class="ghost-button" type="button" @click="loadEntries">Try again</button>
      </div>
    </div>

    <div v-else-if="!entriesVisible" class="empty-state">
      <p>The field is still private because the active tournament has not started yet.</p>
    </div>

    <div v-else-if="entries.length === 0" class="empty-state">
      <p>No entries are available yet.</p>
    </div>

    <div v-else class="entries-grid">
      <article v-for="entry in entries" :key="entry.id" class="group-card entry-card">
        <div class="entry-card-head">
          <div>
            <p class="card-step">{{ entry.display_name }}</p>
            <h4>Entry {{ entry.id }}</h4>
          </div>
          <span class="badge" :class="entry.in_overs ? 'badge-new' : 'badge-existing'">
            {{ entry.in_overs ? 'In overs' : 'Standard' }}
          </span>
        </div>

        <dl class="pick-list">
          <div v-for="(pick, groupName) in entry.picks" :key="`${entry.id}-${groupName}`">
            <dt>{{ groupName }}</dt>
            <dd>{{ pick }}</dd>
          </div>
        </dl>
      </article>
    </div>
  </section>
</template>
