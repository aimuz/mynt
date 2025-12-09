<script lang="ts">
    import { onMount } from "svelte";
    import { api, type SystemStats, type ProcessInfo } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        Activity,
        Cpu,
        MemoryStick,
        HardDrive,
        Network,
        X,
        Search,
        RefreshCw
    } from "@lucide/svelte";

    import CpuPanel from "./activity/CpuPanel.svelte";
    import MemoryPanel from "./activity/MemoryPanel.svelte";
    import NetworkPanel from "./activity/NetworkPanel.svelte";

    // --- State ---
    let activeTab = $state<"cpu" | "memory" | "disk" | "network">("cpu");
    let stats = $state<SystemStats | null>(null);
    let processes = $state<ProcessInfo[]>([]);
    let loading = $state(true);
    let error = $state<string | null>(null);
    let searchQuery = $state("");
    let sortBy = $state<keyof ProcessInfo>("cpu_percent");
    let sortDesc = $state(true);
    let selectedPid = $state<number | null>(null);

    // --- Computed ---
    let filteredProcesses = $derived.by(() => {
        let filtered = processes.filter(p =>
            p.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            p.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
            p.pid.toString().includes(searchQuery)
        );

        return filtered.sort((a, b) => {
            const valA = a[sortBy];
            const valB = b[sortBy];

            if (typeof valA === 'string' && typeof valB === 'string') {
                 return sortDesc ? valB.localeCompare(valA) : valA.localeCompare(valB);
            }
            if (typeof valA === 'number' && typeof valB === 'number') {
                return sortDesc ? valB - valA : valA - valB;
            }
            return 0;
        });
    });

    let history = $state<{
        cpu: number[],
        memory: number[],
        networkRx: number[],
        networkTx: number[]
    }>({
        cpu: new Array(60).fill(0),
        memory: new Array(60).fill(0),
        networkRx: new Array(60).fill(0),
        networkTx: new Array(60).fill(0)
    });

    let lastNetworkStats = $state<{
        bytes_sent: number;
        bytes_recv: number;
        timestamp: number;
    } | null>(null);

    // --- Lifecycle ---
    onMount(() => {
        fetchData();
        const interval = setInterval(fetchData, 2000); // Poll every 2s

        return () => clearInterval(interval);
    });

    // --- Methods ---
    async function fetchData() {
        try {
            const [newStats, newProcs] = await Promise.all([
                api.getSystemStats(),
                api.getProcesses()
            ]);

            // Update history
            const now = Date.now();
            history.cpu = [...history.cpu.slice(1), newStats.cpu.total_usage];
            history.memory = [...history.memory.slice(1), newStats.memory.used_percent];

            // Network Stats calculation
            if (newStats.network) {
                if (lastNetworkStats) {
                    const timeDiff = (now - lastNetworkStats.timestamp) / 1000; // seconds
                    if (timeDiff > 0) {
                        const rxRate = (newStats.network.bytes_recv - lastNetworkStats.bytes_recv) / timeDiff;
                        const txRate = (newStats.network.bytes_sent - lastNetworkStats.bytes_sent) / timeDiff;

                        history.networkRx = [...history.networkRx.slice(1), rxRate];
                        history.networkTx = [...history.networkTx.slice(1), txRate];
                    }
                }
                lastNetworkStats = {
                    bytes_sent: newStats.network.bytes_sent,
                    bytes_recv: newStats.network.bytes_recv,
                    timestamp: now
                };
            }

            stats = newStats;
            processes = newProcs;
            loading = false;
        } catch (e) {
            error = e instanceof Error ? e.message : String(e);
            loading = false;
        }
    }

    async function handleKillProcess() {
        if (!selectedPid) return;
        if (!confirm(`Are you sure you want to kill process ${selectedPid}?`)) return;

        try {
            await api.killProcess(selectedPid);
            fetchData(); // Refresh immediately
            selectedPid = null;
        } catch (e) {
            alert("Failed to kill process: " + (e instanceof Error ? e.message : String(e)));
        }
    }

    function toggleSort(column: keyof ProcessInfo) {
        if (sortBy === column) {
            sortDesc = !sortDesc;
        } else {
            sortBy = column;
            sortDesc = true;
        }
    }

    function switchTab(tab: typeof activeTab) {
        activeTab = tab;
        // Default sorts
        if (tab === 'cpu') {
            sortBy = 'cpu_percent';
            sortDesc = true;
        } else if (tab === 'memory') {
            sortBy = 'memory_percent';
            sortDesc = true;
        } else {
            sortBy = 'pid';
            sortDesc = false;
        }
    }
</script>

