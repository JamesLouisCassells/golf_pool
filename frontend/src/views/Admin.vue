<script setup>
import { computed, reactive, ref } from 'vue'

import { apiFetch, responseMessage } from '../lib/api'

const activeYear = new Date().getFullYear()

const configLoading = ref(false)
const configSaving = ref(false)
const configErrorMessage = ref('')
const configSuccessMessage = ref('')
const operationsLoading = ref(false)
const operationsErrorMessage = ref('')
const operationsSuccessMessage = ref('')

const entriesLoading = ref(false)
const entriesErrorMessage = ref('')
const entriesSuccessMessage = ref('')
const editingEntryId = ref('')
const savingEntryId = ref('')
const deletingEntryId = ref('')
const entryDrafts = ref({})
const entries = ref([])

const form = reactive({
  entry_deadline: '',
  start_date: '',
  end_date: '',
  provider_tournament_id: '',
  mutt_multiplier: '2',
  old_mutt_multiplier: '3',
  frl_winner: '',
  frl_payout: 500000,
  active: false,
  groups_json: '{\n  "Group 1": []\n}',
  pool_payouts_json: '{\n  "1": 4475\n}',
})

const operationsForm = reactive({
  refresh_year: activeYear,
  tournament_id: '',
  round_id: '',
  results_json:
    '[\n' +
    '  {\n' +
    '    "golfer_name": "Scottie Scheffler",\n' +
    '    "position": "T1",\n' +
    '    "score": "-12",\n' +
    '    "today": "-4",\n' +
    '    "thru": "F"\n' +
    '  }\n' +
    ']',
})

const helperMessage = computed(() =>
  'This page now operates tournament config plus active-year entry cleanup while richer admin tooling is still pending.',
)

loadPage()

function loadPage() {
  void loadConfig()
  void loadEntries()
}

async function loadConfig() {
  configLoading.value = true
  configErrorMessage.value = ''
  configSuccessMessage.value = ''

  try {
    const response = await apiFetch(`/api/admin/config/${activeYear}`)
    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load admin config.'))
    }

    const config = await response.json()
    form.entry_deadline = toDateTimeLocalValue(config.entry_deadline)
    form.start_date = toDateInputValue(config.start_date)
    form.end_date = toDateInputValue(config.end_date)
    form.provider_tournament_id = config.provider_tournament_id ?? ''
    form.mutt_multiplier = config.mutt_multiplier ?? '2'
    form.old_mutt_multiplier = config.old_mutt_multiplier ?? '3'
    form.frl_winner = config.frl_winner ?? ''
    form.frl_payout = config.frl_payout ?? 500000
    form.active = Boolean(config.active)
    form.groups_json = prettyJSON(config.groups ?? {})
    form.pool_payouts_json = prettyJSON(config.pool_payouts ?? {})
    if (!String(operationsForm.tournament_id).trim() && config.provider_tournament_id) {
      operationsForm.tournament_id = config.provider_tournament_id
    }
  } catch (error) {
    configErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading admin config.'
  } finally {
    configLoading.value = false
  }
}

