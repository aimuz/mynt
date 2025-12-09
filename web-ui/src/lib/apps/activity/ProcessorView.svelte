<script lang="ts">
    import type { SystemStats } from '$lib/api';

    let { stats = $bindable() }: { stats: SystemStats | null } = $props();

    function formatPercent(val: number) {
        return val.toFixed(1) + '%';
    }
</script>

<div class="space-y-6">
    <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
        <h2 class="text-xl font-semibold mb-6">Processor Usage</h2>

        {#if stats}
            <div class="flex items-center gap-6 mb-8">
                 <div class="w-32 h-32 rounded-full border-8 border-blue-500 flex items-center justify-center text-3xl font-bold">
                    {stats.cpu.global.toFixed(0)}%
                </div>
                 <div class="space-y-1">
                    <div class="text-sm text-zinc-500">Logical Cores</div>
                    <div class="text-2xl font-mono">{stats.cpu.per_core.length}</div>
                    {#if stats.host_info}
                        <div class="text-sm text-zinc-500 mt-2">Model</div>
                         <div class="text-sm">{stats.host_info.platform || 'Unknown'}</div>
                    {/if}
                </div>
            </div>

            <h3 class="font-medium mb-4">Core Utilization</h3>
            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
                {#each stats.cpu.per_core as core, i}
                    <div class="flex flex-col gap-2 items-center">
                         <div class="h-32 w-8 bg-zinc-100 dark:bg-zinc-700 rounded-full flex items-end overflow-hidden relative">
                             <div class="w-full bg-blue-500 transition-all duration-300 rounded-b-full" style="height: {core}%"></div>
                         </div>
                         <span class="text-xs text-zinc-500">CPU {i}</span>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
