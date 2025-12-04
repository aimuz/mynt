<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        Disc,
        HardDrive,
        RefreshCw,
        TriangleAlert,
        CircleCheckBig,
    } from "@lucide/svelte";

    // Disk interface - matches backend disk info
    interface Disk {
        name: string;
        path: string;
        model: string;
        serial: string;
        size: number;
        type: string; // "HDD", "SSD", "NVMe"
        status: string; // "healthy", "degraded", "failed"
        temperature?: number;
        pool?: string; // Pool this disk belongs to
        vdev?: string; // VDev role in pool
    }

    let disks = $state<Disk[]>([]);
    let loading = $state(true);

    onMount(() => {
        loadData();
    });

    async function loadData() {
        loading = true;
        try {
            // TODO: API endpoint for disk listing
            // For now, show placeholder data
            disks = [];
        } catch (err) {
            console.error("Failed to load disks:", err);
        } finally {
            loading = false;
        }
    }

    function getStatusIcon(status: string) {
        switch (status) {
            case "healthy":
                return { icon: CircleCheckBig, color: "text-green-500" };
            case "degraded":
                return { icon: TriangleAlert, color: "text-yellow-500" };
            case "failed":
                return { icon: TriangleAlert, color: "text-red-500" };
            default:
                return { icon: Disc, color: "text-muted-foreground" };
        }
    }

    function getDiskTypeColor(type: string): string {
        switch (type) {
            case "NVMe":
                return "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400";
            case "SSD":
                return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400";
            case "HDD":
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }
</script>

<div class="p-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
        <div>
            <h2 class="text-2xl font-bold text-foreground">磁盘与设备</h2>
            <p class="text-sm text-muted-foreground mt-1">
                查看系统中的物理存储设备
            </p>
        </div>
        <button
            onclick={() => loadData()}
            class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all"
        >
            <RefreshCw class="w-4 h-4" />
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
        {:else if disks.length === 0}
            <!-- Placeholder for API not implemented -->
            <div class="glass-card rounded-xl p-12 text-center">
                <Disc
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    磁盘信息
                </h3>
                <p class="text-sm text-muted-foreground mb-4">
                    磁盘检测功能开发中
                </p>
                <div class="text-xs text-muted-foreground">
                    <p>将显示：</p>
                    <ul class="mt-2 space-y-1">
                        <li>• 物理磁盘列表（HDD/SSD/NVMe）</li>
                        <li>• 磁盘健康状态和 S.M.A.R.T. 信息</li>
                        <li>• 磁盘所属存储池和角色</li>
                        <li>• 温度和性能指标</li>
                    </ul>
                </div>
            </div>
        {:else}
            <!-- Disk List -->
            <div class="space-y-3">
                {#each disks as disk, i}
                    {@const statusInfo = getStatusIcon(disk.status)}
                    <div
                        class="glass-card rounded-lg p-4 fade-in hover:bg-white/5 transition-all"
                        style="animation-delay: {i * 30}ms;"
                    >
                        <div class="flex items-start justify-between">
                            <div class="flex items-center gap-4">
                                <div
                                    class="w-12 h-12 rounded-xl bg-linear-to-br from-slate-500 to-slate-700 flex items-center justify-center"
                                >
                                    <HardDrive class="w-6 h-6 text-white" />
                                </div>
                                <div>
                                    <div class="flex items-center gap-2 mb-1">
                                        <h4
                                            class="font-semibold text-foreground"
                                        >
                                            {disk.name}
                                        </h4>
                                        <span
                                            class="text-xs px-2 py-0.5 rounded-full {getDiskTypeColor(
                                                disk.type,
                                            )}"
                                        >
                                            {disk.type}
                                        </span>
                                    </div>
                                    <p class="text-sm text-muted-foreground">
                                        {disk.model}
                                    </p>
                                    <p class="text-xs text-muted-foreground">
                                        序列号: {disk.serial}
                                    </p>
                                </div>
                            </div>

                            <div class="text-right">
                                <div
                                    class="flex items-center gap-2 justify-end mb-1"
                                >
                                    <statusInfo.icon
                                        class="w-4 h-4 {statusInfo.color}"
                                    />
                                    <span
                                        class="text-sm font-medium text-foreground capitalize"
                                        >{disk.status}</span
                                    >
                                </div>
                                <p class="text-sm text-muted-foreground">
                                    {formatBytes(disk.size)}
                                </p>
                                {#if disk.temperature}
                                    <p class="text-xs text-muted-foreground">
                                        {disk.temperature}°C
                                    </p>
                                {/if}
                            </div>
                        </div>

                        {#if disk.pool}
                            <div class="mt-3 pt-3 border-t border-border/50">
                                <p class="text-sm text-muted-foreground">
                                    存储池：<span
                                        class="text-foreground font-medium"
                                        >{disk.pool}</span
                                    >
                                    {#if disk.vdev}
                                        <span class="mx-2">•</span>
                                        角色：<span
                                            class="text-foreground font-medium"
                                            >{disk.vdev}</span
                                        >
                                    {/if}
                                </p>
                            </div>
                        {/if}
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
