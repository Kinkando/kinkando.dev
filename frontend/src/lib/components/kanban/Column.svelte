<script lang="ts">
  import { createCard } from '$lib/api/kanban';
  import type { KanbanCard, KanbanColumn } from '$lib/api/types';
  import { dndzone } from 'svelte-dnd-action';
  import KanbanCardComponent from './Card.svelte';

  interface DndEvent {
    detail: {
      items: KanbanCard[];
    };
  }

  interface Props {
    column: KanbanColumn;
    cards: KanbanCard[];
    onConsider: (columnId: string, items: KanbanCard[]) => void;
    onFinalize: (columnId: string, items: KanbanCard[]) => void;
    onDeleteCard: (id: string) => void;
  }

  let { column, cards, onConsider, onFinalize, onDeleteCard }: Props = $props();

  let adding = $state(false);
  let newTitle = $state('');
  let newContent = $state('');
  let addLoading = $state(false);
  let addError = $state('');
  let titleInput = $state<HTMLInputElement | null>(null);

  $effect(() => {
    if (adding && titleInput) titleInput.focus();
  });

  async function handleAdd(e: SubmitEvent) {
    e.preventDefault();
    if (!newTitle.trim()) return;
    addLoading = true;
    addError = '';
    try {
      await createCard({ column_id: column.id, title: newTitle.trim(), content: newContent.trim() });
      newTitle = '';
      newContent = '';
      adding = false;
      // parent will re-fetch via onFinalize or manual trigger — we emit a fake finalize
      onFinalize(column.id, cards);
    } catch {
      addError = 'Failed to add card.';
    } finally {
      addLoading = false;
    }
  }

  function cancelAdd() {
    adding = false;
    newTitle = '';
    newContent = '';
    addError = '';
  }
</script>

<div class="flex h-full w-72 shrink-0 flex-col rounded-xl border border-gray-800 bg-gray-900">
  <!-- Column header -->
  <div class="flex items-center justify-between px-4 py-3">
    <h3 class="text-sm font-semibold text-gray-200">{column.name}</h3>
    <span class="rounded-full bg-gray-800 px-2 py-0.5 text-xs text-gray-400">{cards.length}</span>
  </div>

  <!-- Drop zone -->
  <div
    use:dndzone={{ items: cards, flipDurationMs: 150, type: 'kanban-card' }}
    onconsider={(e: DndEvent) => onConsider(column.id, e.detail.items)}
    onfinalize={(e: DndEvent) => onFinalize(column.id, e.detail.items)}
    class="flex min-h-[4rem] flex-1 flex-col gap-2 overflow-y-auto px-3 pb-3"
  >
    {#each cards as card (card.id)}
      <KanbanCardComponent {card} onDelete={onDeleteCard} />
    {/each}
  </div>

  <!-- Add card -->
  <div class="px-3 pb-3">
    {#if adding}
      <form onsubmit={handleAdd} class="space-y-2">
        {#if addError}
          <p class="text-xs text-red-400">{addError}</p>
        {/if}
        <input
          type="text"
          bind:this={titleInput}
          bind:value={newTitle}
          placeholder="Card title"
          required
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <input
          type="text"
          bind:value={newContent}
          placeholder="Description (optional)"
          class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <div class="flex gap-2">
          <button
            type="submit"
            disabled={addLoading}
            class="flex-1 rounded bg-indigo-600 py-1.5 text-xs font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
          >
            {addLoading ? 'Adding…' : 'Add'}
          </button>
          <button type="button" onclick={cancelAdd} class="rounded bg-gray-800 px-3 py-1.5 text-xs text-gray-400 hover:text-white"> Cancel </button>
        </div>
      </form>
    {:else}
      <button
        onclick={() => (adding = true)}
        class="flex w-full items-center gap-1.5 rounded px-2 py-1.5 text-xs text-gray-500 transition-colors hover:bg-gray-800 hover:text-gray-300"
      >
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        Add card
      </button>
    {/if}
  </div>
</div>
