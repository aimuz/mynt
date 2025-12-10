<script lang="ts">
    import { onMount } from "svelte";
    import { api, type SystemStats, type SysProcess } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        Cpu,
        MemoryStick,
        Network,
        HardDrive,
        Activity,
        House,
    } from "@lucide/svelte";
    import OverviewView from "./activity/OverviewView.svelte";
    import CpuView from "./activity/CpuView.svelte";
    import MemoryView from "./activity/MemoryView.svelte";
    import NetworkView from "./activity/NetworkView.svelte";
    import DiskIOView from "./activity/DiskIOView.svelte";
    import ProcessesView from "./activity/ProcessesView.svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    // History configuration
    const HISTORY_SIZE = 60; // 60 data points = 2 minutes at 2s interval

    // View state
    let currentView = $state<string>("overview");
    let stats = $state<SystemStats | null>(null);
    let processes = $state<SysProcess[]>([]);
    let loading = $state(true);
    let error = $state<string | null>(null);

    // History arrays for sparkline charts
    let cpuHistory = $state<number[]>([]);
    let memoryHistory = $state<number[]>([]);
    let networkInHistory = $state<number[]>([]);
    let networkOutHistory = $state<number[]>([]);
    let diskReadHistory = $state<number[]>([]);
    let diskWriteHistory = $state<number[]>([]);

    // Navigation menu items with chart configuration
    interface NavItem {
        id: string;
        name: string;
        icon: any;
        color: string;
        hasChart: boolean; // Whether to show sparkline chart
    }

    const navItems: NavItem[] = [
        {
            id: "overview",
            name: "总览",
            icon: House,
            color: "#8b5cf6",
            hasChart: false,
        },
        {
            id: "processes",
            name: "进程",
            icon: Activity,
            color: "#6366f1",
            hasChart: false,
        },
        { id: "cpu", name: "CPU", icon: Cpu, color: "#10b981", hasChart: true },
        {
            id: "memory",
            name: "内存",
            icon: MemoryStick,
            color: "#ec4899",
            hasChart: true,
        },
        {
            id: "network",
            name: "网络",
            icon: Network,
            color: "#3b82f6",
            hasChart: true,
        },
        {
            id: "diskio",
            name: "磁盘 I/O",
            icon: HardDrive,
            color: "#f97316",
            hasChart: true,
        },
    ];

    // Get history data for a nav item
    function getHistoryData(id: string): number[] | null {
        switch (id) {
            case "cpu":
                return cpuHistory;
            case "memory":
                return memoryHistory;
            case "network":
                return networkInHistory;
            case "diskio":
                return diskReadHistory;
            default:
                return null;
        }
    }

    // Get color for nav item
    function getColor(id: string): string {
        return navItems.find((item) => item.id === id)?.color || "#3b82f6";
    }

    // Get current value display for nav item
    function getCurrentValue(id: string): string {
        if (!stats) return "";
        switch (id) {
            case "cpu":
                return `${stats.cpu.total.toFixed(0)}%`;
            case "memory":
                return `${stats.memory.percent.toFixed(0)}%`;
            case "network":
                const totalSpeedIn = stats.network.reduce(
                    (sum, n) => sum + n.speed_in,
                    0,
                );
                return formatBytes(totalSpeedIn) + "/s";
            case "diskio":
                const totalRead = stats.disk_io.reduce(
                    (sum, d) => sum + d.read_speed,
                    0,
                );
                return formatBytes(totalRead) + "/s";
            default:
                return "";
        }
    }

    onMount(() => {
        loadData();
        // Poll every 2 seconds for real-time updates
        const interval = setInterval(loadData, 1000);
        return () => clearInterval(interval);
    });

    async function loadData() {
        try {
            const [statsData, processData] = await Promise.all([
                api.getSystemStats().catch(() => null),
                currentView === "processes"
                    ? api.listProcesses().catch(() => null)
                    : Promise.resolve(null),
            ]);

            if (statsData) {
                stats = statsData;

                // Update history arrays
                cpuHistory = pushToHistory(cpuHistory, statsData.cpu.total);
                memoryHistory = pushToHistory(
                    memoryHistory,
                    statsData.memory.percent,
                );

                // Network: sum of all interfaces
                const totalSpeedIn = statsData.network.reduce(
                    (sum, n) => sum + n.speed_in,
                    0,
                );
                const totalSpeedOut = statsData.network.reduce(
                    (sum, n) => sum + n.speed_out,
                    0,
                );
                networkInHistory = pushToHistory(
                    networkInHistory,
                    totalSpeedIn / 1024 / 1024,
                ); // MB/s
                networkOutHistory = pushToHistory(
                    networkOutHistory,
                    totalSpeedOut / 1024 / 1024,
                );

                // Disk I/O: sum of all devices
                const totalRead = statsData.disk_io.reduce(
                    (sum, d) => sum + d.read_speed,
                    0,
                );
                const totalWrite = statsData.disk_io.reduce(
                    (sum, d) => sum + d.write_speed,
                    0,
                );
                diskReadHistory = pushToHistory(
                    diskReadHistory,
                    totalRead / 1024 / 1024,
                ); // MB/s
                diskWriteHistory = pushToHistory(
                    diskWriteHistory,
                    totalWrite / 1024 / 1024,
                );
            }

            processes = processData || [];
            loading = false;
            error = null;
        } catch (err) {
            console.error("Failed to load system stats:", err);
            error = "无法加载系统信息";
            loading = false;
        }
    }

    function pushToHistory(history: number[], value: number): number[] {
        const newHistory = [...history, value];
        if (newHistory.length > HISTORY_SIZE) {
            return newHistory.slice(-HISTORY_SIZE);
        }
        return newHistory;
    }

    // Switch view and immediately load data if needed
    async function switchView(viewId: string) {
        currentView = viewId;
        if (viewId === "processes" && processes.length === 0) {
            processes = await api.listProcesses().catch(() => []);
        }
    }

    async function handleKillProcess(
        pid: number,
        signal: "TERM" | "KILL" = "TERM",
    ) {
        try {
            await api.signalProcess(pid, signal);
            // Refresh process list
            processes = await api.listProcesses().catch(() => []);
        } catch (err) {
            console.error("Failed to kill process:", err);
        }
    }
