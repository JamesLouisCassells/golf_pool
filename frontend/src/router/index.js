import { createRouter, createWebHistory } from 'vue-router'

import EntriesView from '../views/Entries.vue'
import EnterView from '../views/Enter.vue'
import HomeView from '../views/Home.vue'
import StandingsView from '../views/Standings.vue'

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
      path: '/enter',
      name: 'enter',
      component: EnterView,
    },
    {
      path: '/standings',
      name: 'standings',
      component: StandingsView,
    },
  ],
})

export default router
