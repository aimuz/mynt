<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type Pool } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import { Database, Plus } from "@lucide/svelte";
    import CreatePoolWindow from "$lib/apps/CreatePoolWindow.svelte";
    import PoolDetailWindow from "$lib/apps/storage/PoolDetailWindow.svelte";
    import EmptyState from "$lib/components/EmptyState.svelte";

    // View state
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

    function handleOpenPoolDetail(pool: Pool) {
        desktop.openWindow(
            `pool-detail-${pool.name}`,
            `存储池详情 - ${pool.name}`,
            Database,
            () => ({
                component: PoolDetailWindow,
                props: { poolName: pool.name, onRefresh: loadData },
            }),
        );
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

<div class="p-6 overflow-auto h-full">
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
            <EmptyState
                icon={Database}
                title="暂无存储池"
                description="创建第一个存储池以开始使用"
                actionLabel="创建存储池"
                onAction={handleCreatePool}
            />
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {#each pools as pool, i}
                    {@const usagePercent =
                        pool.size > 0
                            ? (pool.allocated / pool.size) * 100
                            : 0}
                    <button
                        onclick={() => handleOpenPoolDetail(pool)}
                        class="glass-card rounded-xl p-6 fade-in hover:bg-white/5 transition-all cursor-pointer text-left w-full"
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
                                <span>{usagePercent.toFixed(1)}%</span>
                            </div>
                            <div class="w-full bg-muted rounded-full h-2">
                                <div
                                    class="bg-linear-to-r from-blue-500 to-purple-600 h-2 rounded-full transition-all"
                                    style="width: {usagePercent}%"
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
                    </button>
                {/each}
            </div>
        {/if}
    {/if}
</div>
