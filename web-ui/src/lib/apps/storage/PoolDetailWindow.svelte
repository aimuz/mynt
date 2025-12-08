<script lang="ts">
    import { onMount, getContext } from "svelte";
    import {
        api,
        type Pool,
        type VDevDetail,
        type DiskDetail,
        type ResilverStatus,
    } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        Database,
        HardDrive,
        RefreshCw,
        TriangleAlert,
        CircleCheckBig,
        CircleX,
        Wrench,
        Loader,
        Lightbulb,
    } from "@lucide/svelte";
    import DiskSlotGrid from "./DiskSlotGrid.svelte";

    interface Props {
        poolName: string;
        onRefresh?: () => void;
    }

    let { poolName, onRefresh }: Props = $props();

    const desktop = getContext<{
        openWindow: (
            id: string,
            title: string,
            icon: any,
            component: any,
        ) => void;
        closeWindow: (id: string) => void;
    }>("desktop");

    let pool = $state<Pool | null>(null);
    let vdevs = $state<VDevDetail[]>([]);
    let resilverStatus = $state<ResilverStatus | null>(null);
    let loading = $state(true);
    let error = $state<string | null>(null);

    onMount(() => {
        loadData();
        // Poll for resilver status every 5 seconds if resilvering
        const interval = setInterval(() => {
            if (resilverStatus?.in_progress) {
                loadResilverStatus();
            }
        }, 5000);
        return () => clearInterval(interval);
    });

    async function loadData() {
        loading = true;
        error = null;
        try {
            const [poolData, vdevData] = await Promise.all([
                api.getPool(poolName),
                api.getPoolVDevs(poolName),
            ]);
            pool = poolData;
            vdevs = vdevData || [];
            await loadResilverStatus();
        } catch (err) {
            error = err instanceof Error ? err.message : "加载失败";
        } finally {
            loading = false;
        }
    }

    async function loadResilverStatus() {
        try {
            resilverStatus = await api.getResilverStatus(poolName);
        } catch {
            // Ignore errors for resilver status
        }
    }

    async function handleLocateDisk(disk: DiskDetail) {
        try {
            await api.locateDisk(disk.name, "on");
        } catch (err) {
            console.error("Failed to locate disk:", err);
        }
    }

    async function handleStartScrub() {
        try {
            await api.scrubPool(poolName);
            await loadData();
        } catch (err) {
            console.error("Failed to start scrub:", err);
        }
    }

    function openReplaceWizard(disk: DiskDetail) {
        const module = import("./DiskReplaceWizard.svelte");
        module.then((m) => {
            desktop?.openWindow(
                `replace-disk-${disk.name}`,
                `更换磁盘 - ${disk.name}`,
                Wrench,
                () => ({
                    component: m.default,
                    props: {
                        poolName,
                        faultedDisk: disk,
                        onComplete: () => {
                            loadData();
                            onRefresh?.();
                        },
                    },
                }),
            );
        });
    }

    function getHealthBadgeClass(health: string): string {
        switch (health?.toUpperCase()) {
            case "ONLINE":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "DEGRADED":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            case "FAULTED":
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }

    function getHealthIcon(health: string) {
        switch (health?.toUpperCase()) {
            case "ONLINE":
                return { icon: CircleCheckBig, class: "text-green-500" };
            case "DEGRADED":
                return { icon: TriangleAlert, class: "text-yellow-500" };
            case "FAULTED":
                return { icon: CircleX, class: "text-red-500" };
            default:
                return { icon: HardDrive, class: "text-muted-foreground" };
        }
    }

    function getRiskInfo(
        pool: Pool,
        vdevs: VDevDetail[],
    ): { level: string; description: string; recommendation: string } {
        const health = pool.health?.toUpperCase();

        if (health === "FAULTED") {
            return {
                level: "critical",
                description: "存储池已故障，可能有数据丢失风险！",
                recommendation: "请立即联系技术支持进行数据恢复。",
            };
        }

        if (health === "DEGRADED") {
            // Count failed disks
            let failedDisks = 0;
            for (const vdev of vdevs) {
                for (const disk of vdev.children) {
                    if (disk.status !== "ONLINE") {
                        failedDisks++;
                    }
                }
            }
            // Use redundancy property for accurate risk description
            let description = `存储池降级：${failedDisks} 块磁盘故障。`;
            if (pool.redundancy === 0) {
                description += " 冗余已耗尽，再有磁盘故障可能导致数据丢失！";
            } else {
                description += ` 仍可承受 ${pool.redundancy} 块磁盘故障。`;
            }
            return {
                level: pool.redundancy === 0 ? "critical" : "high",
                description: description,
                recommendation: "请尽快更换故障磁盘。",
            };
        }

        return {
            level: "low",
            description: "存储池健康运行中。",
            recommendation: "",
        };
    }

    // Format redundancy number to localized display string
    function formatRedundancy(redundancy: number | undefined): string {
        if (redundancy === undefined || redundancy <= 0) {
            return "无冗余";
        }
        return `可坏 ${redundancy} 盘`;
    }

    const healthInfo = $derived(pool ? getHealthIcon(pool.health) : null);
    const riskInfo = $derived(pool ? getRiskInfo(pool, vdevs) : null);
    const redundancyText = $derived(
        pool ? formatRedundancy(pool.redundancy) : "",
    );
</script>

<div class="p-6 h-full overflow-auto">
    {#if loading}
        <div class="flex items-center justify-center h-64">
            <div
                class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else if error}
        <div class="flex flex-col items-center justify-center h-64 gap-4">
            <CircleX class="w-16 h-16 text-red-500" />
            <p class="text-red-500">{error}</p>
            <button
                onclick={loadData}
                class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90"
            >
                重试
            </button>
        </div>
    {:else if pool}
        <!-- Health Alert Banner -->
        {#if riskInfo && riskInfo.level !== "low"}
            <div
                class="mb-6 p-4 rounded-xl border {riskInfo.level === 'critical'
                    ? 'bg-red-500/10 border-red-500/30'
                    : 'bg-yellow-500/10 border-yellow-500/30'}"
            >
                <div class="flex items-start gap-3">
                    <TriangleAlert
                        class="w-6 h-6 shrink-0 {riskInfo.level === 'critical'
                            ? 'text-red-500'
                            : 'text-yellow-500'}"
                    />
                    <div>
                        <h3 class="font-semibold text-foreground mb-1">
                            {riskInfo.level === "critical"
                                ? "⚠️ 严重警告"
                                : "⚠️ 警告"}
                        </h3>
                        <p class="text-sm text-foreground/80 mb-2">
                            {riskInfo.description}
                        </p>
                        {#if riskInfo.recommendation}
                            <p class="text-sm font-medium text-foreground">
                                建议：{riskInfo.recommendation}
                            </p>
                        {/if}
                    </div>
                </div>
            </div>
        {/if}

        <!-- Pool Info Header -->
        <div class="glass-card rounded-xl p-6 mb-6">
            <div class="flex items-start justify-between">
                <div class="flex items-center gap-4">
                    <div
                        class="w-16 h-16 rounded-xl bg-linear-to-br from-purple-500 to-blue-600 flex items-center justify-center shadow-lg"
                    >
                        <Database class="w-8 h-8 text-white" />
                    </div>
                    <div>
                        <h2 class="text-2xl font-bold text-foreground">
                            {pool.name}
                        </h2>
                        <div class="flex items-center gap-3 mt-1">
                            <span
                                class="px-2 py-0.5 rounded-full text-xs font-medium {getHealthBadgeClass(
                                    pool.health,
                                )}"
                            >
                                {pool.health}
                            </span>
                            {#if redundancyText}
                                <span class="text-sm text-muted-foreground">
                                    {redundancyText}
                                </span>
                            {/if}
                        </div>
                    </div>
                </div>
                <button
                    onclick={loadData}
                    class="p-2 rounded-lg hover:bg-white/10 transition-colors"
                    title="刷新"
                >
                    <RefreshCw
                        class="w-5 h-5 {loading ? 'animate-spin' : ''}"
                    />
                </button>
            </div>

            <!-- Capacity Bar -->
            <div class="mt-6">
                <div class="flex justify-between text-sm mb-2">
                    <span class="text-muted-foreground">容量使用</span>
                    <span class="text-foreground">
                        {formatBytes(pool.allocated)} / {formatBytes(pool.size)}
                    </span>
                </div>
                <div class="w-full bg-muted rounded-full h-3">
                    <div
                        class="h-3 rounded-full bg-linear-to-r from-blue-500 to-purple-600 transition-all"
                        style="width: {(pool.allocated / pool.size) * 100}%"
                    ></div>
                </div>
            </div>
        </div>

        <!-- Resilver Progress (if in progress) -->
        {#if resilverStatus?.in_progress}
            <div
                class="glass-card rounded-xl p-6 mb-6 border-l-4 border-blue-500"
            >
                <div class="flex items-center gap-3 mb-4">
                    <Loader class="w-6 h-6 text-blue-500 animate-spin" />
                    <h3 class="text-lg font-semibold text-foreground">
                        正在重建中...
                    </h3>
                </div>
                <div class="mb-3">
                    <div class="flex justify-between text-sm mb-2">
                        <span class="text-muted-foreground">重建进度</span>
                        <span class="text-foreground font-medium">
                            {resilverStatus.percent_done.toFixed(1)}%
                        </span>
                    </div>
                    <div class="w-full bg-muted rounded-full h-2">
                        <div
                            class="h-2 rounded-full bg-blue-500 transition-all"
                            style="width: {resilverStatus.percent_done}%"
                        ></div>
                    </div>
                </div>
                <div class="grid grid-cols-2 gap-4 text-sm">
                    <div>
                        <span class="text-muted-foreground">预计剩余时间：</span
                        >
                        <span class="text-foreground font-medium">
                            {resilverStatus.time_remaining || "计算中..."}
                        </span>
                    </div>
                    <div>
                        <span class="text-muted-foreground">重建速度：</span>
                        <span class="text-foreground font-medium">
                            {formatBytes(resilverStatus.rate)}/s
                        </span>
                    </div>
                </div>
                <div class="mt-4 p-3 bg-blue-500/10 rounded-lg">
                    <div class="flex items-start gap-2">
                        <Lightbulb
                            class="w-4 h-4 text-blue-400 shrink-0 mt-0.5"
                        />
                        <p class="text-xs text-blue-300">
                            重建期间请勿关闭 NAS 电源，性能可能会有所下降。
                        </p>
                    </div>
                </div>
            </div>
        {/if}

        <!-- Disk Slot Grid -->
        <div class="glass-card rounded-xl p-6 mb-6">
            <h3 class="text-lg font-semibold text-foreground mb-4">盘位布局</h3>
            <DiskSlotGrid
                {vdevs}
                onDiskClick={(disk) => {
                    if (disk.status !== "ONLINE") {
                        openReplaceWizard(disk);
                    }
                }}
                onLocateDisk={handleLocateDisk}
            />
        </div>

        <!-- Actions -->
        <div class="flex gap-3">
            <button
                onclick={handleStartScrub}
                disabled={pool.scrub_in_progress}
                class="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
                <RefreshCw
                    class="w-4 h-4 {pool.scrub_in_progress
                        ? 'animate-spin'
                        : ''}"
                />
                {pool.scrub_in_progress ? "正在 Scrub..." : "开始 Scrub"}
            </button>
        </div>
    {/if}
</div>
