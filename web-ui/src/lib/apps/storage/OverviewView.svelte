<script lang="ts">
    import { onMount } from "svelte";
    import { api, type Pool, type SnapshotPolicy } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import { Database, HardDrive, Shield, TriangleAlert } from "@lucide/svelte";

    let pools = $state<Pool[]>([]);
    let policies = $state<SnapshotPolicy[]>([]);
    let loading = $state(true);

    // Computed aggregates
    $effect(() => {
        // This will recompute when pools changes
    });

    const stats = $derived({
        totalPools: pools.length,
        healthyPools: pools.filter((p) => p.health.toUpperCase() === "ONLINE")
            .length,
        degradedPools: pools.filter(
            (p) => p.health.toUpperCase() === "DEGRADED",
        ).length,
        faultedPools: pools.filter((p) =>
            ["OFFLINE", "UNAVAIL", "FAULTED"].includes(p.health.toUpperCase()),
        ).length,
        totalCapacity: pools.reduce((sum, p) => sum + p.size, 0),
        totalUsed: pools.reduce((sum, p) => sum + p.allocated, 0),
        totalFree: pools.reduce((sum, p) => sum + p.free, 0),
        usagePercent:
            pools.length > 0
                ? (pools.reduce((sum, p) => sum + p.allocated, 0) /
                      pools.reduce((sum, p) => sum + p.size, 0)) *
                  100
                : 0,
    });

    onMount(() => {
        loadData();
        const interval = setInterval(loadData, 30000);
        return () => clearInterval(interval);
    });

    async function loadData() {
        try {
            pools = (await api.listPools().catch(() => [])) || [];
            policies = (await api.listSnapshotPolicies().catch(() => [])) || [];
        } catch (err) {
            console.error("Failed to load data:", err);
        } finally {
            loading = false;
        }
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

    function getCapacityColor(percent: number): string {
        if (percent >= 90) return "from-red-500 to-orange-600";
        if (percent >= 80) return "from-yellow-500 to-orange-500";
        return "from-blue-500 to-purple-600";
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
        <!-- Page Header -->
        <div class="mb-6">
            <h2 class="text-2xl font-bold text-foreground">储存总览</h2>
            <p class="text-sm text-muted-foreground mt-1">
                查看所有存储资源的健康状态和容量使用情况
            </p>
        </div>

        <!-- Status Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
            <!-- Pool Health Card -->
            <div class="glass-card rounded-xl p-6 fade-in">
                <div class="flex items-center justify-between mb-4">
                    <div
                        class="w-12 h-12 rounded-xl bg-linear-to-br from-green-500 to-emerald-600 flex items-center justify-center shadow-lg"
                    >
                        <Database class="w-6 h-6 text-white" />
                    </div>
                    <div class="text-right">
                        <div class="text-3xl font-bold text-foreground">
                            {stats.totalPools}
                        </div>
                        <div class="text-xs text-muted-foreground">存储池</div>
                    </div>
                </div>
                <div class="space-y-2">
                    <div class="flex justify-between text-sm">
                        <span class="text-muted-foreground">健康</span>
                        <span
                            class="font-semibold text-green-600 dark:text-green-400"
                            >{stats.healthyPools}</span
                        >
                    </div>
                    {#if stats.degradedPools > 0}
                        <div class="flex justify-between text-sm">
                            <span class="text-muted-foreground">降级</span>
                            <span
                                class="font-semibold text-yellow-600 dark:text-yellow-400"
                                >{stats.degradedPools}</span
                            >
                        </div>
                    {/if}
                    {#if stats.faultedPools > 0}
                        <div class="flex justify-between text-sm">
                            <span class="text-muted-foreground">故障</span>
                            <span
                                class="font-semibold text-red-600 dark:text-red-400"
                                >{stats.faultedPools}</span
                            >
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Capacity Card -->
            <div
                class="glass-card rounded-xl p-6 fade-in"
                style="animation-delay: 50ms;"
            >
                <div class="flex items-center justify-between mb-4">
                    <div
                        class="w-12 h-12 rounded-xl bg-linear-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg"
                    >
                        <HardDrive class="w-6 h-6 text-white" />
                    </div>
                    <div class="text-right">
                        <div class="text-3xl font-bold text-foreground">
                            {stats.usagePercent.toFixed(1)}%
                        </div>
                        <div class="text-xs text-muted-foreground">已使用</div>
                    </div>
                </div>
                <div class="mb-3">
                    <div class="w-full bg-muted rounded-full h-2">
                        <div
                            class="bg-linear-to-r {getCapacityColor(
                                stats.usagePercent,
                            )} h-2 rounded-full transition-all"
                            style="width: {stats.usagePercent}%"
                        ></div>
                    </div>
                </div>
                <div class="space-y-2 text-sm">
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">总容量</span>
                        <span class="font-semibold text-foreground"
                            >{formatBytes(stats.totalCapacity)}</span
                        >
                    </div>
                    <div class="flex justify-between">
                        <span class="text-muted-foreground">可用</span>
                        <span class="font-semibold text-foreground"
                            >{formatBytes(stats.totalFree)}</span
                        >
                    </div>
                </div>
            </div>

            <!-- Protection Card -->
            <div
                class="glass-card rounded-xl p-6 fade-in"
                style="animation-delay: 100ms;"
            >
                <div class="flex items-center justify-between mb-4">
                    <div
                        class="w-12 h-12 rounded-xl bg-linear-to-br from-purple-500 to-pink-600 flex items-center justify-center shadow-lg"
                    >
                        <Shield class="w-6 h-6 text-white" />
                    </div>
                    <div class="text-right">
                        <div class="text-3xl font-bold text-foreground">
                            {policies.length}
                        </div>
                        <div class="text-xs text-muted-foreground">
                            快照策略
                        </div>
                    </div>
                </div>
                <div class="space-y-2">
                    <div class="flex justify-between text-sm">
                        <span class="text-muted-foreground">已启用</span>
                        <span
                            class="font-semibold text-green-600 dark:text-green-400"
                        >
                            {policies.filter((p) => p.enabled).length}
                        </span>
                    </div>
                    <div class="flex justify-between text-sm">
                        <span class="text-muted-foreground">已禁用</span>
                        <span class="font-semibold text-muted-foreground">
                            {policies.filter((p) => !p.enabled).length}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Alerts Section -->
        {#if stats.degradedPools > 0 || stats.faultedPools > 0 || stats.usagePercent >= 80}
            <div
                class="glass-card rounded-xl p-6 mb-6 border-l-4 border-yellow-500"
            >
                <div class="flex items-start gap-3">
                    <TriangleAlert
                        class="w-5 h-5 text-yellow-600 dark:text-yellow-400 mt-0.5"
                    />
                    <div class="flex-1">
                        <h3 class="font-semibold text-foreground mb-2">
                            需要注意
                        </h3>
                        <div class="space-y-1 text-sm">
                            {#if stats.faultedPools > 0}
                                <p class="text-red-600 dark:text-red-400">
                                    • {stats.faultedPools} 个存储池处于故障状态，请立即检查
                                </p>
                            {/if}
                            {#if stats.degradedPools > 0}
                                <p class="text-yellow-600 dark:text-yellow-400">
                                    • {stats.degradedPools} 个存储池已降级，建议尽快处理
                                </p>
                            {/if}
                            {#if stats.usagePercent >= 90}
                                <p class="text-red-600 dark:text-red-400">
                                    • 总容量使用率已达 {stats.usagePercent.toFixed(
                                        1,
                                    )}%，存储空间严重不足
                                </p>
                            {:else if stats.usagePercent >= 80}
                                <p class="text-yellow-600 dark:text-yellow-400">
                                    • 总容量使用率已达 {stats.usagePercent.toFixed(
                                        1,
                                    )}%，建议清理或扩容
                                </p>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Pools Section -->
        <div class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-foreground">存储池</h3>
            <span class="text-sm text-muted-foreground">{pools.length} 个</span>
        </div>

        {#if pools.length === 0}
            <div class="glass-card rounded-xl p-12 text-center">
                <Database
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    暂无存储池
                </h3>
                <p class="text-sm text-muted-foreground">
                    请前往"存储池"页面创建第一个存储池
                </p>
            </div>
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {#each pools as pool, i}
                    <div
                        class="glass-card rounded-xl p-6 fade-in hover:bg-white/5 transition-all cursor-pointer"
                        style="animation-delay: {i * 50}ms;"
                    >
                        <div class="flex items-start justify-between mb-4">
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-12 h-12 rounded-xl bg-linear-to-br from-purple-500 to-blue-600 flex items-center justify-center shadow-lg"
                                >
                                    <Database class="w-6 h-6 text-white" />
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
                                        (pool.allocated / pool.size) *
                                        100
                                    ).toFixed(1)}%</span
                                >
                            </div>
                            <div class="w-full bg-muted rounded-full h-2">
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
                                <p class="text-xs text-muted-foreground">
                                    总容量
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {formatBytes(pool.size)}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-muted-foreground">
                                    已用
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {formatBytes(pool.allocated)}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-muted-foreground">
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
