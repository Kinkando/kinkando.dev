<script lang="ts">
  import type { FinanceSummary } from '$lib/api/types';

  interface Props {
    summary: FinanceSummary | null;
  }

  let { summary }: Props = $props();

  function fmt(n: number) {
    return n.toLocaleString('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });
  }
</script>

{#if summary}
  <div class="rounded-xl border border-gray-800 bg-gray-900 p-6">
    <h2 class="mb-4 text-lg font-semibold text-white">Summary</h2>

    <div class="mb-6 grid grid-cols-3 gap-4">
      <div class="rounded-lg bg-gray-800 p-4 text-center">
        <p class="mb-1 text-xs font-medium uppercase tracking-wider text-gray-400">Income</p>
        <p class="text-lg font-bold text-green-400">{fmt(summary.income)}</p>
      </div>
      <div class="rounded-lg bg-gray-800 p-4 text-center">
        <p class="mb-1 text-xs font-medium uppercase tracking-wider text-gray-400">Expense</p>
        <p class="text-lg font-bold text-red-400">{fmt(summary.expense)}</p>
      </div>
      <div class="rounded-lg bg-gray-800 p-4 text-center">
        <p class="mb-1 text-xs font-medium uppercase tracking-wider text-gray-400">Net</p>
        <p class="text-lg font-bold {summary.net >= 0 ? 'text-green-400' : 'text-red-400'}">{fmt(summary.net)}</p>
      </div>
    </div>

    {#if summary.categories.length > 0}
      <h3 class="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-400">By category</h3>
      <div class="space-y-2">
        {#each summary.categories as cat (cat.category + cat.type)}
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="inline-block h-2 w-2 rounded-full {cat.type === 'income' ? 'bg-green-400' : 'bg-red-400'}"></span>
              <span class="text-sm text-gray-300">{cat.category}</span>
              <span class="text-xs text-gray-500">{cat.type}</span>
            </div>
            <span class="text-sm font-medium {cat.type === 'income' ? 'text-green-400' : 'text-red-400'}">
              {fmt(cat.total)}
            </span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
