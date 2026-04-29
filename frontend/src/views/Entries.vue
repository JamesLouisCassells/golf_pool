<script setup>
import { onMounted, ref } from 'vue'

const loading = ref(true)
const errorMessage = ref('')
const entries = ref([])

onMounted(async () => {
  await loadEntries()
})

async function loadEntries() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await fetch('/api/entries')
    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load entries.'))
    }

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
    </div>

    <div class="status-card">
      <p class="status-label">Data Source</p>
      <p class="status-value">`GET /api/entries`</p>
      <p class="status-meta">
        The backend keeps this hidden until the active tournament start date.
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Active Year Entries</p>
        <h3>{{ entries.length }} submissions loaded</h3>
      </div>
    </div>

    <div v-if="loading" class="empty-state">
      <p>Loading entries...</p>
    </div>

    <div v-else-if="errorMessage" class="alert alert-error">
      <p>{{ errorMessage }}</p>
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
