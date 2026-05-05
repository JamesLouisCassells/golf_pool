<script setup>
import { onMounted, onUnmounted, unref, watch } from 'vue'
import { useAuth } from '@clerk/vue'

import {
  clearBackendUser,
  markClerkLoaded,
  refreshBackendUser,
  setSignedInState,
  setTokenGetter,
} from '../lib/auth'

const auth = useAuth()

function resolveValue(candidate) {
  return unref(candidate)
}

async function readToken() {
  const getter = resolveValue(auth.getToken)
  if (typeof getter !== 'function') {
    return null
  }

  return await getter()
}

onMounted(() => {
  setTokenGetter(readToken)
})

onUnmounted(() => {
  setTokenGetter(async () => null)
})

watch(
  () => Boolean(resolveValue(auth.isLoaded)),
  (loaded) => {
    markClerkLoaded(loaded)
  },
  { immediate: true },
)

watch(
  () => [Boolean(resolveValue(auth.isLoaded)), Boolean(resolveValue(auth.isSignedIn))],
  async ([loaded, signedIn]) => {
    if (!loaded) {
      return
    }

    setSignedInState(signedIn)

    if (!signedIn) {
      clearBackendUser()
      return
    }

    try {
      await refreshBackendUser({ force: true })
    } catch (error) {
      console.error(error)
    }
  },
  { immediate: true },
)
</script>

<template>
  <span hidden aria-hidden="true" />
</template>
