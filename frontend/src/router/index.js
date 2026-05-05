import { createRouter, createWebHistory } from 'vue-router'

import AdminView from '../views/Admin.vue'
import EntriesView from '../views/Entries.vue'
import EnterView from '../views/Enter.vue'
import HomeView from '../views/Home.vue'
import SignInView from '../views/SignIn.vue'
import StandingsView from '../views/Standings.vue'
import {
  authMode,
  hasBackendUser,
  isAdmin,
  isSignedIn,
  refreshBackendUser,
  waitForAuthReady,
} from '../lib/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/entries',
      name: 'entries',
      component: EntriesView,
    },
    {
      path: '/admin',
      name: 'admin',
      component: AdminView,
      meta: {
        requiresAuth: true,
        requiresAdmin: true,
      },
    },
    {
      path: '/enter',
      name: 'enter',
      component: EnterView,
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/sign-in',
      name: 'sign-in',
      component: SignInView,
    },
    {
      path: '/standings',
      name: 'standings',
      component: StandingsView,
    },
  ],
})

router.beforeEach(async (to) => {
  await waitForAuthReady()

  if (authMode.value !== 'clerk') {
    return true
  }

  if (!to.meta.requiresAuth) {
    return true
  }

  if (!isSignedIn.value) {
    return {
      name: 'sign-in',
      query: {
        redirect_url: to.fullPath,
      },
    }
  }

  if (!to.meta.requiresAdmin) {
    return true
  }

  if (!hasBackendUser.value) {
    try {
      await refreshBackendUser()
    } catch {
      return {
        name: 'home',
        query: {
          auth: 'session-error',
        },
      }
    }
  }

  if (!isAdmin.value) {
    return {
      name: 'home',
      query: {
        auth: 'admin-required',
      },
    }
  }

  return true
})

export default router
