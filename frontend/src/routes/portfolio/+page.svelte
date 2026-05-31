<script lang="ts">
  import { getProjects, getSkills } from '$lib/api/portfolio';
  import type { Project, SkillGroup } from '$lib/api/types';
  import { onMount } from 'svelte';

  let projects = $state<Project[]>([]);
  let skills = $state<SkillGroup[]>([]);
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    try {
      [projects, skills] = await Promise.all([getProjects(), getSkills()]);
    } catch {
      error = 'Failed to load portfolio data.';
    } finally {
      loading = false;
    }
  });
</script>

<svelte:head>
  <title>Portfolio — kinkando.dev</title>
</svelte:head>

<div class="mx-auto max-w-6xl px-4 py-12">
  {#if loading}
    <div class="flex items-center justify-center py-24">
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-indigo-500 border-t-transparent"></div>
    </div>
  {:else if error}
    <p class="py-12 text-center text-red-400">{error}</p>
  {:else}
    <!-- Projects -->
    <section class="mb-16">
      <h2 class="mb-2 text-3xl font-bold text-white">Projects</h2>
      <p class="mb-8 text-gray-400">Things I've built</p>

      {#if projects.length === 0}
        <p class="text-gray-500">No projects yet.</p>
      {:else}
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {#each projects as project (project.name)}
            <div class="flex flex-col rounded-xl border border-gray-800 bg-gray-900 p-6 transition-colors hover:border-indigo-700">
              <h3 class="mb-2 text-lg font-semibold text-white">{project.name}</h3>
              <p class="mb-4 flex-1 text-sm text-gray-400">{project.description}</p>

              <div class="mb-4 flex flex-wrap gap-1.5">
                {#each project.tags as tag (tag)}
                  <span class="rounded-full border border-indigo-800 bg-indigo-900/50 px-2 py-0.5 text-xs text-indigo-300">
                    {tag}
                  </span>
                {/each}
              </div>

              {#if project.url}
                <a href={project.url} target="_blank" rel="noopener noreferrer" class="text-sm font-medium text-indigo-400 hover:text-indigo-300">
                  View project →
                </a>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </section>

    <!-- Skills -->
    <section>
      <h2 class="mb-2 text-3xl font-bold text-white">Skills</h2>
      <p class="mb-8 text-gray-400">Technologies I work with</p>

      {#if skills.length === 0}
        <p class="text-gray-500">No skills listed.</p>
      {:else}
        <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {#each skills as group (group.category)}
            <div class="rounded-xl border border-gray-800 bg-gray-900 p-6">
              <h3 class="mb-4 text-sm font-semibold uppercase tracking-wider text-indigo-400">
                {group.category}
              </h3>
              <div class="flex flex-wrap gap-2">
                {#each group.items as item (item)}
                  <span class="rounded-full border border-gray-700 bg-gray-800 px-3 py-1 text-sm text-gray-300">
                    {item}
                  </span>
                {/each}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
</div>
