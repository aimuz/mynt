<script lang="ts">
    import type { SystemStats } from '$lib/api';

    let { stats = $bindable() }: { stats: SystemStats | null } = $props();

    function formatBytes(bytes: number) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
</script>

<div class="space-y-6">
    {#if stats}
        <!-- RAM -->
        {@const usedPercent = (stats.memory.used / stats.memory.total) * 100}
        {@const cachedPercent = (stats.memory.cached / stats.memory.total) * 100}
        <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
            <h2 class="text-xl font-semibold mb-6">Memory</h2>

            <div class="flex items-center gap-6 mb-8">
                 <div class="w-32 h-32 rounded-full border-8 border-green-500 flex items-center justify-center text-3xl font-bold">
                    {usedPercent.toFixed(0)}%
                </div>
                 <div class="space-y-1">
                    <div class="text-sm text-zinc-500">Total Memory</div>
                    <div class="text-2xl font-mono">{formatBytes(stats.memory.total)}</div>
                </div>
            </div>

            <div class="space-y-4">
                 <div>
                    <div class="flex justify-between text-sm mb-1">
                        <span class="flex items-center gap-2"><div class="w-3 h-3 rounded-full bg-green-500"></div> Used</span>
                        <span>{formatBytes(stats.memory.used)}</span>
                    </div>
                 </div>
                 <div>
                    <div class="flex justify-between text-sm mb-1">
                        <span class="flex items-center gap-2"><div class="w-3 h-3 rounded-full bg-blue-400"></div> Cached</span>
                        <span>{formatBytes(stats.memory.cached)}</span>
                    </div>
                 </div>
                 <div>
                    <div class="flex justify-between text-sm mb-1">
                        <span class="flex items-center gap-2"><div class="w-3 h-3 rounded-full bg-zinc-300"></div> Free</span>
                        <span>{formatBytes(stats.memory.free)}</span>
                    </div>
                 </div>

                 <!-- Stacked Bar -->
                 <div class="h-4 w-full rounded-full overflow-hidden flex mt-4">
                     <div class="h-full bg-green-500" style="width: {usedPercent}%"></div>
                     <div class="h-full bg-blue-400" style="width: {cachedPercent}%"></div>
                     <div class="h-full bg-zinc-200 dark:bg-zinc-700" style="flex: 1"></div>
                 </div>
            </div>
        </div>

        <!-- Swap -->
         <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
            <h2 class="text-xl font-semibold mb-6">Swap</h2>
             {#if stats.memory.swap_total > 0}
                {@const swapPercent = (stats.memory.swap_used / stats.memory.swap_total) * 100}
                <div class="flex items-center gap-6">
                     <div class="w-24 h-24 rounded-full border-8 border-yellow-500 flex items-center justify-center text-2xl font-bold">
                        {swapPercent.toFixed(0)}%
                    </div>
                    <div class="flex-1">
                        <div class="flex justify-between text-sm mb-2">
                             <span>Used</span>
                             <span>{formatBytes(stats.memory.swap_used)} / {formatBytes(stats.memory.swap_total)}</span>
                        </div>
                        <div class="h-2 bg-zinc-100 dark:bg-zinc-700 rounded-full overflow-hidden">
                            <div class="h-full bg-yellow-500 transition-all duration-500" style="width: {swapPercent}%"></div>
                        </div>
                    </div>
                </div>
             {:else}
                <p class="text-zinc-500">No swap configured.</p>
             {/if}
         </div>
    {/if}
</div>