</script>

<div class="flex h-full">
    <!-- Left Navigation Sidebar with Mini Charts -->
    <nav class="w-56 glass-card border-r border-border/50 flex flex-col">
        <div class="flex-1 overflow-y-auto p-2">
            {#each navItems as item}
                {@const historyData = getHistoryData(item.id) || []}
                <button
                    onclick={() => switchView(item.id)}
                    class="w-full text-left px-3 py-2 rounded-lg text-sm transition-all mb-1 {currentView ===
                    item.id
                        ? 'bg-primary/10 text-primary'
                        : 'text-muted-foreground hover:bg-white/5 hover:text-foreground'}"
                >
                    <div class="flex items-center gap-2 mb-1">
                        <item.icon class="w-4 h-4" />
                        <span class="font-medium">{item.name}</span>
                        {#if getCurrentValue(item.id)}
                            <span class="ml-auto text-xs opacity-70">
                                {getCurrentValue(item.id)}
                            </span>
                        {/if}
                    </div>
                    {#if item.hasChart}
                        <div class="mt-1">
                            <Sparkline
                                data={historyData}
                                height={32}
                                color={item.color}
                                strokeWidth={1.5}
                            />
                        </div>
                    {/if}
                </button>
            {/each}
        </div>
    </nav>

    <!-- Main Content Area -->
    <div class="flex-1 overflow-hidden flex flex-col">
        {#if loading}
            <div class="flex items-center justify-center h-full">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else if error}
            <div class="flex items-center justify-center h-full">
                <div class="text-center text-muted-foreground">
                    <Activity class="w-16 h-16 mx-auto mb-4 opacity-50" />
                    <p class="text-lg">{error}</p>
                </div>
            </div>
        {:else if stats}
            {#if currentView === "overview"}
                <OverviewView
                    {stats}
                    {cpuHistory}
                    {memoryHistory}
                    {networkInHistory}
                    {diskReadHistory}
                />
            {:else if currentView === "cpu"}
                <CpuView cpu={stats.cpu} history={cpuHistory} />
            {:else if currentView === "memory"}
                <MemoryView memory={stats.memory} history={memoryHistory} />
            {:else if currentView === "network"}
                <NetworkView
                    network={stats.network}
                    inHistory={networkInHistory}
                    outHistory={networkOutHistory}
                />
            {:else if currentView === "diskio"}
                <DiskIOView
                    diskIO={stats.disk_io}
                    readHistory={diskReadHistory}
                    writeHistory={diskWriteHistory}
                />
            {:else if currentView === "processes"}
                <ProcessesView {processes} onKill={handleKillProcess} />
            {/if}
        {/if}
    </div>
</div>
