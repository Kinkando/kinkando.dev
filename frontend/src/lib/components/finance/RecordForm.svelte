<script lang="ts">
  import { createRecord } from '$lib/api/finance';
  import type { RecordType } from '$lib/api/types';

  interface Props {
    onCreated: () => void;
  }

  let { onCreated }: Props = $props();

  let type = $state<RecordType>('expense');
  let amount = $state('');
  let category = $state('');
  let note = $state('');
  let date = $state(new Date().toISOString().slice(0, 10));
  let loading = $state(false);
  let error = $state('');

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      await createRecord({
        type,
        amount: parseFloat(amount),
        category,
        note,
        date
      });
      amount = '';
      category = '';
      note = '';
      onCreated();
    } catch {
      error = 'Failed to create record.';
    } finally {
      loading = false;
    }
  }
</script>

<div class="rounded-xl border border-gray-800 bg-gray-900 p-6">
  <h2 class="mb-4 text-lg font-semibold text-white">Add Record</h2>

  {#if error}
    <p class="mb-4 rounded border border-red-700 bg-red-900/50 px-3 py-2 text-sm text-red-300">{error}</p>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-3">
    <div class="grid grid-cols-2 gap-3">
      <div>
        <label for="rec-type" class="mb-1 block text-xs font-medium text-gray-400">Type</label>
        <select
          id="rec-type"
          bind:value={type}
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <option value="income">Income</option>
          <option value="expense">Expense</option>
        </select>
      </div>

      <div>
        <label for="rec-amount" class="mb-1 block text-xs font-medium text-gray-400">Amount</label>
        <input
          id="rec-amount"
          type="number"
          min="0.01"
          step="0.01"
          bind:value={amount}
          required
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="0.00"
        />
      </div>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div>
        <label for="rec-category" class="mb-1 block text-xs font-medium text-gray-400">Category</label>
        <input
          id="rec-category"
          type="text"
          bind:value={category}
          required
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          placeholder="e.g. Food"
        />
      </div>

      <div>
        <label for="rec-date" class="mb-1 block text-xs font-medium text-gray-400">Date</label>
        <input
          id="rec-date"
          type="date"
          bind:value={date}
          required
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>
    </div>

    <div>
      <label for="rec-note" class="mb-1 block text-xs font-medium text-gray-400">Note</label>
      <input
        id="rec-note"
        type="text"
        bind:value={note}
        class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        placeholder="Optional note"
      />
    </div>

    <button
      type="submit"
      disabled={loading}
      class="w-full rounded bg-indigo-600 py-2 text-sm font-semibold text-white transition-colors hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
    >
      {loading ? 'Adding…' : 'Add Record'}
    </button>
  </form>
</div>
