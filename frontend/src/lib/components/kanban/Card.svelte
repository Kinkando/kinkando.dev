<script lang="ts">
  import type { KanbanCard } from '$lib/api/types';

  interface Props {
    card: KanbanCard;
    onDelete: (id: string) => void;
  }

  let { card, onDelete }: Props = $props();
</script>

<div class="group cursor-grab rounded-lg border border-gray-700 bg-gray-800 p-3 active:cursor-grabbing">
  <div class="flex items-start justify-between gap-2">
    <p class="text-sm font-medium text-white">{card.title}</p>
    <button
      onclick={() => onDelete(card.id)}
      class="shrink-0 text-gray-600 opacity-0 transition-opacity hover:text-red-400 group-hover:opacity-100"
      aria-label="Delete card"
    >
      <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
  {#if card.content}
    <p class="mt-1 text-xs text-gray-400">{card.content}</p>
  {/if}
</div>