async function saveConfig() {
  if (configSaving.value) {
    return
  }

  configErrorMessage.value = ''
  configSuccessMessage.value = ''

  let groups
  let poolPayouts

  try {
    groups = JSON.parse(form.groups_json)
    poolPayouts = JSON.parse(form.pool_payouts_json)
  } catch {
    configErrorMessage.value = 'Groups and pool payouts must be valid JSON.'
    return
  }

  const payload = {
    entry_deadline: form.entry_deadline ? new Date(form.entry_deadline).toISOString() : null,
    start_date: form.start_date ? new Date(`${form.start_date}T00:00:00`).toISOString() : null,
    end_date: form.end_date ? new Date(`${form.end_date}T00:00:00`).toISOString() : null,
    provider_tournament_id: form.provider_tournament_id.trim() || null,
    groups,
    mutt_multiplier: form.mutt_multiplier.trim(),
    old_mutt_multiplier: form.old_mutt_multiplier.trim(),
    pool_payouts: poolPayouts,
    frl_winner: form.frl_winner.trim() || null,
    frl_payout: Number(form.frl_payout),
    active: form.active,
  }

  configSaving.value = true

  try {
    const response = await apiFetch(`/api/admin/config/${activeYear}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to save admin config.'))
    }

    const updated = await response.json()
    form.entry_deadline = toDateTimeLocalValue(updated.entry_deadline)
    form.start_date = toDateInputValue(updated.start_date)
    form.end_date = toDateInputValue(updated.end_date)
    form.provider_tournament_id = updated.provider_tournament_id ?? ''
    form.groups_json = prettyJSON(updated.groups ?? {})
    form.pool_payouts_json = prettyJSON(updated.pool_payouts ?? {})
    if (updated.provider_tournament_id) {
      operationsForm.tournament_id = updated.provider_tournament_id
    }
    configSuccessMessage.value = 'Tournament config saved.'
  } catch (error) {
    configErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while saving admin config.'
  } finally {
    configSaving.value = false
  }
}

async function refreshResults() {
  if (operationsLoading.value) {
    return
  }

  operationsErrorMessage.value = ''
  operationsSuccessMessage.value = ''

  let results
  try {
    results = JSON.parse(operationsForm.results_json)
  } catch {
    operationsErrorMessage.value = 'Results JSON must be valid before refreshing standings.'
    return
  }

  if (!Array.isArray(results)) {
    operationsErrorMessage.value = 'Results JSON must be an array of golfer result objects.'
    return
  }

  operationsLoading.value = true

  try {
    const response = await apiFetch('/api/admin/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        year: Number(operationsForm.refresh_year),
        results,
      }),
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to refresh golfer results.'))
    }

    const payload = await response.json()
    operationsSuccessMessage.value = `Stored ${payload.result_count} golfer results for ${payload.year}.`
  } catch (error) {
    operationsErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while refreshing golfer results.'
  } finally {
    operationsLoading.value = false
  }
}

async function fetchProviderResults() {
  if (operationsLoading.value) {
    return
  }

  operationsErrorMessage.value = ''
  operationsSuccessMessage.value = ''

  const tournamentID = String(operationsForm.tournament_id).trim()

  const roundIDValue = String(operationsForm.round_id).trim()
  const roundID = roundIDValue === '' ? null : Number(roundIDValue)
  if (roundIDValue !== '' && (!Number.isInteger(roundID) || roundID <= 0)) {
    operationsErrorMessage.value = 'Round ID must be a positive number when provided.'
    return
  }

  operationsLoading.value = true

  try {
    const response = await apiFetch('/api/admin/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        year: Number(operationsForm.refresh_year),
        tournament_id: tournamentID || null,
        round_id: roundID,
      }),
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to fetch golfer results from the provider.'))
    }

    const payload = await response.json()
    operationsSuccessMessage.value = `Fetched ${payload.result_count} golfer results for ${payload.year} from the provider.`
  } catch (error) {
    operationsErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while fetching provider results.'
  } finally {
    operationsLoading.value = false
  }
}

async function lockEntriesNow() {
  if (operationsLoading.value) {
    return
  }

  const confirmed = window.confirm('Lock the active tournament entries right now? This will pull the deadline to now.')
  if (!confirmed) {
    return
  }

  operationsErrorMessage.value = ''
  operationsSuccessMessage.value = ''
  operationsLoading.value = true

  try {
    const response = await apiFetch('/api/admin/lock', {
      method: 'POST',
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to lock active entries.'))
    }

    const payload = await response.json()
    operationsSuccessMessage.value = `Locked ${payload.locked_entries} entries for ${payload.year}.`
    await loadConfig()
    await loadEntries()
  } catch (error) {
    operationsErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while locking entries.'
  } finally {
    operationsLoading.value = false
  }
}

async function loadEntries() {
  entriesLoading.value = true
  entriesErrorMessage.value = ''
  entriesSuccessMessage.value = ''

  try {
    const response = await apiFetch('/api/admin/entries')
    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load admin entries.'))
    }

    entries.value = await response.json()
  } catch (error) {
    entriesErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading admin entries.'
  } finally {
    entriesLoading.value = false
  }
}

function startEditingEntry(entry) {
  editingEntryId.value = entry.id
  entryDrafts.value = {
    ...entryDrafts.value,
    [entry.id]: {
      display_name: entry.display_name,
      in_overs: Boolean(entry.in_overs),
      picks_json: prettyJSON(entry.picks ?? {}),
    },
  }
  entriesErrorMessage.value = ''
  entriesSuccessMessage.value = ''
}

function cancelEditingEntry() {
  editingEntryId.value = ''
}

async function saveEntry(entry) {
  if (savingEntryId.value) {
    return
  }

  const draft = entryDrafts.value[entry.id]
  if (!draft) {
    return
  }

  entriesErrorMessage.value = ''
  entriesSuccessMessage.value = ''

  let picks

  try {
    picks = JSON.parse(draft.picks_json)
  } catch {
    entriesErrorMessage.value = 'Entry picks JSON must be valid before saving.'
    return
  }

  const payload = {
    display_name: draft.display_name.trim(),
    picks,
    in_overs: draft.in_overs,
  }

  savingEntryId.value = entry.id

  try {
    const response = await apiFetch(`/api/admin/entries/${entry.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to save entry.'))
    }

    const updated = await response.json()
    entries.value = entries.value.map((currentEntry) => (currentEntry.id === updated.id ? updated : currentEntry))
    startEditingEntry(updated)
    editingEntryId.value = ''
    entriesSuccessMessage.value = `Saved ${updated.display_name}.`
  } catch (error) {
    entriesErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while saving the entry.'
  } finally {
    savingEntryId.value = ''
  }
}

async function deleteEntry(entry) {
  if (deletingEntryId.value) {
    return
  }

  const confirmed = window.confirm(`Delete ${entry.display_name}'s entry?`)
  if (!confirmed) {
    return
  }

  entriesErrorMessage.value = ''
  entriesSuccessMessage.value = ''
  deletingEntryId.value = entry.id

  try {
    const response = await apiFetch(`/api/admin/entries/${entry.id}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to delete entry.'))
    }

    entries.value = entries.value.filter((currentEntry) => currentEntry.id !== entry.id)
    if (editingEntryId.value === entry.id) {
      editingEntryId.value = ''
    }
    entriesSuccessMessage.value = `Deleted ${entry.display_name}.`
  } catch (error) {
    entriesErrorMessage.value = error instanceof Error ? error.message : 'Something went wrong while deleting the entry.'
  } finally {
    deletingEntryId.value = ''
  }
}

function toDateTimeLocalValue(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)
  const offsetMs = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function toDateInputValue(value) {
  if (!value) {
    return ''
  }

  return new Date(value).toISOString().slice(0, 10)
}

function prettyJSON(value) {
  return JSON.stringify(value, null, 2)
}

</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Admin Controls</p>
      <h2>Operate the tournament year</h2>
      <p>
        This route now covers both year config and active-year entries. It is
        still built for mock-admin use right now, with real Clerk admin state
        to follow later.
      </p>
    </div>

    <div class="status-card">
      <p class="status-label">Editing Year</p>
      <p class="status-value">{{ activeYear }}</p>
      <p class="status-meta">{{ helperMessage }}</p>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Tournament Operations</p>
        <h3>Refresh standings data or lock the field</h3>
      </div>
    </div>

    <div class="entry-form">
      <div v-if="operationsErrorMessage" class="alert alert-error">
        <p>{{ operationsErrorMessage }}</p>
      </div>

      <div v-if="operationsSuccessMessage" class="alert alert-success">
        <p>{{ operationsSuccessMessage }}</p>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Refresh year</span>
          <input v-model="operationsForm.refresh_year" :disabled="operationsLoading" type="number" min="2024" step="1" />
        </label>

        <label class="field">
          <span>Provider tournament ID</span>
          <input
            v-model="operationsForm.tournament_id"
            :disabled="operationsLoading"
            type="text"
            placeholder="Optional if saved in config"
          />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Provider round ID</span>
          <input
            v-model="operationsForm.round_id"
            :disabled="operationsLoading"
            type="number"
            min="1"
            step="1"
            placeholder="Optional"
          />
        </label>

        <div class="field admin-ops-card">
          <span>Entry lock</span>
          <p class="helper-copy">
            Use this if you need to slam the window shut immediately without hand-editing the deadline.
          </p>
          <button class="danger-button" type="button" :disabled="operationsLoading" @click="lockEntriesNow">
            {{ operationsLoading ? 'Working...' : 'Lock Entries Now' }}
          </button>
        </div>
      </div>

      <div class="field">
        <span>Golfer results JSON</span>
        <textarea
          v-model="operationsForm.results_json"
          :disabled="operationsLoading"
          rows="12"
          class="json-area"
        />
      </div>

      <div class="form-footer">
        <p class="helper-copy">
          You can either fetch a live snapshot from the configured golf provider using tournament IDs, or paste a manual results array and store that directly. Given the free-tier limits, the safer workflow is usually: fetch once, store the snapshot, then do most testing against the saved standings data instead of refetching.
        </p>
        <div class="entry-actions">
          <button class="ghost-button" type="button" :disabled="operationsLoading" @click="fetchProviderResults">
            {{ operationsLoading ? 'Working...' : 'Fetch Provider Snapshot' }}
          </button>
          <button class="submit-button" type="button" :disabled="operationsLoading" @click="refreshResults">
            {{ operationsLoading ? 'Saving Snapshot...' : 'Save Manual Snapshot' }}
          </button>
        </div>
      </div>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Tournament Config</p>
        <h3>Deadline, payouts, and player groups</h3>
      </div>
      <button class="ghost-button" type="button" :disabled="configLoading || configSaving" @click="loadConfig">
        Reload
      </button>
    </div>

    <div v-if="configLoading" class="empty-state">
      <p>Loading config...</p>
    </div>

    <div v-else class="entry-form">
      <div v-if="configErrorMessage" class="alert alert-error">
        <p>{{ configErrorMessage }}</p>
      </div>

      <div v-if="configSuccessMessage" class="alert alert-success">
        <p>{{ configSuccessMessage }}</p>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Entry deadline</span>
          <input v-model="form.entry_deadline" :disabled="configSaving" type="datetime-local" />
        </label>

        <label class="field">
          <span>First-round winner</span>
          <input
            v-model="form.frl_winner"
            :disabled="configSaving"
            type="text"
            placeholder="Optional until round one finishes"
          />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Start date</span>
          <input v-model="form.start_date" :disabled="configSaving" type="date" />
        </label>

        <label class="field">
          <span>End date</span>
          <input v-model="form.end_date" :disabled="configSaving" type="date" />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Provider tournament ID</span>
          <input
            v-model="form.provider_tournament_id"
            :disabled="configSaving"
            type="text"
            placeholder="Keep zero padding, for example 033"
          />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Mutt multiplier</span>
          <input v-model="form.mutt_multiplier" :disabled="configSaving" type="text" />
        </label>

        <label class="field">
          <span>Old mutt multiplier</span>
          <input v-model="form.old_mutt_multiplier" :disabled="configSaving" type="text" />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>FRL payout</span>
          <input v-model="form.frl_payout" :disabled="configSaving" type="number" min="0" step="1" />
        </label>

        <label class="checkbox-field">
          <input v-model="form.active" :disabled="configSaving" type="checkbox" />
          <span>Mark this config as the active tournament year</span>
        </label>
      </div>

      <div class="field">
        <span>Groups JSON</span>
        <textarea v-model="form.groups_json" :disabled="configSaving" rows="12" class="json-area" />
      </div>

      <div class="field">
        <span>Pool payouts JSON</span>
        <textarea v-model="form.pool_payouts_json" :disabled="configSaving" rows="8" class="json-area" />
      </div>

      <div class="form-footer">
        <p class="helper-copy">
          Save will call the admin-only backend route. With mock auth, this page
          requires `MOCK_AUTH_ADMIN=true` in your local environment.
        </p>

        <button class="submit-button" type="button" :disabled="configSaving || configLoading" @click="saveConfig">
          {{ configSaving ? 'Saving Config...' : 'Save Config' }}
        </button>
      </div>
    </div>
  </section>

  <section class="panel">
    <div class="section-heading">
      <div>
        <p class="kicker">Active-Year Entries</p>
        <h3>{{ entries.length }} entries available for admin review</h3>
      </div>
      <button class="ghost-button" type="button" :disabled="entriesLoading" @click="loadEntries">
        Reload Entries
      </button>
    </div>

    <div v-if="entriesErrorMessage" class="alert alert-error">
      <p>{{ entriesErrorMessage }}</p>
    </div>

    <div v-if="entriesSuccessMessage" class="alert alert-success">
      <p>{{ entriesSuccessMessage }}</p>
    </div>

    <div v-if="entriesLoading" class="empty-state">
      <p>Loading entries...</p>
    </div>

    <div v-else-if="entries.length === 0" class="empty-state">
      <p>No active-year entries are available yet.</p>
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

        <div v-if="editingEntryId === entry.id" class="entry-form">
          <label class="field">
            <span>Display name</span>
            <input v-model="entryDrafts[entry.id].display_name" :disabled="savingEntryId === entry.id" type="text" />
          </label>

          <label class="checkbox-field">
            <input v-model="entryDrafts[entry.id].in_overs" :disabled="savingEntryId === entry.id" type="checkbox" />
            <span>Mark this entry as in overs</span>
          </label>

          <label class="field">
            <span>Picks JSON</span>
            <textarea
              v-model="entryDrafts[entry.id].picks_json"
              :disabled="savingEntryId === entry.id"
              rows="10"
              class="json-area"
            />
          </label>

          <div class="entry-actions">
            <button class="ghost-button" type="button" :disabled="savingEntryId === entry.id" @click="cancelEditingEntry">
              Cancel
            </button>
            <button class="submit-button" type="button" :disabled="savingEntryId === entry.id" @click="saveEntry(entry)">
              {{ savingEntryId === entry.id ? 'Saving Entry...' : 'Save Entry' }}
            </button>
          </div>
        </div>

        <template v-else>
          <dl class="pick-list">
            <div v-for="(pick, groupName) in entry.picks" :key="`${entry.id}-${groupName}`">
              <dt>{{ groupName }}</dt>
              <dd>{{ pick }}</dd>
            </div>
          </dl>

          <div class="entry-actions">
            <button class="ghost-button" type="button" :disabled="Boolean(deletingEntryId)" @click="startEditingEntry(entry)">
              Edit
            </button>
            <button
              class="danger-button"
              type="button"
              :disabled="deletingEntryId === entry.id"
              @click="deleteEntry(entry)"
            >
              {{ deletingEntryId === entry.id ? 'Deleting...' : 'Delete' }}
            </button>
          </div>
        </template>
      </article>
    </div>
  </section>
</template>
