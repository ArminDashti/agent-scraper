<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { fetchExpenses, type ExpenseRow } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

const rows = ref<ExpenseRow[]>([])
const loading = ref(false)
const errorMessage = ref('')

async function load() {
  loading.value = true
  try {
    const data = await fetchExpenses()
    rows.value = data.expenses ?? []
    errorMessage.value = ''
  } catch (err) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to load expenses'
  } finally {
    loading.value = false
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
        <CardTitle>Expenses</CardTitle>
      </CardHeader>
      <CardContent>
        <p v-if="errorMessage" class="mb-3 text-sm text-destructive">{{ errorMessage }}</p>
        <p v-else-if="loading" class="text-sm text-muted-foreground">Loading…</p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Shop</TableHead>
              <TableHead>Item</TableHead>
              <TableHead>Expense</TableHead>
              <TableHead>Created</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="rows.length === 0">
              <TableCell colspan="4" class="text-muted-foreground">No expenses stored yet.</TableCell>
            </TableRow>
            <TableRow v-for="row in rows" :key="row.id">
              <TableCell>{{ row.shop }}</TableCell>
              <TableCell>{{ row.item }}</TableCell>
              <TableCell>{{ row.expense }}</TableCell>
              <TableCell>{{ row.createdAt }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