<div class="flex flex-col h-full bg-background text-foreground">
    <!-- Toolbar -->
    <div class="flex items-center gap-2 p-2 border-b border-border glass-card">
        <div class="flex bg-muted rounded-md p-1">
            <button
                class="px-3 py-1 text-sm rounded-sm transition-colors {activeTab === 'cpu' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}"
                onclick={() => switchTab('cpu')}>CPU</button>
            <button
                class="px-3 py-1 text-sm rounded-sm transition-colors {activeTab === 'memory' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}"
                onclick={() => switchTab('memory')}>Memory</button>
            <button
                class="px-3 py-1 text-sm rounded-sm transition-colors {activeTab === 'disk' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}"
                onclick={() => switchTab('disk')}>Disk</button>
            <button
                class="px-3 py-1 text-sm rounded-sm transition-colors {activeTab === 'network' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}"
                onclick={() => switchTab('network')}>Network</button>
        </div>

        <div class="flex-1"></div>

        <div class="relative">
            <Search class="absolute left-2 top-1.5 w-4 h-4 text-muted-foreground" />
            <input
                type="text"
                placeholder="Search"
                bind:value={searchQuery}
                class="pl-8 pr-3 py-1 text-sm bg-muted rounded-md border-none focus:ring-1 focus:ring-primary w-48"
            />
        </div>

        <button
            onclick={handleKillProcess}
            disabled={!selectedPid}
            class="p-1.5 rounded-md hover:bg-red-500/20 hover:text-red-500 disabled:opacity-50 disabled:hover:bg-transparent disabled:hover:text-inherit transition-colors"
            title="Kill Process"
        >
            <X class="w-5 h-5" />
        </button>
    </div>

    <!-- Main Content -->
    <div class="flex-1 overflow-hidden flex flex-col">
        <!-- Process Table -->
        <div class="flex-1 overflow-auto">
            <table class="w-full text-left text-sm border-collapse">
                <thead class="sticky top-0 bg-muted/80 backdrop-blur-md z-10 text-xs font-medium text-muted-foreground">
                    <tr>
                        <th class="p-2 pl-4 cursor-pointer hover:bg-white/5" onclick={() => toggleSort('name')}>
                            Process Name {sortBy === 'name' ? (sortDesc ? '↓' : '↑') : ''}
                        </th>
                        {#if activeTab === 'cpu'}
                            <th class="p-2 w-20 text-right cursor-pointer hover:bg-white/5" onclick={() => toggleSort('cpu_percent')}>
                                % CPU {sortBy === 'cpu_percent' ? (sortDesc ? '↓' : '↑') : ''}
                            </th>
                        {/if}
                        {#if activeTab === 'memory'}
                            <th class="p-2 w-20 text-right cursor-pointer hover:bg-white/5" onclick={() => toggleSort('memory_percent')}>
                                % Mem {sortBy === 'memory_percent' ? (sortDesc ? '↓' : '↑') : ''}
                            </th>
                        {/if}
                         <th class="p-2 w-24 cursor-pointer hover:bg-white/5" onclick={() => toggleSort('username')}>
                            User {sortBy === 'username' ? (sortDesc ? '↓' : '↑') : ''}
                        </th>
                        <th class="p-2 w-20 text-right cursor-pointer hover:bg-white/5" onclick={() => toggleSort('pid')}>
                            PID {sortBy === 'pid' ? (sortDesc ? '↓' : '↑') : ''}
                        </th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-border/30">
                    {#each filteredProcesses as proc (proc.pid)}
                        <tr
                            class="hover:bg-primary/10 cursor-default select-none {selectedPid === proc.pid ? 'bg-primary/20' : ''}"
                            onclick={() => selectedPid = proc.pid}
                        >
                            <td class="p-2 pl-4 flex items-center gap-2">
                                <div class="truncate max-w-[200px]" title={proc.cmdline || proc.name}>
                                    {proc.name}
                                </div>
                            </td>
                            {#if activeTab === 'cpu'}
                                <td class="p-2 text-right font-mono">{proc.cpu_percent.toFixed(1)}</td>
                            {/if}
                            {#if activeTab === 'memory'}
                                <td class="p-2 text-right font-mono">{proc.memory_percent.toFixed(1)}</td>
                            {/if}
                            <td class="p-2 text-muted-foreground truncate max-w-[100px]">{proc.username}</td>
                            <td class="p-2 text-right text-muted-foreground font-mono">{proc.pid}</td>
                        </tr>
                    {/each}
                </tbody>
            </table>

            {#if loading && processes.length === 0}
                <div class="flex justify-center items-center h-full">
                    <RefreshCw class="animate-spin w-8 h-8 text-muted-foreground" />
                </div>
            {/if}
        </div>

        <!-- Bottom Stats Panel -->
        {#if stats}
            {#if activeTab === 'cpu'}
                <CpuPanel {history} {stats} />
            {:else if activeTab === 'memory'}
                <MemoryPanel {history} {stats} />
            {:else if activeTab === 'network'}
                <NetworkPanel {history} {stats} />
            {:else}
                <div class="h-40 border-t border-border bg-muted/30 p-4 flex items-center justify-center text-muted-foreground">
                    No data for {activeTab}
                </div>
            {/if}
        {/if}
    </div>
</div>
