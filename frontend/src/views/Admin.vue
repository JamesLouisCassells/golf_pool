<script setup>
import { computed, reactive, ref } from 'vue'

const activeYear = new Date().getFullYear()

const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

const form = reactive({
  entry_deadline: '',
  start_date: '',
  end_date: '',
  mutt_multiplier: '2',
  old_mutt_multiplier: '3',
  frl_winner: '',
  frl_payout: 500000,
  active: false,
  groups_json: '{\n  "Group 1": []\n}',
  pool_payouts_json: '{\n  "1": 4475\n}',
})

const helperMessage = computed(() =>
  'This admin page is intentionally lightweight. It is enough to operate tournament config while richer admin tooling is still pending.',
)

loadConfig()

async function loadConfig() {
  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`/api/admin/config/${activeYear}`)
    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load admin config.'))
    }

    const config = await response.json()
    form.entry_deadline = toDateTimeLocalValue(config.entry_deadline)
    form.start_date = toDateInputValue(config.start_date)
    form.end_date = toDateInputValue(config.end_date)
    form.mutt_multiplier = config.mutt_multiplier ?? '2'
    form.old_mutt_multiplier = config.old_mutt_multiplier ?? '3'
    form.frl_winner = config.frl_winner ?? ''
    form.frl_payout = config.frl_payout ?? 500000
    form.active = Boolean(config.active)
    form.groups_json = prettyJSON(config.groups ?? {})
    form.pool_payouts_json = prettyJSON(config.pool_payouts ?? {})
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while loading admin config.'
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  if (saving.value) {
    return
  }

  errorMessage.value = ''
  successMessage.value = ''

  let groups
  let poolPayouts

  try {
    groups = JSON.parse(form.groups_json)
    poolPayouts = JSON.parse(form.pool_payouts_json)
  } catch {
    errorMessage.value = 'Groups and pool payouts must be valid JSON.'
    return
  }

  const payload = {
    entry_deadline: form.entry_deadline ? new Date(form.entry_deadline).toISOString() : null,
    start_date: form.start_date ? new Date(`${form.start_date}T00:00:00`).toISOString() : null,
    end_date: form.end_date ? new Date(`${form.end_date}T00:00:00`).toISOString() : null,
    groups,
    mutt_multiplier: form.mutt_multiplier.trim(),
    old_mutt_multiplier: form.old_mutt_multiplier.trim(),
    pool_payouts: poolPayouts,
    frl_winner: form.frl_winner.trim() || null,
    frl_payout: Number(form.frl_payout),
    active: form.active,
  }

  saving.value = true

  try {
    const response = await fetch(`/api/admin/config/${activeYear}`, {
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
    form.groups_json = prettyJSON(updated.groups ?? {})
    form.pool_payouts_json = prettyJSON(updated.pool_payouts ?? {})
    successMessage.value = 'Tournament config saved.'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Something went wrong while saving admin config.'
  } finally {
    saving.value = false
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

async function responseMessage(response, fallback) {
  const text = (await response.text()).trim()
  return text || fallback
}
</script>

<template>
  <section class="panel hero-panel">
    <div class="hero-copy">
      <p class="kicker">Admin Config</p>
      <h2>Operate the tournament year</h2>
      <p>
        This route is the control surface for the year-specific config row. It
        is meant for mock-admin use right now, with real Clerk admin state to
        follow later.
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
        <p class="kicker">Tournament Config</p>
        <h3>Deadline, payouts, and player groups</h3>
      </div>
      <button class="ghost-button" type="button" :disabled="loading || saving" @click="loadConfig">
        Reload
      </button>
    </div>

    <div v-if="loading" class="empty-state">
      <p>Loading config...</p>
    </div>

    <div v-else class="entry-form">
      <div v-if="errorMessage" class="alert alert-error">
        <p>{{ errorMessage }}</p>
      </div>

      <div v-if="successMessage" class="alert alert-success">
        <p>{{ successMessage }}</p>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Entry deadline</span>
          <input v-model="form.entry_deadline" :disabled="saving" type="datetime-local" />
        </label>

        <label class="field">
          <span>First-round winner</span>
          <input v-model="form.frl_winner" :disabled="saving" type="text" placeholder="Optional until round one finishes" />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Start date</span>
          <input v-model="form.start_date" :disabled="saving" type="date" />
        </label>

        <label class="field">
          <span>End date</span>
          <input v-model="form.end_date" :disabled="saving" type="date" />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>Mutt multiplier</span>
          <input v-model="form.mutt_multiplier" :disabled="saving" type="text" />
        </label>

        <label class="field">
          <span>Old mutt multiplier</span>
          <input v-model="form.old_mutt_multiplier" :disabled="saving" type="text" />
        </label>
      </div>

      <div class="field-grid">
        <label class="field">
          <span>FRL payout</span>
          <input v-model="form.frl_payout" :disabled="saving" type="number" min="0" step="1" />
        </label>

        <label class="checkbox-field">
          <input v-model="form.active" :disabled="saving" type="checkbox" />
          <span>Mark this config as the active tournament year</span>
        </label>
      </div>

      <div class="field">
        <span>Groups JSON</span>
        <textarea v-model="form.groups_json" :disabled="saving" rows="12" class="json-area" />
      </div>

      <div class="field">
        <span>Pool payouts JSON</span>
        <textarea v-model="form.pool_payouts_json" :disabled="saving" rows="8" class="json-area" />
      </div>

      <div class="form-footer">
        <p class="helper-copy">
          Save will call the admin-only backend route. With mock auth, this page
          requires `MOCK_AUTH_ADMIN=true` in your local environment.
        </p>

        <button class="submit-button" type="button" :disabled="saving || loading" @click="saveConfig">
          {{ saving ? 'Saving Config...' : 'Save Config' }}
        </button>
      </div>
    </div>
  </section>
</template>
