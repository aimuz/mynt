<script lang="ts">
    import { onMount } from 'svelte';
    import { api, type ProcessInfo } from '$lib/api';

    let processes: ProcessInfo[] = $state([]);
    let searchQuery = $state('');
    let sortField: keyof ProcessInfo = $state('cpu_percent');
    let sortDirection: 'asc' | 'desc' = $state('desc');
    let loading = $state(false);
    let error: string | null = $state(null);
    let selectedPid: number | null = $state(null);

    async function loadProcesses() {
        loading = true;
        try {
            processes = await api.getProcesses();
            error = null;
        } catch (e: any) {
            error = e.message;
        } finally {
            loading = false;
        }
    }

    onMount(() => {
        loadProcesses();
        const interval = setInterval(loadProcesses, 3000); // Refresh every 3s
        return () => clearInterval(interval);
    });

    async function killProcess(pid: number) {
        if (!confirm(`Are you sure you want to kill process ${pid}?`)) return;
        try {
            await api.killProcess(pid);
            await loadProcesses();
        } catch (e: any) {
            alert('Failed to kill process: ' + e.message);
        }
    }

    let filteredProcesses = $derived(processes
        .filter(p => p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                     p.pid.toString().includes(searchQuery))
        .sort((a, b) => {
            const aVal = a[sortField];
            const bVal = b[sortField];
            if (aVal < bVal) return sortDirection === 'asc' ? -1 : 1;
            if (aVal > bVal) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        })
    );

    function toggleSort(field: keyof ProcessInfo) {
        if (sortField === field) {
            sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
        } else {
            sortField = field;
            sortDirection = 'desc';
        }
    }

    function formatBytes(bytes: number) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
</script>

<div class="h-full flex flex-col bg-white dark:bg-zinc-800 rounded-lg shadow-sm border border-zinc-200 dark:border-zinc-700">
    <div class="p-4 border-b border-zinc-200 dark:border-zinc-700 flex justify-between items-center">
        <h2 class="text-xl font-semibold">Processes</h2>
        <div class="flex gap-2">
            <input
                type="text"
                bind:value={searchQuery}
                placeholder="Search processes..."
                class="px-3 py-1.5 rounded-md border border-zinc-300 dark:border-zinc-600 bg-transparent text-sm"
            />
            {#if selectedPid}
                <button
                    onclick={() => killProcess(selectedPid!)}
                    class="px-3 py-1.5 rounded-md bg-red-500 hover:bg-red-600 text-white text-sm font-medium"
                >
                    End Task
                </button>
            {/if}
        </div>
    </div>

    <div class="flex-1 overflow-auto">
        <table class="w-full text-sm text-left">
            <thead class="text-xs text-zinc-500 uppercase bg-zinc-50 dark:bg-zinc-900 sticky top-0">
                <tr>
                    <th class="px-4 py-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800" onclick={() => toggleSort('name')}>Name</th>
                    <th class="px-4 py-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800" onclick={() => toggleSort('pid')}>PID</th>
                    <th class="px-4 py-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800" onclick={() => toggleSort('username')}>User</th>
                    <th class="px-4 py-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800 text-right" onclick={() => toggleSort('cpu_percent')}>CPU %</th>
                    <th class="px-4 py-3 cursor-pointer hover:bg-zinc-100 dark:hover:bg-zinc-800 text-right" onclick={() => toggleSort('mem_rss')}>Memory</th>
                </tr>
            </thead>
            <tbody class="divide-y divide-zinc-200 dark:divide-zinc-700">
                {#each filteredProcesses as p (p.pid)}
                    <tr
                        class="hover:bg-zinc-50 dark:hover:bg-zinc-700 cursor-pointer {selectedPid === p.pid ? 'bg-blue-50 dark:bg-blue-900/20' : ''}"
                        onclick={() => selectedPid = p.pid}
                    >
                        <td class="px-4 py-2 font-medium truncate max-w-[200px]" title={p.cmdline || p.name}>{p.name}</td>
                        <td class="px-4 py-2 text-zinc-500">{p.pid}</td>
                        <td class="px-4 py-2 text-zinc-500">{p.username}</td>
                        <td class="px-4 py-2 text-right font-mono">{p.cpu_percent.toFixed(1)}%</td>
                        <td class="px-4 py-2 text-right font-mono">{formatBytes(p.mem_rss)}</td>
                    </tr>
                {/each}
            </tbody>
        </table>

        {#if loading && processes.length === 0}
            <div class="p-8 text-center text-zinc-500">Loading processes...</div>
        {/if}
    </div>

    <div class="p-2 border-t border-zinc-200 dark:border-zinc-700 text-xs text-zinc-500 text-right">
        Total: {processes.length} processes
    </div>
</div>
