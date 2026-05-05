import { createApp } from 'vue'
import { clerkPlugin } from '@clerk/vue'

import './style.css'
import App from './App.vue'
import router from './router'
import { configureAuthMode } from './lib/auth'

const publishableKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY

const app = createApp(App)

if (publishableKey) {
  configureAuthMode('clerk')
  app.use(clerkPlugin, {
    publishableKey,
    routerPush: (to) => router.push(to),
    routerReplace: (to) => router.replace(to),
    signInFallbackRedirectUrl: '/enter',
    signUpFallbackRedirectUrl: '/enter',
    afterSignOutUrl: '/',
  })
} else {
  configureAuthMode('mock')
}

app.use(router).mount('#app')
