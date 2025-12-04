<script lang="ts">
    import type { Component } from "svelte";

    let {
        icon: Icon,
        title,
        description,
        actionLabel,
        onAction,
    }: {
        icon?: Component;
        title: string;
        description?: string;
        actionLabel?: string;
        onAction?: () => void;
    } = $props();
</script>

<div class="glass-card rounded-xl p-8 text-center">
    {#if Icon}
        <Icon class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground" />
    {/if}
    <h3 class="text-lg font-semibold text-foreground mb-2">
        {title}
    </h3>
    {#if description}
        <p class="text-sm text-muted-foreground mb-6">
            {description}
        </p>
    {/if}
    {#if actionLabel && onAction}
        <button
            onclick={onAction}
            class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
        >
            {actionLabel}
        </button>
    {/if}
</div>
