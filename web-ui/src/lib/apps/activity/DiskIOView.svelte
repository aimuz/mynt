<script lang="ts">
    import { formatBytes, formatSpeed } from "$lib/utils";
    import type { DiskIOStats } from "$lib/api";
    import { HardDrive, ArrowDown, ArrowUp } from "@lucide/svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    interface Props {
        diskIO: DiskIOStats[];
        readHistory?: number[];
        writeHistory?: number[];
    }

    let { diskIO, readHistory = [], writeHistory = [] }: Props = $props();

    // Calculate totals
    let totalRead = $derived(diskIO.reduce((sum, d) => sum + d.read_speed, 0));
    let totalWrite = $derived(
        diskIO.reduce((sum, d) => sum + d.write_speed, 0),
    );
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">磁盘 I/O</h2>
        <p class="text-sm text-muted-foreground mt-1">磁盘读写速度监控</p>
    </div>

    <!-- Disk I/O Summary with History -->
    {#if readHistory.length > 1 || writeHistory.length > 1}
        <div class="glass-card rounded-xl p-6 mb-6 fade-in">
            <div class="flex items-center gap-4 mb-4">
                <div
                    class="w-12 h-12 rounded-lg bg-linear-to-br from-orange-500 to-red-600 flex items-center justify-center"
                >
                    <HardDrive class="w-6 h-6 text-white" />
                </div>
                <div class="flex-1">
                    <h3 class="font-semibold text-foreground">磁盘活动</h3>
                    <div class="flex gap-6 mt-1">
                        <span class="text-green-400"
                            >读取: {formatSpeed(totalRead)}</span
                        >
                        <span class="text-blue-400"
                            >写入: {formatSpeed(totalWrite)}</span
                        >
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div>
                    <p class="text-xs text-muted-foreground mb-1">
                        读取历史 (MB/s)
                    </p>
                    <Sparkline
                        data={readHistory}
                        height={80}
                        color="#22c55e"
                        strokeWidth={2}
                    />
                </div>
                <div>
                    <p class="text-xs text-muted-foreground mb-1">
                        写入历史 (MB/s)
                    </p>
                    <Sparkline
                        data={writeHistory}
                        height={80}
                        color="#3b82f6"
                        strokeWidth={2}
                    />
                </div>
            </div>
        </div>
    {/if}

    {#if diskIO.length === 0}
        <div class="flex items-center justify-center h-64">
            <div class="text-center text-muted-foreground">
                <HardDrive class="w-16 h-16 mx-auto mb-4 opacity-50" />
                <p class="text-lg">未检测到磁盘设备</p>
            </div>
        </div>
    {:else}
        <div class="space-y-4">
            {#each diskIO as disk, i}
                <div
                    class="glass-card rounded-xl p-6 fade-in"
                    style="animation-delay: {i * 50}ms;"
                >
                    <div class="flex items-center gap-4 mb-4">
                        <div
                            class="w-12 h-12 rounded-lg bg-linear-to-br from-orange-500 to-red-600 flex items-center justify-center"
                        >
                            <HardDrive class="w-6 h-6 text-white" />
                        </div>
                        <div class="flex-1">
                            <h3 class="font-semibold text-foreground">
                                {disk.device}
                            </h3>
                            <p class="text-xs text-muted-foreground">
                                总读取: {formatBytes(disk.read_bytes)} | 总写入:
                                {formatBytes(disk.write_bytes)}
                            </p>
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                        <!-- Read -->
                        <div
                            class="p-3 rounded-lg bg-green-500/10 border border-green-500/20"
                        >
                            <div class="flex items-center gap-2 mb-1">
                                <ArrowDown class="w-4 h-4 text-green-400" />
                                <span class="text-xs text-muted-foreground"
                                    >读取</span
                                >
                            </div>
                            <p class="text-lg font-bold text-green-400">
                                {formatSpeed(disk.read_speed)}
                            </p>
                        </div>

                        <!-- Write -->
                        <div
                            class="p-3 rounded-lg bg-blue-500/10 border border-blue-500/20"
                        >
                            <div class="flex items-center gap-2 mb-1">
                                <ArrowUp class="w-4 h-4 text-blue-400" />
                                <span class="text-xs text-muted-foreground"
                                    >写入</span
                                >
                            </div>
                            <p class="text-lg font-bold text-blue-400">
                                {formatSpeed(disk.write_speed)}
                            </p>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
