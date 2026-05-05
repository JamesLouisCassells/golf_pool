<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'

import { apiFetch, responseMessage } from '../lib/api'

const activeYear = new Date().getFullYear()

const loading = ref(true)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const config = ref(null)
const existingEntry = ref(null)
const now = ref(Date.now())

const form = reactive({
  displayName: '',
  inOvers: false,
  picks: {},
})

let clockId

onMounted(async () => {
  clockId = window.setInterval(() => {
    now.value = Date.now()
  }, 1000)

  await loadPage()
})

onBeforeUnmount(() => {
  if (clockId) {
    window.clearInterval(clockId)
  }
})

const deadline = computed(() => {
  if (!config.value?.entry_deadline) {
    return null
  }

  return new Date(config.value.entry_deadline)
})

const isLocked = computed(() => {
  if (!deadline.value) {
    return false
  }

  return deadline.value.getTime() <= now.value
})

const countdown = computed(() => {
  if (!deadline.value) {
    return 'Deadline will appear once tournament config is seeded.'
  }

  const difference = deadline.value.getTime() - now.value
  if (difference <= 0) {
    return 'Entry window is closed.'
  }

  const totalSeconds = Math.floor(difference / 1000)
  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  return `${days}d ${hours}h ${minutes}m ${seconds}s remaining`
})

const groupSections = computed(() => normalizeGroups(config.value?.groups ?? {}))

const submitLabel = computed(() => {
  if (saving.value) {
    return existingEntry.value ? 'Saving Changes...' : 'Submitting Entry...'
  }

  return existingEntry.value ? 'Save Entry Changes' : 'Submit Entry'
})

async function loadPage() {
  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const configResponse = await apiFetch(`/api/config/${activeYear}`)
    if (!configResponse.ok) {
      throw new Error(await responseMessage(configResponse, 'Failed to load tournament config.'))
    }

    const configPayload = await configResponse.json()
    config.value = configPayload

    initializeEmptyPicks(configPayload.groups)

    const entryResponse = await apiFetch('/api/entries/mine')
    if (entryResponse.status === 404) {
      existingEntry.value = null
      form.displayName = ''
      form.inOvers = false
      return
    }

    if (!entryResponse.ok) {
      throw new Error(await responseMessage(entryResponse, 'Failed to load your entry.'))
    }

    const entryPayload = await entryResponse.json()
    existingEntry.value = entryPayload
    hydrateFormFromEntry(entryPayload)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading the page.'
  } finally {
    loading.value = false
  }
}

function initializeEmptyPicks(groups) {
  const nextPicks = {}
  for (const section of normalizeGroups(groups ?? {})) {
    nextPicks[section.name] = ''
  }

  replacePicks(nextPicks)
}

function hydrateFormFromEntry(entry) {
  form.displayName = entry.display_name ?? ''
  form.inOvers = Boolean(entry.in_overs)

  const nextPicks = {}
  for (const section of groupSections.value) {
    nextPicks[section.name] = entry.picks?.[section.name] ?? ''
  }

  replacePicks(nextPicks)
}

function replacePicks(nextPicks) {
  for (const key of Object.keys(form.picks)) {
    delete form.picks[key]
  }

  for (const [key, value] of Object.entries(nextPicks)) {
    form.picks[key] = value
  }
}

