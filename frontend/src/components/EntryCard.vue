<script setup>
import { computed, ref } from 'vue'
import GolferBadge from './GolferBadge.vue'

const props = defineProps({
  entry: {
    type: Object,
    required: true,
  },
  autoExpand: {
    type: Boolean,
    default: false,
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

const expanded = ref(props.autoExpand)
const pickCount = computed(() => props.entry.picks?.length ?? 0)
const leadersCount = computed(() =>
  (props.entry.picks ?? []).filter((pick) => {
    const position = String(pick.position || '').trim().toUpperCase()
    return position === '1' || position === 'T1'
  }).length,
)
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

    <div class="standings-card-summary">
      <p>{{ pickCount }} picks tracked</p>
      <p>{{ leadersCount }} in first</p>
      <button class="ghost-button standings-toggle" type="button" @click="expanded = !expanded">
        {{ expanded ? 'Hide details' : 'Show details' }}
      </button>
    </div>

    <div v-if="entry.frl_bonus" class="alert alert-success standings-frl">
      <p>FRL bonus applied: {{ formatMoney(entry.frl_bonus) }}</p>
    </div>

    <dl v-if="expanded" class="pick-list standings-pick-list standings-pick-list-expanded">
      <div v-for="pick in entry.picks" :key="`${entry.entry_id}-${pick.group_name}`">
        <dt>
          <span>{{ pick.group_name }}</span>
          <strong>{{ pick.golfer_name }}</strong>
        </dt>
        <dd>
          <GolferBadge :pick="pick" />
          <span>{{ payoutLabel(pick) }}</span>
        </dd>
      </div>
    </dl>
  </article>
</template>
