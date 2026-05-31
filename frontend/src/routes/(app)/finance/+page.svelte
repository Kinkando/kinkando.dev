<script lang="ts">
  import { getSummary, listRecords } from '$lib/api/finance';
  import RecordForm from '$lib/components/finance/RecordForm.svelte';
  import RecordList from '$lib/components/finance/RecordList.svelte';
  import SummaryPanel from '$lib/components/finance/SummaryPanel.svelte';
  import type { FinanceRecord, FinanceSummary } from '$lib/api/types';

  const now = new Date();
  let month = $state(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`);

  let records = $state<FinanceRecord[]>([]);
  let summary = $state<FinanceSummary | null>(null);
  let loading = $state(false);
  let error = $state('');

  async function load() {
    loading = true;
    error = '';
    try {
      [records, summary] = await Promise.all([listRecords(month), getSummary(month)]);
    } catch {
      error = 'Failed to load finance data.';
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    // Re-fetch whenever month changes
    void load();
  });
</script>

<svelte:head>
  <title>Finance — kinkando.dev</title>
</svelte:head>

<div class="mx-auto max-w-6xl px-4 py-8">
  <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
    <h1 class="text-2xl font-bold text-white">Finance</h1>
    <div class="flex items-center gap-2">
      <label for="month-picker" class="text-sm text-gray-400">Month</label>
      <input
        id="month-picker"
        type="month"
        bind:value={month}
        class="rounded border border-gray-700 bg-gray-800 px-3 py-1.5 text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
      />
    </div>
  </div>

  {#if error}
    <p class="mb-6 rounded border border-red-700 bg-red-900/50 px-4 py-3 text-sm text-red-300">{error}</p>
  {/if}

  <div class="grid gap-6 lg:grid-cols-3">
    <div class="space-y-6 lg:col-span-1">
      <RecordForm onCreated={load} />
      <SummaryPanel {summary} />
    </div>

    <div class="lg:col-span-2">
      {#if loading}
        <div class="flex items-center justify-center py-24">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent"></div>
        </div>
      {:else}
        <RecordList {records} onDeleted={load} />
      {/if}
    </div>
  </div>
</div>
