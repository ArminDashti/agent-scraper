<script setup lang="ts">
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { clearSession, getToken } from '@/lib/api'
import { computed } from 'vue'

const router = useRouter()
const route = useRoute()
const isAuthenticated = computed(() => {
  void route.fullPath
  return !!getToken()
})

const links = [
  { to: '/jobs', label: 'Jobs' },
  { to: '/expenses', label: 'Expenses' },
  { to: '/sources', label: 'Sources' },
  { to: '/forwarding', label: 'Forwarding' },
]

function onLogout() {
  clearSession()
  void router.push('/login')
}

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <div class="flex h-full w-full flex-col bg-background text-foreground">
    <header class="sticky top-0 z-40 flex shrink-0 items-center justify-between gap-4 border-b border-border px-4 py-3.5">
      <nav class="flex items-center gap-3">
        <RouterLink to="/jobs" class="text-base font-semibold tracking-tight">Agent Scraper</RouterLink>
        <template v-if="isAuthenticated">
          <RouterLink
            v-for="link in links"
            :key="link.to"
            :to="link.to"
            class="text-sm"
            :class="isActive(link.to) ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'"
          >
            {{ link.label }}
          </RouterLink>
        </template>
      </nav>
      <div class="flex items-center gap-2">
        <Button v-if="isAuthenticated" variant="ghost" size="sm" @click="onLogout">Log out</Button>
        <RouterLink v-else to="/login">
          <Button variant="outline" size="sm">Log in</Button>
        </RouterLink>
      </div>
    </header>
    <main class="min-h-0 flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
