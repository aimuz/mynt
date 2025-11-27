<script lang="ts">
    import { onMount } from "svelte";
    import { api, type Notification } from "$lib/api";
    import { formatDate } from "$lib/utils";
    import { Bell, CircleCheck } from "@lucide/svelte";

    let notifications = $state<Notification[]>([]);
    let loading = $state(true);

    onMount(async () => {
        try {
            notifications = await api.listNotifications("unread", 5);
            loading = false;
        } catch (error) {
            console.error("Failed to load notifications:", error);
            loading = false;
        }
    });
</script>

{#if loading}
    <div class="flex items-center justify-center py-8">
        <div
            class="animate-spin rounded-full h-6 w-6 border-2 border-primary border-t-transparent"
        ></div>
    </div>
{:else if notifications.length === 0}
    <div class="text-center py-8 text-foreground/50">
        <CircleCheck class="w-8 h-8 mx-auto mb-2 opacity-50" />
        <p class="text-sm">All caught up!</p>
    </div>
{:else}
    <div class="space-y-2">
        {#each notifications as notif}
            <div
                class="p-2 rounded-lg hover:bg-foreground/5 transition-colors border border-foreground/10"
            >
                <div class="flex items-start gap-2">
                    <Bell class="w-3 h-3 text-blue-500 mt-0.5 flex-shrink-0" />
                    <div class="flex-1 min-w-0">
                        <p class="text-xs font-medium text-foreground truncate">
                            {notif.type}
                        </p>
                        <p class="text-xs text-foreground/60 mt-0.5">
                            {formatDate(notif.created_at)}
                        </p>
                    </div>
                </div>
            </div>
        {/each}
    </div>
{/if}
