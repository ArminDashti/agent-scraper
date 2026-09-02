<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchJobs, runJobNow, type JobRun } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const cronSchedule = ref('')
const runs = ref<JobRun[]>([])
const loading = ref(false)
const running = ref(false)
const errorMessage = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchJobs()
    cronSchedule.value = data.cronSchedule
    runs.value = data.runs ?? []
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load jobs'
  } finally {
    loading.value = false
  }
}

async function onRunNow() {
  running.value = true
  try {
    const data = await runJobNow()
    if (data.runs) runs.value = data.runs
    else await load()
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Run failed'
  } finally {
    running.value = false
  }
}

onMounted(() => {
  void load()
})
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-4 px-4 py-8">
    <Card>
      <CardHeader class="flex flex-row items-center justify-between space-y-0">
        <CardTitle>Jobs</CardTitle>
        <Button size="sm" :disabled="running" @click="onRunNow">
          {{ running ? 'Running…' : 'Run now' }}
        </Button>
      </CardHeader>
      <CardContent>
        <p class="mb-4 text-sm text-muted-foreground">
          Schedule: {{ cronSchedule || '—' }} (placeholder until the daily time is set)
        </p>
        <p v-if="errorMessage" class="mb-3 text-sm text-destructive">{{ errorMessage }}</p>
        <p v-else-if="loading" class="text-sm text-muted-foreground">Loading…</p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Status</TableHead>
              <TableHead>Started</TableHead>
              <TableHead>Finished</TableHead>
              <TableHead>Extracted</TableHead>
              <TableHead>Error</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="runs.length === 0">
              <TableCell colspan="5" class="text-muted-foreground">No runs yet.</TableCell>
            </TableRow>
            <TableRow v-for="run in runs" :key="run.id">
              <TableCell>{{ run.status }}</TableCell>
              <TableCell>{{ run.startedAt }}</TableCell>
              <TableCell>{{ run.finishedAt || '—' }}</TableCell>
              <TableCell>{{ run.extractedCount }}</TableCell>
              <TableCell>{{ run.errorMessage || '—' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
