<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import type { SystemStats } from "$lib/api";
    import { Cpu, MemoryStick, Network, HardDrive } from "@lucide/svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    interface Props {
        stats: SystemStats;
        cpuHistory?: number[];
        memoryHistory?: number[];
        networkInHistory?: number[];
        diskReadHistory?: number[];
    }

    let {
        stats,
        cpuHistory = [],
        memoryHistory = [],
        networkInHistory = [],
        diskReadHistory = [],
    }: Props = $props();

    function formatSpeed(bytesPerSec: number): string {
        if (bytesPerSec === 0) return "0 B/s";
        return formatBytes(bytesPerSec) + "/s";
    }
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">系统概览</h2>
        <p class="text-sm text-muted-foreground mt-1">实时系统资源监控</p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- CPU Card -->
        <div class="glass-card rounded-xl p-6 fade-in">
            <div class="flex items-center gap-3 mb-4">
                <div
                    class="w-10 h-10 rounded-lg bg-linear-to-br from-green-500 to-teal-600 flex items-center justify-center"
                >
                    <Cpu class="w-5 h-5 text-white" />
                </div>
                <div>
                    <h3 class="font-semibold text-foreground">CPU</h3>
                    <p class="text-xs text-muted-foreground">
                        {stats.cpu.core_count} 核心
                    </p>
                </div>
                <span class="ml-auto text-2xl font-bold text-foreground">
                    {stats.cpu.total.toFixed(1)}%
                </span>
            </div>
            {#if cpuHistory.length > 1}
                <Sparkline
                    data={cpuHistory}
                    height={60}
                    color="#10b981"
                    strokeWidth={2}
                />
            {:else}
                <div class="w-full bg-muted rounded-full h-3">
                    <div
                        class="bg-linear-to-r from-green-500 to-teal-600 h-3 rounded-full transition-all"
                        style="width: {Math.min(stats.cpu.total, 100)}%"
                    ></div>
                </div>
            {/if}
            <div
                class="flex justify-between mt-2 text-xs text-muted-foreground"
            >
                <span>{stats.cpu.frequency.toFixed(0)} MHz</span>
                {#if stats.cpu.temperature > 0}
                    <span>{stats.cpu.temperature.toFixed(0)}°C</span>
                {/if}
            </div>
        </div>

        <!-- Memory Card -->
        <div
            class="glass-card rounded-xl p-6 fade-in"
            style="animation-delay: 50ms;"
        >
            <div class="flex items-center gap-3 mb-4">
                <div
                    class="w-10 h-10 rounded-lg bg-linear-to-br from-purple-500 to-pink-600 flex items-center justify-center"
                >
                    <MemoryStick class="w-5 h-5 text-white" />
                </div>
                <div>
                    <h3 class="font-semibold text-foreground">内存</h3>
                    <p class="text-xs text-muted-foreground">
                        {formatBytes(stats.memory.used)} / {formatBytes(
                            stats.memory.total,
                        )}
                    </p>
                </div>
                <span class="ml-auto text-2xl font-bold text-foreground">
                    {stats.memory.percent.toFixed(1)}%
                </span>
            </div>
            {#if memoryHistory.length > 1}
                <Sparkline
                    data={memoryHistory}
                    height={60}
                    color="#ec4899"
                    strokeWidth={2}
                />
            {:else}
                <div class="w-full bg-muted rounded-full h-3">
                    <div
                        class="bg-linear-to-r from-purple-500 to-pink-600 h-3 rounded-full transition-all"
                        style="width: {Math.min(stats.memory.percent, 100)}%"
                    ></div>
                </div>
            {/if}
            <div
                class="flex justify-between mt-2 text-xs text-muted-foreground"
            >
                <span>可用: {formatBytes(stats.memory.available)}</span>
                <span>缓存: {formatBytes(stats.memory.cached)}</span>
            </div>
        </div>

        <!-- Network Card -->
        <div
            class="glass-card rounded-xl p-6 fade-in"
            style="animation-delay: 100ms;"
        >
            <div class="flex items-center gap-3 mb-4">
                <div
                    class="w-10 h-10 rounded-lg bg-linear-to-br from-blue-500 to-cyan-600 flex items-center justify-center"
                >
                    <Network class="w-5 h-5 text-white" />
                </div>
                <div>
                    <h3 class="font-semibold text-foreground">网络</h3>
                    <p class="text-xs text-muted-foreground">
                        {stats.network.filter((n) => n.is_up).length} 个活跃接口
                    </p>
                </div>
            </div>
            {#if networkInHistory.length > 1}
                <Sparkline
                    data={networkInHistory}
                    height={60}
                    color="#3b82f6"
                    strokeWidth={2}
                />
            {/if}
            <div class="space-y-2 mt-2">
                {#each stats.network
                    .filter((n) => n.is_up)
                    .slice(0, 2) as iface}
                    <div class="flex justify-between text-sm">
                        <span class="text-muted-foreground">{iface.name}</span>
                        <div class="flex gap-4">
                            <span class="text-green-400"
                                >↓ {formatSpeed(iface.speed_in)}</span
                            >
                            <span class="text-blue-400"
                                >↑ {formatSpeed(iface.speed_out)}</span
                            >
                        </div>
                    </div>
                {/each}
            </div>
        </div>

        <!-- Disk I/O Card -->
        <div
            class="glass-card rounded-xl p-6 fade-in"
            style="animation-delay: 150ms;"
        >
            <div class="flex items-center gap-3 mb-4">
                <div
                    class="w-10 h-10 rounded-lg bg-linear-to-br from-orange-500 to-red-600 flex items-center justify-center"
                >
                    <HardDrive class="w-5 h-5 text-white" />
                </div>
                <div>
                    <h3 class="font-semibold text-foreground">磁盘 I/O</h3>
                    <p class="text-xs text-muted-foreground">
                        {stats.disk_io.length} 个设备
                    </p>
                </div>
            </div>
            {#if diskReadHistory.length > 1}
                <Sparkline
                    data={diskReadHistory}
                    height={60}
                    color="#f97316"
                    strokeWidth={2}
                />
            {/if}
            <div class="space-y-2 mt-2">
                {#each stats.disk_io.slice(0, 3) as disk}
                    <div class="flex justify-between text-sm">
                        <span class="text-muted-foreground">{disk.device}</span>
                        <div class="flex gap-4">
                            <span class="text-green-400"
                                >R: {formatSpeed(disk.read_speed)}</span
                            >
                            <span class="text-blue-400"
                                >W: {formatSpeed(disk.write_speed)}</span
                            >
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    </div>
</div>
