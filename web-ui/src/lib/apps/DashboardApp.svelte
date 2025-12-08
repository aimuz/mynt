<script module lang="ts">
    import { Activity as AppIcon } from "@lucide/svelte";

    export function launch(api: any, component: any) {
        api.openWindow({
            id: "dashboard",
            title: "Dashboard",
            icon: AppIcon,
            component: component,
            width: 1024,
            height: 768
        });
    }
</script>

<script lang="ts">
    import { onMount } from "svelte";
    import { api, type Pool, type Disk, type Share } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import { HardDrive, Database, FolderOpen, Bell, Activity } from "@lucide/svelte";
    import Chart from "chart.js/auto";

    let pools = $state<Pool[]>([]);
    let disks = $state<Disk[]>([]);
    let shares = $state<Share[]>([]);
    let notifCount = $state({ unread: 0, total: 0 });
    let loading = $state(true);
    let capacityChart: Chart | null = null;

    onMount(() => {
        loadData();
        const interval = setInterval(loadData, 30000);
        return () => clearInterval(interval);
    });

    async function loadData() {
        try {
            const [poolsData, disksData, sharesData, notifData] =
                await Promise.all([
                    api.listPools().catch(() => []),
                    api.listDisks().catch(() => []),
                    api.listShares().catch(() => []),
                    api
                        .getNotificationCount()
                        .catch(() => ({ unread: 0, total: 0 })),
                ]);

            pools = poolsData;
            disks = disksData;
            shares = sharesData;
            notifCount = notifData;

            loading = false;

            // Update capacity chart
            updateCapacityChart();
        } catch (error) {
            console.error("Failed to load data:", error);
            loading = false;
        }
    }

    function updateCapacityChart() {
        const ctx = document.getElementById(
            "capacityChart",
        ) as HTMLCanvasElement;
        if (!ctx) return;

        const totalCapacity = pools.reduce((sum, p) => sum + p.size, 0);
        const totalUsed = pools.reduce((sum, p) => sum + p.allocated, 0);
        const totalFree = totalCapacity - totalUsed;

        if (capacityChart) {
            capacityChart.destroy();
        }

        capacityChart = new Chart(ctx, {
            type: "doughnut",
            data: {
                labels: ["Used", "Free"],
                datasets: [
                    {
                        data: [totalUsed, totalFree],
                        backgroundColor: [
                            "hsl(222.2 47.4% 11.2%)",
                            "hsl(210 40% 96.1%)",
                        ],
                        borderWidth: 0,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                cutout: "70%",
                plugins: {
                    legend: {
                        display: false,
                    },
                    tooltip: {
                        callbacks: {
                            label: (context) => {
                                const value = context.parsed;
                                return `${context.label}: ${formatBytes(value)}`;
                            },
                        },
                    },
                },
            },
        });
    }

    function getHealthColor(health: string): string {
        switch (health.toUpperCase()) {
            case "ONLINE":
                return "text-green-600 dark:text-green-400";
            case "DEGRADED":
                return "text-yellow-600 dark:text-yellow-400";
            case "OFFLINE":
            case "UNAVAIL":
                return "text-red-600 dark:text-red-400";
            default:
                return "text-gray-600 dark:text-gray-400";
        }
    }

    function getShareTypeColor(shareType: string): string {
        switch (shareType) {
            case "public":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "restricted":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }
</script>

<div class="p-6 h-full overflow-auto">
    {#if loading}
        <div class="flex items-center justify-center h-64">
            <div
                class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else}
        <!-- Stats Grid -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <div class="glass-card rounded-xl p-4 fade-in">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-medium text-muted-foreground">
                            Disks
                        </p>
                        <p class="text-2xl font-bold text-foreground mt-1">
                            {disks.length}
                        </p>
                    </div>
                    <div class="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                        <HardDrive
                            class="w-5 h-5 text-blue-600 dark:text-blue-400"
                        />
                    </div>
                </div>
            </div>

            <div
                class="glass-card rounded-xl p-4 fade-in"
                style="animation-delay: 0.1s;"
            >
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-medium text-muted-foreground">
                            Pools
                        </p>
                        <p class="text-2xl font-bold text-foreground mt-1">
                            {pools.length}
                        </p>
                    </div>
                    <div
                        class="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg"
                    >
                        <Database
                            class="w-5 h-5 text-purple-600 dark:text-purple-400"
                        />
                    </div>
                </div>
            </div>

            <div
                class="glass-card rounded-xl p-4 fade-in"
                style="animation-delay: 0.2s;"
            >
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-medium text-muted-foreground">
                            Shares
                        </p>
                        <p class="text-2xl font-bold text-foreground mt-1">
                            {shares.length}
                        </p>
                    </div>
                    <div
                        class="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg"
                    >
                        <FolderOpen
                            class="w-5 h-5 text-green-600 dark:text-green-400"
                        />
                    </div>
                </div>
            </div>

            <div
                class="glass-card rounded-xl p-4 fade-in"
                style="animation-delay: 0.3s;"
            >
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-xs font-medium text-muted-foreground">
                            Notifications
                        </p>
                        <p class="text-2xl font-bold text-foreground mt-1">
                            {notifCount.unread}
                        </p>
                    </div>
                    <div
                        class="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg"
                    >
                        <Bell
                            class="w-5 h-5 text-orange-600 dark:text-orange-400"
                        />
                    </div>
                </div>
            </div>
        </div>

        <!-- Capacity & Pools -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-4">
            <div class="glass-card rounded-xl p-4">
                <h3 class="text-sm font-semibold text-foreground mb-4">
                    Capacity
                </h3>
                <div class="relative h-48">
                    <canvas id="capacityChart"></canvas>
                </div>
            </div>

            <div class="glass-card rounded-xl p-4 lg:col-span-2">
                <h3 class="text-sm font-semibold text-foreground mb-4">
                    Storage Pools
                </h3>
                {#if pools.length === 0}
                    <div class="text-center py-8 text-muted-foreground text-sm">
                        <Database class="w-10 h-10 mx-auto mb-2 opacity-50" />
                        <p>No pools</p>
                    </div>
                {:else}
                    <div class="space-y-3">
                        {#each pools as pool}
                            <div class="border border-border/50 rounded-lg p-3">
                                <div
                                    class="flex items-center justify-between mb-2"
                                >
                                    <h4
                                        class="font-semibold text-sm text-foreground"
                                    >
                                        {pool.name}
                                    </h4>
                                    <span
                                        class="text-xs font-medium {getHealthColor(
                                            pool.health,
                                        )}"
                                    >
                                        {pool.health}
                                    </span>
                                </div>
                                <div class="w-full bg-muted rounded-full h-1.5">
                                    <div
                                        class="bg-primary h-1.5 rounded-full transition-all"
                                        style="width: {pool.capacity}%"
                                    ></div>
                                </div>
                                <div
                                    class="flex justify-between text-xs text-muted-foreground mt-1"
                                >
                                    <span
                                        >{formatBytes(pool.allocated)} used</span
                                    >
                                    <span>{formatBytes(pool.free)} free</span>
                                </div>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        </div>

        <!-- Shares -->
        <div class="glass-card rounded-xl p-4">
            <h3 class="text-sm font-semibold text-foreground mb-4">Shares</h3>
            {#if shares.length === 0}
                <div class="text-center py-8 text-muted-foreground text-sm">
                    <FolderOpen class="w-10 h-10 mx-auto mb-2 opacity-50" />
                    <p>No shares</p>
                </div>
            {:else}
                <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                    {#each shares.slice(0, 6) as share}
                        <div
                            class="border border-border/50 rounded-lg p-3 hover:bg-white/5 transition"
                        >
                            <div class="flex items-start justify-between mb-1">
                                <h4
                                    class="font-semibold text-sm text-foreground"
                                >
                                    {share.name}
                                </h4>
                                <span
                                    class="text-xs px-2 py-0.5 rounded-full {getShareTypeColor(
                                        share.share_type,
                                    )}"
                                >
                                    {share.share_type}
                                </span>
                            </div>
                            <p class="text-xs text-muted-foreground truncate">
                                {share.path}
                            </p>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    {/if}
</div>
