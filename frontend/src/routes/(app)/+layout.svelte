<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { ensureUserOnce, authState } from '$lib/stores/auth.svelte';

  let { children } = $props();

  // When auth resolves, redirect unauthenticated users to login
  $effect(() => {
    if (!authState.loading && !authState.user) {
      goto(`/login?redirect=${encodeURIComponent($page.url.pathname)}`);
    }
  });

  // Await ensureUserOnce so finance endpoints work before children mount
  let ready = $state(false);
  $effect(() => {
    if (!authState.loading && authState.user) {
      ensureUserOnce().then(() => {
        ready = true;
      });
    }
  });
</script>

{#if authState.loading}
  <!-- Auth resolving — render-gate prevents protected content flash -->
  <div class="flex min-h-[calc(100vh-56px)] items-center justify-center">
    <div class="h-8 w-8 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent"></div>
  </div>
{:else if authState.user && ready}
  {@render children()}
{/if}
