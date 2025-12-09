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

    function formatRate(bytes: number) {
        return formatBytes(bytes) + '/s';
    }
</script>

<div class="space-y-6">
    <div class="bg-white dark:bg-zinc-800 rounded-lg p-6 shadow-sm border border-zinc-200 dark:border-zinc-700">
        <h2 class="text-xl font-semibold mb-6">Network Interfaces</h2>

        {#if stats}
            <div class="space-y-4">
                {#each Object.entries(stats.network) as [name, net]}
                    <div class="border border-zinc-200 dark:border-zinc-700 rounded-lg p-4">
                        <div class="flex items-center justify-between mb-4">
                            <div class="flex items-center gap-3">
                                <div class="w-10 h-10 rounded bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-zinc-500">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"></rect><rect x="2" y="14" width="20" height="8" rx="2" ry="2"></rect><line x1="6" y1="6" x2="6" y2="6"></line><line x1="6" y1="18" x2="6" y2="18"></line></svg>
                                </div>
                                <div>
                                    <div class="font-medium">{name}</div>
                                    <div class="text-xs text-zinc-500">{net.ip_address || 'No IP'}</div>
                                </div>
                            </div>
                             {#if net.is_up}
                                <span class="px-2 py-1 text-xs rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400">Up</span>
                             {:else}
                                <span class="px-2 py-1 text-xs rounded-full bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400">Down</span>
                             {/if}
                        </div>

                        <div class="grid grid-cols-2 gap-4">
                            <div class="bg-zinc-50 dark:bg-zinc-900/50 p-3 rounded">
                                <div class="text-xs text-zinc-500 mb-1">Download</div>
                                <div class="text-lg font-mono">{formatRate(net.rx_rate)}</div>
                                <div class="text-xs text-zinc-400 mt-1">Total: {formatBytes(net.rx_bytes)}</div>
                            </div>
                            <div class="bg-zinc-50 dark:bg-zinc-900/50 p-3 rounded">
                                <div class="text-xs text-zinc-500 mb-1">Upload</div>
                                <div class="text-lg font-mono">{formatRate(net.tx_rate)}</div>
                                <div class="text-xs text-zinc-400 mt-1">Total: {formatBytes(net.tx_bytes)}</div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
