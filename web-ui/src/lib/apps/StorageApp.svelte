<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type Pool } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        Database,
        Plus,
        House,
        HardDrive,
        Disc,
        Camera,
        Activity,
    } from "@lucide/svelte";
    import CreatePoolWindow from "$lib/apps/CreatePoolWindow.svelte";
    import SnapshotView from "$lib/apps/storage/SnapshotView.svelte";
    import OverviewView from "$lib/apps/storage/OverviewView.svelte";

    // View state
    let currentView = $state<string>("overview");
    let pools = $state<Pool[]>([]);
    let loading = $state(true);

    // Get desktop context for window management
    const desktop = getContext<{
        openWindow: (
            id: string,
            title: string,
            icon: any,
            component: any,
        ) => void;
        closeWindow: (id: string) => void;
    }>("desktop");

    // Navigation menu items
    const navItems = [
        { id: "overview", name: "总览", icon: House },
        { id: "pools", name: "存储池", icon: Database },
        { id: "spaces", name: "存储空间", icon: HardDrive },
        { id: "disks", name: "磁盘", icon: Disc },
        { id: "snapshots", name: "快照", icon: Camera },
        { id: "tasks", name: "任务", icon: Activity },
    ];

    onMount(() => {
        loadData();
        const interval = setInterval(loadData, 30000);
        return () => clearInterval(interval);
    });

    async function loadData() {
        try {
            pools = (await api.listPools().catch(() => [])) || [];
            loading = false;
        } catch (err) {
            console.error("Failed to load data:", err);
            loading = false;
        }
    }

    function handleCreatePool() {
        desktop.openWindow("create-pool", "创建存储池", Database, () => ({
            component: CreatePoolWindow,
            props: { onRefreshStorage: loadData },
        }));
    }

    function getHealthBadgeColor(health: string): string {
        switch (health.toUpperCase()) {
            case "ONLINE":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "DEGRADED":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            case "OFFLINE":
            case "UNAVAIL":
            case "FAULTED":
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }
</script>

<div class="flex h-full">
    <!-- Left Navigation Sidebar -->
    <nav class="w-48 glass-card border-r border-border/50 flex flex-col">
        <div class="p-4 border-b border-border/50">
            <h2
                class="text-lg font-bold text-foreground flex items-center gap-2"
            >
                <Database class="w-5 h-5" />
                储存管理
            </h2>
        </div>

        <div class="flex-1 overflow-y-auto p-2">
            {#each navItems as item}
                <button
                    onclick={() => (currentView = item.id)}
                    class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-all {currentView ===
                    item.id
                        ? 'bg-primary/10 text-primary font-medium'
                        : 'text-muted-foreground hover:bg-white/5 hover:text-foreground'}"
                >
                    <svelte:component this={item.icon} class="w-4 h-4" />
                    {item.name}
                </button>
            {/each}
        </div>
    </nav>

    <!-- Main Content Area -->
    <div class="flex-1 overflow-hidden flex flex-col">
        {#if currentView === "pools"}
            <!-- Pools View -->
            <div class="p-6 overflow-auto flex-1">
                {#if loading}
                    <div class="flex items-center justify-center h-64">
                        <div
                            class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                        ></div>
                    </div>
                {:else}
                    <!-- Header -->
                    <div class="flex items-center justify-between mb-6">
                        <div>
                            <h2 class="text-2xl font-bold text-foreground">
                                存储池
                            </h2>
                            <p class="text-sm text-muted-foreground mt-1">
                                管理 ZFS 存储池
                            </p>
                        </div>
                        <button
                            onclick={handleCreatePool}
                            class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl"
                        >
                            <Plus class="w-4 h-4" />
                            创建存储池
                        </button>
                    </div>

                    <!-- Pools Grid -->
                    {#if pools.length === 0}
                        <div
                            class="glass-card rounded-xl p-12 text-center fade-in"
                        >
                            <Database
                                class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                            />
                            <h3
                                class="text-lg font-semibold text-foreground mb-2"
                            >
                                暂无存储池
                            </h3>
                            <p class="text-sm text-muted-foreground mb-6">
                                创建第一个存储池以开始使用
                            </p>
                            <button
                                onclick={handleCreatePool}
                                class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                            >
                                <Plus class="w-4 h-4" />
                                创建存储池
                            </button>
                        </div>
                    {:else}
                        <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                            {#each pools as pool, i}
                                <div
                                    class="glass-card rounded-xl p-6 fade-in hover:bg-white/5 transition-all cursor-pointer"
                                    style="animation-delay: {i * 50}ms;"
                                >
                                    <div
                                        class="flex items-start justify-between mb-4"
                                    >
                                        <div class="flex items-center gap-3">
                                            <div
                                                class="w-12 h-12 rounded-xl bg-linear-to-br from-purple-500 to-blue-600 flex items-center justify-center shadow-lg"
                                            >
                                                <Database
                                                    class="w-6 h-6 text-white"
                                                />
                                            </div>
                                            <div>
                                                <h3
                                                    class="font-semibold text-lg text-foreground"
                                                >
                                                    {pool.name}
                                                </h3>
                                                <span
                                                    class="text-xs px-2 py-0.5 rounded-full {getHealthBadgeColor(
                                                        pool.health,
                                                    )}"
                                                >
                                                    {pool.health}
                                                </span>
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Capacity Bar -->
                                    <div class="mb-3">
                                        <div
                                            class="flex justify-between text-xs text-muted-foreground mb-1"
                                        >
                                            <span>容量</span>
                                            <span
                                                >{(
                                                    (pool.allocated /
                                                        pool.size) *
                                                    100
                                                ).toFixed(1)}%</span
                                            >
                                        </div>
                                        <div
                                            class="w-full bg-muted rounded-full h-2"
                                        >
                                            <div
                                                class="bg-linear-to-r from-blue-500 to-purple-600 h-2 rounded-full transition-all"
                                                style="width: {(pool.allocated /
                                                    pool.size) *
                                                    100}%"
                                            ></div>
                                        </div>
                                    </div>

                                    <!-- Stats -->
                                    <div
                                        class="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-border/50"
                                    >
                                        <div>
                                            <p
                                                class="text-xs text-muted-foreground"
                                            >
                                                总容量
                                            </p>
                                            <p
                                                class="text-sm font-semibold text-foreground mt-0.5"
                                            >
                                                {formatBytes(pool.size)}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-xs text-muted-foreground"
                                            >
                                                已用
                                            </p>
                                            <p
                                                class="text-sm font-semibold text-foreground mt-0.5"
                                            >
                                                {formatBytes(pool.allocated)}
                                            </p>
                                        </div>
                                        <div>
                                            <p
                                                class="text-xs text-muted-foreground"
                                            >
                                                可用
                                            </p>
                                            <p
                                                class="text-sm font-semibold text-foreground mt-0.5"
                                            >
                                                {formatBytes(pool.free)}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            {/each}
                        </div>
                    {/if}
                {/if}
            </div>
        {:else if currentView === "overview"}
            <!-- Overview View -->
            <OverviewView />
        {:else if currentView === "spaces"}
            <!-- Storage Spaces Placeholder -->
            <div class="p-6 flex-1 flex items-center justify-center">
                <div class="text-center">
                    <HardDrive
                        class="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50"
                    />
                    <h3 class="text-lg font-semibold text-foreground mb-2">
                        存储空间
                    </h3>
                    <p class="text-sm text-muted-foreground">即将推出</p>
                </div>
            </div>
        {:else if currentView === "disks"}
            <!-- Disks Placeholder -->
            <div class="p-6 flex-1 flex items-center justify-center">
                <div class="text-center">
                    <Disc
                        class="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50"
                    />
                    <h3 class="text-lg font-semibold text-foreground mb-2">
                        磁盘与设备
                    </h3>
                    <p class="text-sm text-muted-foreground">即将推出</p>
                </div>
            </div>
        {:else if currentView === "snapshots"}
            <!-- Snapshots View -->
            <SnapshotView />
        {:else if currentView === "tasks"}
            <!-- Tasks Placeholder -->
            <div class="p-6 flex-1 flex items-center justify-center">
                <div class="text-center">
                    <Activity
                        class="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50"
                    />
                    <h3 class="text-lg font-semibold text-foreground mb-2">
                        任务与日志
                    </h3>
                    <p class="text-sm text-muted-foreground">即将推出</p>
                </div>
            </div>
        {/if}
    </div>
</div>
