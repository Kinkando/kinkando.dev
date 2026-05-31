<script lang="ts">
  import { deleteCard, getBoard, moveCard } from '$lib/api/kanban';
  import type { KanbanCard, KanbanColumn } from '$lib/api/types';
  import KanbanColumn from '$lib/components/kanban/Column.svelte';
  import { onMount } from 'svelte';

  let columns = $state<KanbanColumn[]>([]);
  let cardsByColumn = $state<Record<string, KanbanCard[]>>({});
  let loading = $state(true);
  let error = $state('');

  async function load(showSpinner = true) {
    if (showSpinner) loading = true;
    error = '';
    try {
      const board = await getBoard();
      columns = [...board.columns].sort((a, b) => a.order - b.order);
      const grouped: Record<string, KanbanCard[]> = {};
      for (const col of columns) grouped[col.id] = [];
      for (const card of board.cards) {
        if (grouped[card.column_id]) {
          grouped[card.column_id].push(card);
        }
      }
      // Sort cards by order within each column
      for (const id of Object.keys(grouped)) {
        grouped[id].sort((a, b) => a.order - b.order);
      }
      cardsByColumn = grouped;
    } catch {
      error = 'Failed to load kanban board.';
    } finally {
      if (showSpinner) loading = false;
    }
  }

  /** Silent refresh — does not show a loading spinner (used after mutations). */
  function refresh() {
    return load(false);
  }

  onMount(load);

  function handleConsider(columnId: string, items: KanbanCard[]) {
    // Optimistic local update during drag
    cardsByColumn = { ...cardsByColumn, [columnId]: items };
  }

  async function handleFinalize(columnId: string, items: KanbanCard[]) {
    // Apply the drop locally first
    const previous = { ...cardsByColumn };
    cardsByColumn = { ...cardsByColumn, [columnId]: items };

    // Find any card that moved to this column or changed order
    const persists = items.map(async (card, idx) => {
      const expectedColumnId = columnId;
      const expectedOrder = idx;
      // Only PATCH if something actually changed
      if (card.column_id !== expectedColumnId || card.order !== expectedOrder) {
        try {
          await moveCard(card.id, { column_id: expectedColumnId, order: expectedOrder });
        } catch {
          // Revert on failure
          cardsByColumn = previous;
          await refresh();
        }
      }
    });
    await Promise.all(persists);
    // Refresh to get authoritative state
    await refresh();
  }

  async function handleDeleteCard(id: string) {
    // Optimistic remove
    const next: Record<string, KanbanCard[]> = {};
    for (const [colId, cards] of Object.entries(cardsByColumn)) {
      next[colId] = cards.filter((c) => c.id !== id);
    }
    cardsByColumn = next;
    try {
      await deleteCard(id);
    } catch {
      // Revert
      await refresh();
    }
  }
</script>

<svelte:head>
  <title>Kanban — kinkando.dev</title>
</svelte:head>

<div class="px-4 py-8">
  <h1 class="mb-6 text-2xl font-bold text-white">Kanban Board</h1>

  {#if loading}
    <div class="flex items-center justify-center py-24">
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent"></div>
    </div>
  {:else if error}
    <div class="flex flex-col items-center gap-4 py-24">
      <p class="text-red-400">{error}</p>
      <button onclick={load} class="rounded bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"> Retry </button>
    </div>
  {:else}
    <div class="flex gap-4 overflow-x-auto pb-4">
      {#each columns as column (column.id)}
        <KanbanColumn
          {column}
          cards={cardsByColumn[column.id] ?? []}
          onConsider={handleConsider}
          onFinalize={handleFinalize}
          onDeleteCard={handleDeleteCard}
        />
      {/each}
    </div>
  {/if}
</div>
