import { computed, ref } from 'vue'
import { apiFetch, responseMessage } from './api'

const mode = ref('mock')
const clerkLoaded = ref(false)
const clerkSignedIn = ref(false)
const backendUser = ref(null)
const backendUserLoaded = ref(false)
const backendUserLoading = ref(false)

let tokenGetter = async () => null
let waiters = []
let backendUserRequest = null

function resolveWaiters() {
  for (const resolve of waiters) {
    resolve()
  }

  waiters = []
}

export const authMode = computed(() => mode.value)
export const isClerkEnabled = computed(() => mode.value === 'clerk')
export const isAuthReady = computed(() => mode.value !== 'clerk' || clerkLoaded.value)
export const isSignedIn = computed(() => (mode.value === 'clerk' ? clerkSignedIn.value : false))
export const currentBackendUser = computed(() => backendUser.value)
export const isAdmin = computed(() => Boolean(backendUser.value?.is_admin))
export const hasBackendUser = computed(() => backendUserLoaded.value)

export function configureAuthMode(nextMode) {
  mode.value = nextMode

  if (nextMode !== 'clerk') {
    clerkLoaded.value = true
    clerkSignedIn.value = false
    backendUser.value = null
    backendUserLoaded.value = false
    backendUserLoading.value = false
    tokenGetter = async () => null
    resolveWaiters()
    return
  }

  clerkLoaded.value = false
  clerkSignedIn.value = false
  backendUser.value = null
  backendUserLoaded.value = false
  backendUserLoading.value = false
}

export function setTokenGetter(nextGetter) {
  tokenGetter = typeof nextGetter === 'function' ? nextGetter : async () => null
}

export async function getAuthToken() {
  return await tokenGetter()
}

export function markClerkLoaded(loaded) {
  clerkLoaded.value = loaded

  if (loaded) {
    resolveWaiters()
  }
}

export function setSignedInState(nextValue) {
  clerkSignedIn.value = Boolean(nextValue)
}

export function clearBackendUser() {
  backendUser.value = null
  backendUserLoaded.value = true
  backendUserLoading.value = false
  backendUserRequest = null
}

export async function waitForAuthReady() {
  if (isAuthReady.value) {
    return
  }

  await new Promise((resolve) => {
    waiters.push(resolve)
  })
}

export async function refreshBackendUser(options = {}) {
  const { force = false } = options

  if (mode.value !== 'clerk') {
    return null
  }

  if (!clerkSignedIn.value) {
    clearBackendUser()
    return null
  }

  if (backendUserRequest && !force) {
    return backendUserRequest
  }

  if (backendUserLoaded.value && !force) {
    return backendUser.value
  }

  backendUserLoading.value = true
  backendUserRequest = (async () => {
    const response = await apiFetch('/api/me')

    if (!response.ok) {
      throw new Error(await responseMessage(response, 'Failed to load the authenticated user.'))
    }

    const payload = await response.json()
    backendUser.value = payload
    backendUserLoaded.value = true
    return payload
  })()

  try {
    return await backendUserRequest
  } finally {
    backendUserLoading.value = false
    backendUserRequest = null
  }
}
