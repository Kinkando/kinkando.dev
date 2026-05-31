<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth } from '$lib/firebase/client';
  import { authState } from '$lib/stores/auth.svelte';
  import { signOut } from 'firebase/auth';

  async function logout() {
    await signOut(auth);
    goto('/login');
  }

  function isActive(path: string) {
    return $page.url.pathname === path || $page.url.pathname.startsWith(path + '/');
  }
</script>

<nav class="border-b border-gray-800 bg-gray-900">
  <div class="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
    <a href="/" class="text-lg font-bold tracking-tight text-white">kinkando.dev</a>

    <div class="flex items-center gap-1">
      <a
        href="/portfolio"
        class="rounded px-3 py-1.5 text-sm font-medium transition-colors {isActive('/portfolio')
          ? 'bg-indigo-600 text-white'
          : 'text-gray-300 hover:bg-gray-800 hover:text-white'}"
      >
        Portfolio
      </a>

      {#if authState.user}
        <a
          href="/kanban"
          class="rounded px-3 py-1.5 text-sm font-medium transition-colors {isActive('/kanban')
            ? 'bg-indigo-600 text-white'
            : 'text-gray-300 hover:bg-gray-800 hover:text-white'}"
        >
          Kanban
        </a>
        <a
          href="/finance"
          class="rounded px-3 py-1.5 text-sm font-medium transition-colors {isActive('/finance')
            ? 'bg-indigo-600 text-white'
            : 'text-gray-300 hover:bg-gray-800 hover:text-white'}"
        >
          Finance
        </a>
        <button
          onclick={logout}
          class="ml-2 rounded px-3 py-1.5 text-sm font-medium text-gray-300 transition-colors hover:bg-gray-800 hover:text-white"
        >
          Logout
        </button>
      {:else}
        <a
          href="/login"
          class="rounded px-3 py-1.5 text-sm font-medium transition-colors {isActive('/login')
            ? 'bg-indigo-600 text-white'
            : 'text-gray-300 hover:bg-gray-800 hover:text-white'}"
        >
          Login
        </a>
        <a href="/register" class="ml-1 rounded bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white transition-colors hover:bg-indigo-500">
          Register
        </a>
      {/if}
    </div>
  </div>
</nav>
