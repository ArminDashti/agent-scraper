<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { createSource, deleteSource, fetchSources, updateSource, type SourceRow } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const rows = ref<SourceRow[]>([])
const loading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const name = ref('')
const websiteUrl = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchSources()
    rows.value = data.sources ?? []
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load sources'
  } finally {
    loading.value = false
  }
}

async function onCreate() {
  saving.value = true
  try {
    await createSource({
      name: name.value.trim(),
      websiteUrl: websiteUrl.value.trim(),
      isEnabled: false,
    })
    name.value = ''
    websiteUrl.value = ''
    await load()
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Create failed'
  } finally {
    saving.value = false
  }
}

async function onToggle(row: SourceRow, next: boolean) {
  try {
    await updateSource(row.id, { name: row.name, websiteUrl: row.websiteUrl, isEnabled: next })
    await load()
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Update failed'
  }
}

async function onDelete(id: number) {
  try {
    await deleteSource(id)
    await load()
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Delete failed'
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-4 px-4 py-8">
    <Card>
      <CardHeader>
        <CardTitle>Sources</CardTitle>
      </CardHeader>
      <CardContent class="space-y-6">
        <p class="text-sm text-muted-foreground">Website placeholders. Real extractors wait until sites are provided.</p>
        <form class="grid gap-3 sm:grid-cols-4" @submit.prevent="onCreate">
          <input
            v-model="name"
            required
            placeholder="Name"
            class="rounded-md border border-input bg-background px-3 py-2 text-sm"
          />
          <input
            v-model="websiteUrl"
            placeholder="Website URL"
            class="rounded-md border border-input bg-background px-3 py-2 text-sm sm:col-span-2"
          />
          <Button type="submit" :disabled="saving">{{ saving ? 'Saving…' : 'Add source' }}</Button>
        </form>
        <p v-if="errorMessage" class="text-sm text-destructive">{{ errorMessage }}</p>
        <p v-else-if="loading" class="text-sm text-muted-foreground">Loading…</p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Website</TableHead>
              <TableHead>Enabled</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="rows.length === 0">
              <TableCell colspan="4" class="text-muted-foreground">No sources yet.</TableCell>
            </TableRow>
            <TableRow v-for="row in rows" :key="row.id">
              <TableCell>{{ row.name }}</TableCell>
              <TableCell>{{ row.websiteUrl || '—' }}</TableCell>
              <TableCell>
                <Switch :checked="row.isEnabled" @change="(next) => onToggle(row, next)" />
              </TableCell>
              <TableCell>
                <Button variant="ghost" size="sm" @click="onDelete(row.id)">Delete</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
