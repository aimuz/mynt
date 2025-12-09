<script lang="ts">
    import { onMount } from 'svelte';
    import { api, type SystemStats } from '$lib/api';

    let { stats = $bindable() }: { stats: SystemStats | null } = $props();

    // Chart logic would go here, for now using simple progress bars

    function formatBytes(bytes: number) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    function formatRate(bytes: number) {
        return formatBytes(bytes) + '/s';
    }
</script>

<div class="space-y-6">
    <!-- CPU -->
    <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
        <h3 class="text-lg font-medium text-zinc-900 dark:text-zinc-100 mb-4">Processor</h3>
        {#if stats}
            <div class="flex items-center gap-4">
                <div class="w-16 h-16 rounded-full border-4 border-blue-500 flex items-center justify-center text-sm font-bold">
                    {stats.cpu.global.toFixed(0)}%
                </div>
                <div class="flex-1 space-y-2">
                    <div class="flex justify-between text-sm">
                        <span class="text-zinc-500">Total Usage</span>
                        <span class="text-zinc-900 dark:text-zinc-100">{stats.cpu.global.toFixed(1)}%</span>
                    </div>
                    <div class="h-2 bg-zinc-100 dark:bg-zinc-700 rounded-full overflow-hidden">
                        <div class="h-full bg-blue-500 transition-all duration-500" style="width: {stats.cpu.global}%"></div>
                    </div>
                </div>
            </div>
            <div class="grid grid-cols-4 gap-2 mt-4">
                 {#each stats.cpu.per_core as core, i}
                    <div class="flex flex-col gap-1">
                        <div class="h-10 bg-zinc-100 dark:bg-zinc-700 rounded flex items-end overflow-hidden">
                             <div class="w-full bg-blue-400 opacity-80 transition-all duration-500" style="height: {core}%"></div>
                        </div>
                    </div>
                 {/each}
            </div>
        {:else}
             <div class="animate-pulse h-20 bg-zinc-100 dark:bg-zinc-700 rounded"></div>
        {/if}
    </div>

    <!-- Memory -->
    <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
        <h3 class="text-lg font-medium text-zinc-900 dark:text-zinc-100 mb-4">Memory</h3>
        {#if stats}
            {@const usedPercent = (stats.memory.used / stats.memory.total) * 100}
            {@const swapPercent = stats.memory.swap_total ? (stats.memory.swap_used / stats.memory.swap_total) * 100 : 0}

            <div class="space-y-4">
                <div>
                    <div class="flex justify-between text-sm mb-1">
                        <span class="text-zinc-500">RAM ({formatBytes(stats.memory.used)} / {formatBytes(stats.memory.total)})</span>
                        <span class="text-zinc-900 dark:text-zinc-100">{usedPercent.toFixed(1)}%</span>
                    </div>
                     <div class="h-2 bg-zinc-100 dark:bg-zinc-700 rounded-full overflow-hidden">
                        <div class="h-full bg-green-500 transition-all duration-500" style="width: {usedPercent}%"></div>
                    </div>
                </div>

                 <div>
                    <div class="flex justify-between text-sm mb-1">
                        <span class="text-zinc-500">Swap ({formatBytes(stats.memory.swap_used)} / {formatBytes(stats.memory.swap_total)})</span>
                        <span class="text-zinc-900 dark:text-zinc-100">{swapPercent.toFixed(1)}%</span>
                    </div>
                     <div class="h-2 bg-zinc-100 dark:bg-zinc-700 rounded-full overflow-hidden">
                        <div class="h-full bg-yellow-500 transition-all duration-500" style="width: {swapPercent}%"></div>
                    </div>
                </div>
            </div>
        {:else}
             <div class="animate-pulse h-20 bg-zinc-100 dark:bg-zinc-700 rounded"></div>
        {/if}
    </div>

    <!-- Network Summary -->
    <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
        <h3 class="text-lg font-medium text-zinc-900 dark:text-zinc-100 mb-4">Network</h3>
         {#if stats}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                {#each Object.entries(stats.network) as [name, net]}
                    {#if net.rx_rate > 0 || net.tx_rate > 0}
                        <div class="p-3 bg-zinc-50 dark:bg-zinc-900 rounded border border-zinc-100 dark:border-zinc-800">
                            <div class="font-medium text-sm mb-2">{name}</div>
                            <div class="flex justify-between text-xs text-zinc-500">
                                <span>↓ {formatRate(net.rx_rate)}</span>
                                <span>↑ {formatRate(net.tx_rate)}</span>
                            </div>
                        </div>
                    {/if}
                {/each}
            </div>
         {:else}
             <div class="animate-pulse h-20 bg-zinc-100 dark:bg-zinc-700 rounded"></div>
         {/if}
    </div>
</div>
