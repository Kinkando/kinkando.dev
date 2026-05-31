<script lang="ts">
  import { deleteRecord } from '$lib/api/finance';
  import type { FinanceRecord } from '$lib/api/types';

  interface Props {
    records: FinanceRecord[];
    onDeleted: () => void;
  }

  let { records, onDeleted }: Props = $props();

  let deletingId = $state<string | null>(null);

  async function handleDelete(id: string) {
    deletingId = id;
    try {
      await deleteRecord(id);
      onDeleted();
    } catch {
      // silently ignore — parent will not refresh
    } finally {
      deletingId = null;
    }
  }

  function fmt(n: number) {
    return n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });
  }

  function fmtDate(d: string) {
    return new Date(d).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
</script>

<div class="rounded-xl border border-gray-800 bg-gray-900 p-6">
  <h2 class="mb-4 text-lg font-semibold text-white">Records</h2>

  {#if records.length === 0}
    <p class="py-8 text-center text-gray-500">No records for this month.</p>
  {:else}
    <div class="space-y-2">
      {#each records as record (record.id)}
        <div class="flex items-center justify-between rounded-lg bg-gray-800 px-4 py-3">
          <div class="flex min-w-0 items-center gap-3">
            <span
              class="shrink-0 rounded-full px-2 py-0.5 text-xs font-medium {record.type === 'income'
                ? 'bg-green-900/60 text-green-300'
                : 'bg-red-900/60 text-red-300'}"
            >
              {record.type}
            </span>
            <div class="min-w-0">
              <p class="truncate text-sm text-white">{record.category}</p>
              {#if record.note}
                <p class="truncate text-xs text-gray-500">{record.note}</p>
              {/if}
            </div>
          </div>
          <div class="ml-4 flex shrink-0 items-center gap-4">
            <div class="text-right">
              <p class="text-sm font-semibold {record.type === 'income' ? 'text-green-400' : 'text-red-400'}">
                {record.type === 'income' ? '+' : '-'}{fmt(record.amount)}
              </p>
              <p class="text-xs text-gray-500">{fmtDate(record.date)}</p>
            </div>
            <button
              onclick={() => handleDelete(record.id)}
              disabled={deletingId === record.id}
              class="text-gray-600 transition-colors hover:text-red-400 disabled:opacity-50"
              aria-label="Delete record"
            >
              <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
