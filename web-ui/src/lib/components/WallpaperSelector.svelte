<script lang="ts">
    import {
        currentWallpaper,
        predefinedWallpapers,
        type Wallpaper,
    } from "$lib/stores/wallpaper";
    import { Check } from "@lucide/svelte";

    interface Props {
        onClose: () => void;
    }

    let { onClose }: Props = $props();
    let selectedWallpaper = $state<Wallpaper>($currentWallpaper);

    function handleSelect(wallpaper: Wallpaper) {
        selectedWallpaper = wallpaper;
    }

    function handleApply() {
        currentWallpaper.set(selectedWallpaper);
        onClose();
    }
</script>

<div class="p-6 h-full flex flex-col">
    <p class="text-sm text-muted-foreground mb-6">
        Select a wallpaper for your desktop
    </p>

    <!-- Wallpaper Grid -->
    <div
        class="grid grid-cols-2 md:grid-cols-3 gap-4 mb-6 flex-1 overflow-y-auto overflow-x-hidden p-2"
    >
        {#each predefinedWallpapers as wallpaper (wallpaper.id)}
            <button
                onclick={() => handleSelect(wallpaper)}
                class="group relative aspect-video rounded-xl overflow-hidden border-2 transition-all hover:scale-105 hover:shadow-lg {selectedWallpaper.id ===
                wallpaper.id
                    ? 'border-blue-500 ring-2 ring-blue-500/30'
                    : 'border-white/20 hover:border-white/40'}"
            >
                <!-- Wallpaper Preview -->
                {#if wallpaper.type === "image"}
                    <img
                        src={wallpaper.thumbnail || wallpaper.value}
                        alt={wallpaper.name}
                        class="w-full h-full object-cover"
                    />
                {:else}
                    <div
                        class="w-full h-full"
                        style="background: {wallpaper.value};"
                    ></div>
                {/if}

                <!-- Selected Indicator -->
                {#if selectedWallpaper.id === wallpaper.id}
                    <div
                        class="absolute top-2 right-2 w-6 h-6 bg-blue-500 rounded-full flex items-center justify-center shadow-lg"
                    >
                        <Check class="w-4 h-4 text-white" />
                    </div>
                {/if}

                <!-- Name Overlay -->
                <div
                    class="absolute bottom-0 left-0 right-0 bg-linear-to-t from-black/60 to-transparent p-3"
                >
                    <p class="text-white text-sm font-medium">
                        {wallpaper.name}
                    </p>
                </div>
            </button>
        {/each}
    </div>

    <!-- Actions -->
    <div
        class="flex items-center justify-end gap-3 pt-4 border-t border-white/10"
    >
        <button
            onclick={onClose}
            class="px-4 py-2 rounded-lg hover:bg-black/5 dark:hover:bg-white/5 transition-colors text-sm font-medium"
        >
            Cancel
        </button>
        <button
            onclick={handleApply}
            class="px-4 py-2 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition-colors text-sm font-medium shadow-lg shadow-blue-500/30"
        >
            Apply Wallpaper
        </button>
    </div>
</div>
