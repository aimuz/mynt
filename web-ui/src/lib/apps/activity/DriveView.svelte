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
        <h2 class="text-xl font-semibold mb-6">Disk Activity</h2>

        {#if stats}
            <div class="space-y-4">
                {#each Object.entries(stats.disk) as [name, disk]}
                    <!-- Only show physical disks usually, but gopsutil returns partitions too.
                         We'll display whatever we get for now. -->
                    <div class="border border-zinc-200 dark:border-zinc-700 rounded-lg p-4">
                        <div class="flex items-center gap-3 mb-4">
                            <div class="w-10 h-10 rounded bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-zinc-500">
                                <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="12" x2="2" y2="12"></line><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"></path><line x1="6" y1="16" x2="6.01" y2="16"></line><line x1="10" y1="16" x2="10.01" y2="16"></line></svg>
                            </div>
                            <div class="font-medium">{name}</div>
                        </div>

                        <div class="grid grid-cols-2 gap-4">
                            <div class="bg-zinc-50 dark:bg-zinc-900/50 p-3 rounded">
                                <div class="text-xs text-zinc-500 mb-1">Read</div>
                                <div class="text-lg font-mono">{formatRate(disk.read_rate)}</div>
                                <div class="text-xs text-zinc-400 mt-1">Total: {formatBytes(disk.read_bytes)}</div>
                            </div>
                            <div class="bg-zinc-50 dark:bg-zinc-900/50 p-3 rounded">
                                <div class="text-xs text-zinc-500 mb-1">Write</div>
                                <div class="text-lg font-mono">{formatRate(disk.write_rate)}</div>
                                <div class="text-xs text-zinc-400 mt-1">Total: {formatBytes(disk.write_bytes)}</div>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
