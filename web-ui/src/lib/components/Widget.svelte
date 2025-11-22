<script lang="ts">
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
        {@render children()}
    </div>
</div>
