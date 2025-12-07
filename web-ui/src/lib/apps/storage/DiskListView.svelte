<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type Disk } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        HardDrive,
        RefreshCw,
        CircleCheckBig,
        TriangleAlert,
        CircleAlert,
        CircleQuestionMark,
        Thermometer,
        Database,
    } from "@lucide/svelte";
    import DiskDetailWindow from "./DiskDetailWindow.svelte";

    let disks = $state<Disk[]>([]);
    let loading = $state(true);
    let error = $state<string | null>(null);

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
    });

    async function loadData() {
        loading = true;
        error = null;
        try {
            disks = (await api.listDisks()) || [];
        } catch (err) {
            console.error("Failed to load disks:", err);
            error = err instanceof Error ? err.message : "加载磁盘失败";
            disks = [];
        } finally {
            loading = false;
        }
    }

    function openDiskDetail(disk: Disk) {
        if (desktop) {
            desktop.openWindow(
                `disk-detail-${disk.name}`,
                `磁盘详情 - ${disk.name}`,
                HardDrive,
                () => ({
                    component: DiskDetailWindow,
                    props: { disk },
                }),
            );
        }
    }

    function getStatusIcon(status: string) {
        switch (status) {
            case "healthy":
                return { icon: CircleCheckBig, color: "text-green-500" };
            case "warning":
                return { icon: TriangleAlert, color: "text-yellow-500" };
            case "failed":
                return { icon: CircleAlert, color: "text-red-500" };
            default:
                return {
                    icon: CircleQuestionMark,
                    color: "text-muted-foreground",
                };
        }
    }

    function getSmartHealthBadge(health: string): string {
        switch (health) {
            case "good":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "warning":
                return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400";
            case "failed":
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }

    function getSmartHealthLabel(health: string): string {
        switch (health) {
            case "good":
                return "良好";
            case "warning":
                return "警告";
            case "failed":
                return "故障";
            default:
                return "未知";
        }
    }

    function getDiskTypeColor(type: string): string {
        switch (type) {
            case "NVMe":
                return "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400";
            case "SSD":
                return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400";
            case "HDD":
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
            case "USB":
                return "bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }

    function getStatusLabel(status: string): string {
        switch (status) {
            case "healthy":
                return "正常";
            case "warning":
                return "警告";
            case "failed":
                return "故障";
            default:
                return "未知";
        }
    }

    function getTemperatureColor(temp: number | undefined): string {
        if (!temp) return "text-muted-foreground";
        if (temp >= 60) return "text-red-500";
        if (temp >= 50) return "text-yellow-500";
        return "text-green-500";
    }
</script>

<div class="p-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
        <div>
            <h2 class="text-2xl font-bold text-foreground">储存设备</h2>
            <p class="text-sm text-muted-foreground mt-1">
                查看物理磁盘健康状态和 S.M.A.R.T. 信息
            </p>
        </div>
        <button
            onclick={() => loadData()}
            class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all"
            disabled={loading}
        >
            <RefreshCw class="w-4 h-4 {loading ? 'animate-spin' : ''}" />
            刷新
        </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto">
        {#if loading}
            <div class="flex items-center justify-center h-64">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else if error}
            <div class="glass-card rounded-xl p-8 text-center">
                <CircleAlert class="w-12 h-12 mx-auto mb-4 text-red-500" />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    加载失败
                </h3>
                <p class="text-sm text-muted-foreground mb-4">{error}</p>
                <button
                    onclick={() => loadData()}
                    class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90"
                >
                    重试
                </button>
            </div>
        {:else if disks.length === 0}
            <div class="glass-card rounded-xl p-12 text-center">
                <HardDrive
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    未检测到磁盘
                </h3>
                <p class="text-sm text-muted-foreground">
                    系统中没有可用的物理存储设备
                </p>
            </div>
        {:else}
            <!-- Disk Table -->
            <div class="glass-card rounded-xl overflow-hidden">
                <table class="w-full">
                    <thead class="bg-muted/50 border-b border-border">
                        <tr>
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >槽位</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >磁盘</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >类型</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >容量</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >所属池</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >状态</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >S.M.A.R.T.</th
                            >
                            <th
                                class="text-left px-4 py-3 text-xs font-semibold text-muted-foreground uppercase tracking-wider"
                                >温度</th
                            >
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-border">
                        {#each disks as disk, i}
                            {@const statusInfo = getStatusIcon(disk.status)}
                            <tr
                                class="hover:bg-white/5 cursor-pointer transition-colors fade-in"
                                style="animation-delay: {i * 30}ms;"
                                onclick={() => openDiskDetail(disk)}
                            >
                                <td class="px-4 py-4">
                                    <span class="text-sm text-muted-foreground">
                                        {disk.slot || "-"}
                                    </span>
                                </td>
                                <td class="px-4 py-4">
                                    <div class="flex items-center gap-3">
                                        <div
                                            class="w-10 h-10 rounded-lg bg-linear-to-br from-slate-500 to-slate-700 flex items-center justify-center flex-shrink-0"
                                        >
                                            <HardDrive
                                                class="w-5 h-5 text-white"
                                            />
                                        </div>
                                        <div class="min-w-0">
                                            <p
                                                class="font-medium text-foreground truncate"
                                            >
                                                {disk.model || disk.name}
                                            </p>
                                            <p
                                                class="text-xs text-muted-foreground truncate"
                                            >
                                                {disk.serial}
                                            </p>
                                        </div>
                                    </div>
                                </td>
                                <td class="px-4 py-4">
                                    <span
                                        class="text-xs px-2 py-1 rounded-full {getDiskTypeColor(
                                            disk.type,
                                        )}"
                                    >
                                        {disk.type}
                                    </span>
                                </td>
                                <td class="px-4 py-4">
                                    <span class="text-sm text-foreground">
                                        {formatBytes(disk.size)}
                                    </span>
                                </td>
                                <td class="px-4 py-4">
                                    {#if disk.pool}
                                        <div class="flex items-center gap-1.5">
                                            <Database
                                                class="w-3.5 h-3.5 text-primary"
                                            />
                                            <span
                                                class="text-sm text-foreground"
                                                >{disk.pool}</span
                                            >
                                        </div>
                                    {:else}
                                        <span
                                            class="text-sm text-muted-foreground"
                                            >-</span
                                        >
                                    {/if}
                                </td>
                                <td class="px-4 py-4">
                                    <div class="flex items-center gap-2">
                                        <statusInfo.icon
                                            class="w-4 h-4 {statusInfo.color}"
                                        />
                                        <span class="text-sm text-foreground"
                                            >{getStatusLabel(disk.status)}</span
                                        >
                                    </div>
                                </td>
                                <td class="px-4 py-4">
                                    <span
                                        class="text-xs px-2 py-1 rounded-full {getSmartHealthBadge(
                                            disk.smart_health,
                                        )}"
                                    >
                                        {getSmartHealthLabel(disk.smart_health)}
                                    </span>
                                </td>
                                <td class="px-4 py-4">
                                    {#if disk.temperature}
                                        <div
                                            class="flex items-center gap-1 {getTemperatureColor(
                                                disk.temperature,
                                            )}"
                                        >
                                            <Thermometer class="w-4 h-4" />
                                            <span class="text-sm font-medium"
                                                >{disk.temperature}°C</span
                                            >
                                        </div>
                                    {:else}
                                        <span
                                            class="text-sm text-muted-foreground"
                                            >-</span
                                        >
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>

            <!-- Summary -->
            <div
                class="mt-4 flex items-center gap-6 text-sm text-muted-foreground"
            >
                <span>共 {disks.length} 个磁盘</span>
                {#if disks.some((d) => d.status === "warning")}
                    <span class="flex items-center gap-1 text-yellow-500">
                        <TriangleAlert class="w-4 h-4" />
                        {disks.filter((d) => d.status === "warning").length} 个警告
                    </span>
                {/if}
                {#if disks.some((d) => d.status === "failed")}
                    <span class="flex items-center gap-1 text-red-500">
                        <CircleAlert class="w-4 h-4" />
                        {disks.filter((d) => d.status === "failed").length} 个故障
                    </span>
                {/if}
            </div>
        {/if}
    </div>
</div>
