<script lang="ts">
    export interface DesktopApp {
        id: string;
        name: string;
        icon: any;
        color: string;
        onClick: () => void;
    }

    let { apps = [] }: { apps: DesktopApp[] } = $props();
</script>

<div
    class="fixed bottom-0 left-0 right-0 flex justify-center pb-4 px-4 pointer-events-none z-50"
>
    <div
        class="glass-strong rounded-2xl px-3 py-2 shadow-2xl pointer-events-auto slide-up"
    >
        <div class="flex items-end gap-2">
            {#each apps as app (app.id)}
                <button
                    onclick={app.onClick}
                    class="dock-item group relative flex flex-col items-center"
                >
                    <div
                        class="w-14 h-14 rounded-xl flex items-center justify-center shadow-lg"
                        style="background: {app.color};"
                    >
                        <app.icon class="w-8 h-8 text-white" />
                    </div>

                    <!-- App Name Tooltip -->
                    <div
                        class="absolute -top-12 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none"
                    >
                        <div
                            class="glass-card px-3 py-1.5 rounded-lg shadow-lg"
                        >
                            <span
                                class="text-sm font-medium text-foreground whitespace-nowrap"
                            >
                                {app.name}
                            </span>
                        </div>
                    </div>
                </button>
            {/each}
        </div>
    </div>
</div>

<style lang="postcss">
    @reference "tailwindcss";
    .dock-item {
        @apply transition-transform duration-200 ease-out;
        will-change: transform;
    }

    .dock-item:hover {
        transform: translateY(-8px) scale(1.1);
    }
</style>
