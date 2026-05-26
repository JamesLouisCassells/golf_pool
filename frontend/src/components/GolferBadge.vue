<script setup>
import { computed } from 'vue'

const props = defineProps({
  pick: {
    type: Object,
    required: true,
  },
})

const positionLabel = computed(() => props.pick.position || 'Pending')
const todayLabel = computed(() => props.pick.today || 'E')
const thruLabel = computed(() => props.pick.thru || 'Awaiting tee time')
const badgeClass = computed(() => {
  const thru = String(props.pick.thru || '').trim().toUpperCase()
  const position = String(props.pick.position || '').trim().toUpperCase()

  if (thru === 'CUT' || position === 'CUT') {
    return 'golfer-badge-cut'
  }
  if (thru === 'F') {
    return 'golfer-badge-final'
  }
  if (position.startsWith('T1') || position === '1') {
    return 'golfer-badge-leading'
  }

  return 'golfer-badge-live'
})
</script>

<template>
  <div class="golfer-badge" :class="badgeClass">
    <span class="golfer-badge-position">{{ positionLabel }}</span>
    <span class="golfer-badge-score">{{ pick.score || 'No score yet' }}</span>
    <span class="golfer-badge-meta">{{ todayLabel }} today</span>
    <span class="golfer-badge-meta">{{ thruLabel }}</span>
  </div>
</template>
