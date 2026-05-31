<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import NavBar from '$lib/components/NavBar.svelte';
  import { initAuth } from '$lib/stores/auth.svelte';

  let { children } = $props();

  onMount(() => {
    initAuth();
  });
</script>

<NavBar />
<main class="min-h-screen bg-gray-950 text-gray-100">
  {@render children()}
</main>
