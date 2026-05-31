<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth } from '$lib/firebase/client';
  import { GoogleAuthProvider, createUserWithEmailAndPassword, signInWithEmailAndPassword, signInWithPopup } from 'firebase/auth';

  interface Props {
    mode: 'login' | 'register';
  }

  let { mode }: Props = $props();

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  function friendlyError(code: string): string {
    switch (code) {
      case 'auth/invalid-email':
        return 'Invalid email address.';
      case 'auth/user-not-found':
      case 'auth/wrong-password':
      case 'auth/invalid-credential':
        return 'Invalid email or password.';
      case 'auth/email-already-in-use':
        return 'An account with this email already exists.';
      case 'auth/weak-password':
        return 'Password must be at least 6 characters.';
      case 'auth/too-many-requests':
        return 'Too many attempts. Please try again later.';
      case 'auth/popup-closed-by-user':
        return 'Sign-in popup was closed.';
      default:
        return 'Something went wrong. Please try again.';
    }
  }

  function redirectAfterAuth() {
    const redirect = $page.url.searchParams.get('redirect') ?? '/kanban';
    goto(redirect);
  }

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      if (mode === 'login') {
        await signInWithEmailAndPassword(auth, email, password);
      } else {
        await createUserWithEmailAndPassword(auth, email, password);
      }
      redirectAfterAuth();
    } catch (err: unknown) {
      const code = (err as { code?: string }).code ?? '';
      error = friendlyError(code);
    } finally {
      loading = false;
    }
  }

  async function handleGoogle() {
    error = '';
    loading = true;
    try {
      await signInWithPopup(auth, new GoogleAuthProvider());
      redirectAfterAuth();
    } catch (err: unknown) {
      const code = (err as { code?: string }).code ?? '';
      error = friendlyError(code);
    } finally {
      loading = false;
    }
  }
</script>

<div class="mx-auto w-full max-w-sm">
  <h1 class="mb-6 text-center text-2xl font-bold text-white">
    {mode === 'login' ? 'Sign in' : 'Create account'}
  </h1>

  {#if error}
    <p class="mb-4 rounded border border-red-700 bg-red-900/50 px-4 py-3 text-sm text-red-300">
      {error}
    </p>
  {/if}

  <form onsubmit={handleSubmit} class="space-y-4">
    <div>
      <label for="email" class="mb-1 block text-sm font-medium text-gray-300">Email</label>
      <input
        id="email"
        type="email"
        bind:value={email}
        required
        autocomplete="email"
        class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-white placeholder-gray-500 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-indigo-500"
        placeholder="you@example.com"
      />
    </div>

    <div>
      <label for="password" class="mb-1 block text-sm font-medium text-gray-300">Password</label>
      <input
        id="password"
        type="password"
        bind:value={password}
        required
        autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
        class="w-full rounded border border-gray-700 bg-gray-800 px-3 py-2 text-white placeholder-gray-500 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-indigo-500"
        placeholder="••••••••"
      />
    </div>

    <button
      type="submit"
      disabled={loading}
      class="w-full rounded bg-indigo-600 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
    >
      {#if loading}
        {mode === 'login' ? 'Signing in…' : 'Creating account…'}
      {:else}
        {mode === 'login' ? 'Sign in' : 'Create account'}
      {/if}
    </button>
  </form>

  <div class="my-4 flex items-center gap-3">
    <div class="h-px flex-1 bg-gray-700"></div>
    <span class="text-xs text-gray-500">or</span>
    <div class="h-px flex-1 bg-gray-700"></div>
  </div>

  <button
    onclick={handleGoogle}
    disabled={loading}
    class="flex w-full items-center justify-center gap-2 rounded bg-white py-2.5 text-sm font-semibold text-gray-900 transition-colors hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50"
  >
    <svg class="h-4 w-4" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
    Continue with Google
  </button>

  <p class="mt-4 text-center text-sm text-gray-400">
    {#if mode === 'login'}
      Don't have an account? <a href="/register" class="text-indigo-400 hover:text-indigo-300">Register</a>
    {:else}
      Already have an account? <a href="/login" class="text-indigo-400 hover:text-indigo-300">Sign in</a>
    {/if}
  </p>
</div>
