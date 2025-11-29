<script lang="ts">
    import { CircleAlert, RotateCw } from "@lucide/svelte";

    interface WidgetProps {
        title: string;
        children: any;
        icon?: any;
        size?: "small" | "medium" | "large";
    }

    let { title, icon, children, size = "medium" }: WidgetProps = $props();

    const sizeClasses = {
        small: "col-span-1 row-span-1",
        medium: "col-span-2 row-span-1",
        large: "col-span-2 row-span-2",
    };

    function handleError(error: unknown, reset: () => void) {
        console.error(`[Widget: ${title}] Error caught:`, error);
    }
</script>

<div
    class="glass-card rounded-2xl p-4 shadow-lg {sizeClasses[
        size
    ]} fade-in hover:shadow-xl transition-all"
>
    <div class="flex items-center gap-2 mb-3">
        {#if icon}
            {@const SvelteComponent = icon}
            <SvelteComponent class="w-4 h-4 text-foreground/70" />
        {/if}
        <h3 class="text-sm font-semibold text-foreground">{title}</h3>
    </div>

    <div class="text-foreground/90">
        <svelte:boundary onerror={handleError}>
            {@render children()}
            {#snippet failed(error, reset)}
                <div
                    class="flex flex-col items-center justify-center py-6 text-center"
                >
                    <CircleAlert class="w-8 h-8 text-red-400 mb-2" />
                    <p class="text-sm text-foreground/70 mb-1">
                        Something went wrong
                    </p>
                    <p
                        class="text-xs text-foreground/50 mb-3 max-w-full truncate"
                        title={error instanceof Error
                            ? error.message
                            : String(error)}
                    >
                        {error instanceof Error ? error.message : String(error)}
                    </p>
                    <button
                        onclick={reset}
                        class="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-primary/20 hover:bg-primary/30 text-primary rounded-lg transition-colors"
                    >
                        <RotateCw class="w-3 h-3" />
                        Retry
                    </button>
                </div>
            {/snippet}
        </svelte:boundary>
    </div>
</div>
