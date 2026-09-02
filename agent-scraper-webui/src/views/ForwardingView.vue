<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchForwarding, saveForwarding } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const targetApiUrl = ref('')
const lastStatus = ref('')
const lastError = ref('')
const lastAt = ref('')
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchForwarding()
    targetApiUrl.value = data.targetApiUrl
    lastStatus.value = data.lastStatus
    lastError.value = data.lastError
    lastAt.value = data.lastAt ?? ''
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load forwarding'
  } finally {
    loading.value = false
  }
}

async function onSave() {
  saving.value = true
  try {
    const data = await saveForwarding(targetApiUrl.value.trim())
    targetApiUrl.value = data.targetApiUrl
    lastStatus.value = data.lastStatus
    lastError.value = data.lastError
    lastAt.value = data.lastAt ?? ''
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Save failed'
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="mx-auto max-w-3xl space-y-4 px-4 py-8">
    <Card>
      <CardHeader>
        <CardTitle>Forwarding</CardTitle>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">
          Destination POST API. Leave empty until the URL is known. Payload is
          <code>{ shop, item, expense }</code>.
        </p>
        <form class="space-y-3" @submit.prevent="onSave">
          <label class="block space-y-1 text-sm">
            <span>Target API URL</span>
            <input
              v-model="targetApiUrl"
              type="url"
              placeholder="https://example.com/expenses"
              class="w-full rounded-md border border-input bg-background px-3 py-2"
            />
          </label>
          <p v-if="errorMessage" class="text-sm text-destructive">{{ errorMessage }}</p>
          <p v-else-if="loading" class="text-sm text-muted-foreground">Loading…</p>
          <p v-else class="text-sm text-muted-foreground">
            Last POST: {{ lastStatus || 'none' }}
            <span v-if="lastAt"> at {{ lastAt }}</span>
            <span v-if="lastError"> — {{ lastError }}</span>
          </p>
          <Button type="submit" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
