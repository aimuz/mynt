<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { api } from "$lib/api";

  onMount(async () => {
    // Check if system needs setup
    // In a real app, you'd check via an API endpoint
    // For now, we'll check if user is authenticated

    // If authenticated, go to desktop
    if (api.isAuthenticated()) {
      goto("/desktop");
    } else {
      // Check if setup is needed (you can add an API call here)
      // For now, assume if not authenticated, go to login
      // In production, add: const needsSetup = await api.checkSetupStatus();
      goto("/login");
    }
  });
</script>

<div class="min-h-screen desktop-bg flex items-center justify-center">
  <div class="text-center">
    <div
      class="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"
    ></div>
    <p class="text-foreground/60">Loading...</p>
  </div>
</div>