async function submitEntry() {
  if (saving.value || isLocked.value) {
    return
  }

  errorMessage.value = ''
  successMessage.value = ''

  const payload = {
    display_name: form.displayName.trim(),
    picks: { ...form.picks },
    in_overs: form.inOvers,
  }

  if (!allGroupsSelected(payload.picks)) {
    errorMessage.value = 'Every group needs a golfer selection before you can save.'
    return
  }

  saving.value = true

  try {
    const target = existingEntry.value ? `/api/entries/${existingEntry.value.id}` : '/api/entries'
    const method = existingEntry.value ? 'PUT' : 'POST'

    const response = await apiFetch(target, {
      method,
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to save your entry.'))
    }

    const savedEntry = await response.json()
    existingEntry.value = savedEntry
    hydrateFormFromEntry(savedEntry)
    successMessage.value = existingEntry.value ? 'Entry saved.' : 'Entry submitted.'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while saving your entry.'
  } finally {
    saving.value = false
  }
}

function allGroupsSelected(picks) {
  return Object.values(picks).every((value) => String(value).trim() !== '')
}

function normalizeGroups(groups) {
  return Object.entries(groups).map(([name, rawGroup]) => ({
    name,
    options: extractPlayers(rawGroup),
  }))
}

function extractPlayers(rawGroup) {
  if (Array.isArray(rawGroup)) {
    return rawGroup.map(normalizePlayerOption).filter(Boolean)
  }

  if (rawGroup && typeof rawGroup === 'object') {
    if (Array.isArray(rawGroup.players)) {
      return rawGroup.players.map(normalizePlayerOption).filter(Boolean)
    }

    if (Array.isArray(rawGroup.options)) {
      return rawGroup.options.map(normalizePlayerOption).filter(Boolean)
    }
  }

  return []
}

function normalizePlayerOption(player) {
  if (typeof player === 'string') {
    return {
      value: player,
      label: player,
    }
  }

  if (player && typeof player === 'object') {
    const label = player.name ?? player.player ?? player.label ?? player.value
    if (!label) {
      return null
    }

    return {
      value: String(label),
      label: String(label),
    }
  }

  return null
}

</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Active Year {{ activeYear }}</p>
      <h2>Build your Masters ticket</h2>
      <p>
        This view is talking to the real Go API. It loads the tournament config,
        checks for your existing entry, and saves through the same create and
        edit endpoints the final app will use.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Entry Window</p>
      <p class="status-value" :class="{ closed: isLocked }">
        {{ countdown }}
      </p>
      <p class="status-meta">
        {{ deadline ? deadline.toLocaleString() : 'No deadline configured yet.' }}
      </p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Entry Flow</p>
        <h3>{{ existingEntry ? 'Edit Existing Entry' : 'Create New Entry' }}</h3>
      </div>
      <span class="badge" :class="existingEntry ? 'badge-existing' : 'badge-new'">
        {{ existingEntry ? 'Existing entry loaded' : 'No entry on file yet' }}
      </span>
    </div>

    <div v-if="loading" class="empty-state">
      <p>Loading config and entry data...</p>
    </div>

    <div v-else-if="errorMessage" class="alert alert-error">
      <p>{{ errorMessage }}</p>
    </div>

    <form v-else class="entry-form" @submit.prevent="submitEntry">
      <div v-if="successMessage" class="alert alert-success">
        <p>{{ successMessage }}</p>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Display name</span>
          <input
            v-model="form.displayName"
            :disabled="saving || isLocked"
            type="text"
            name="display-name"
            placeholder="How your entry should appear"
          />
        </label>

        <label class="checkbox-field">
          <input
            v-model="form.inOvers"
            :disabled="saving || isLocked"
            type="checkbox"
            name="in-overs"
          />
          <span>Mark this entry as in overs</span>
        </label>
      </div>

      <div class="groups-grid">
        <label
          v-for="group in groupSections"
          :key="group.name"
          class="field group-card"
        >
          <span>{{ group.name }}</span>
          <select
            v-model="form.picks[group.name]"
            :disabled="saving || isLocked"
            :name="group.name"
          >
            <option disabled value="">Choose a golfer</option>
            <option
              v-for="option in group.options"
              :key="`${group.name}-${option.value}`"
              :value="option.value"
            >
              {{ option.label }}
            </option>
          </select>
        </label>
      </div>

      <div class="form-footer">
        <p class="helper-copy">
          {{ isLocked
            ? 'The deadline has passed, so changes are locked.'
            : 'Selections are saved against the active tournament year in the backend.' }}
        </p>

        <button
          class="submit-button"
          :disabled="saving || isLocked || loading"
          type="submit"
        >
          {{ submitLabel }}
        </button>
      </div>
    </form>
  </section>
</template>
