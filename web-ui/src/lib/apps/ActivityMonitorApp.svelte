<script lang="ts">
    import { onMount } from 'svelte';
    import { getContext } from 'svelte';
    import { api, type SystemStats } from '$lib/api';

    // Views
    import OverviewView from './activity/OverviewView.svelte';
    import ProcessorView from './activity/ProcessorView.svelte';
    import MemoryView from './activity/MemoryView.svelte';
    import NetworkView from './activity/NetworkView.svelte';
    import DriveView from './activity/DriveView.svelte';
    import ProcessView from './activity/ProcessView.svelte';

    let activeView = $state('overview');
    let stats: SystemStats | null = $state(null);
    let interval: ReturnType<typeof setInterval>;

    const desktop = getContext('desktop') as any; // Type hack for now

    async function loadStats() {
        try {
            stats = await api.getSystemStats();
        } catch (e) {
            console.error('Failed to load system stats:', e);
        }
    }

    onMount(() => {
        loadStats();
        interval = setInterval(loadStats, 1000); // 1s polling for smooth graphs
        return () => clearInterval(interval);
    });

    const views = [
        { id: 'overview', label: 'Overview', icon: 'M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z' },
        { id: 'processor', label: 'Processor', icon: 'M4 4h16v16H4V4zm2 2v12h12V6H6zm2 2h8v8H8V8z' }, // Simplified chip
        { id: 'memory', label: 'Memory', icon: 'M4 4h16v16H4V4zm2 2v2h12V6H6zm0 4v2h12v-2H6zm0 4v2h12v-2H6zm0 4v2h12v-2H6z' }, // RAM sticks
        { id: 'network', label: 'Network', icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z' }, // Globe-ish
        { id: 'drive', label: 'Disks', icon: 'M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2H6zm7 4c2.21 0 4 1.79 4 4s-1.79 4-4 4-4-1.79-4-4 1.79-4 4-4zm0 6c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2z' }, // HDD
        { id: 'processes', label: 'Processes', icon: 'M3 3h18v18H3V3zm2 2v14h14V5H5zm2 2h10v2H7V7zm0 4h10v2H7v-2zm0 4h7v2H7v-2z' } // List
    ];

    // SVG Paths are just placeholders, replacing with SVG elements directly in loop
</script>

<div class="flex h-full bg-zinc-50 dark:bg-zinc-900 text-zinc-900 dark:text-zinc-100">
    <!-- Sidebar -->
    <div class="w-64 border-r border-zinc-200 dark:border-zinc-800 bg-white dark:bg-zinc-900 flex flex-col">
        <div class="p-4 border-b border-zinc-200 dark:border-zinc-800">
            <h1 class="text-xl font-bold tracking-tight px-2">Activity Monitor</h1>
        </div>
        <nav class="flex-1 p-2 space-y-1">
            {#each views as view}
                <button
                    class="w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors
                           {activeView === view.id ? 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800'}"
                    onclick={() => activeView = view.id}
                >
                    <!-- Simple icons -->
                    <svg class="w-5 h-5 opacity-75" viewBox="0 0 24 24" fill="currentColor">
                        <path d={view.icon} />
                    </svg>
                    {view.label}
                </button>
            {/each}
        </nav>

        {#if stats?.host_info}
            <div class="p-4 text-xs text-zinc-400 border-t border-zinc-200 dark:border-zinc-800">
                <div class="truncate font-mono">{stats.host_info.hostname}</div>
                <div class="truncate">{stats.host_info.os} {stats.host_info.kernel}</div>
                <div class="mt-1">Uptime: {(stats.uptime / 3600).toFixed(1)}h</div>
            </div>
        {/if}
    </div>

    <!-- Main Content -->
    <div class="flex-1 overflow-auto p-6">
        {#if activeView === 'overview'}
            <OverviewView bind:stats />
        {:else if activeView === 'processor'}
            <ProcessorView bind:stats />
        {:else if activeView === 'memory'}
            <MemoryView bind:stats />
        {:else if activeView === 'network'}
            <NetworkView bind:stats />
        {:else if activeView === 'drive'}
            <DriveView bind:stats />
        {:else if activeView === 'processes'}
            <ProcessView />
        {/if}
    </div>
</div>
