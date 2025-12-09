<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import type { NetStats } from "$lib/api";
    import { Network, Wifi, ArrowDown, ArrowUp } from "@lucide/svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    interface Props {
        network: NetStats[];
        inHistory?: number[];
        outHistory?: number[];
    }

    let { network, inHistory = [], outHistory = [] }: Props = $props();

    function formatSpeed(bytesPerSec: number): string {
        if (bytesPerSec === 0) return "0 B/s";
        return formatBytes(bytesPerSec) + "/s";
    }

    // Calculate totals
    let totalSpeedIn = $derived(
        network.reduce((sum, n) => sum + n.speed_in, 0),
    );
    let totalSpeedOut = $derived(
        network.reduce((sum, n) => sum + n.speed_out, 0),
    );
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">网络</h2>
        <p class="text-sm text-muted-foreground mt-1">网络接口和流量监控</p>
    </div>

    <!-- Network Summary with History -->
    {#if inHistory.length > 1 || outHistory.length > 1}
        <div class="glass-card rounded-xl p-6 mb-6 fade-in">
            <div class="flex items-center gap-4 mb-4">
                <div
                    class="w-12 h-12 rounded-lg bg-linear-to-br from-blue-500 to-cyan-600 flex items-center justify-center"
                >
                    <Network class="w-6 h-6 text-white" />
                </div>
                <div class="flex-1">
                    <h3 class="font-semibold text-foreground">网络流量</h3>
                    <div class="flex gap-6 mt-1">
                        <span class="text-green-400"
                            >↓ {formatSpeed(totalSpeedIn)}</span
                        >
                        <span class="text-blue-400"
                            >↑ {formatSpeed(totalSpeedOut)}</span
                        >
                    </div>
                </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
                <div>
                    <p class="text-xs text-muted-foreground mb-1">
                        下载历史 (MB/s)
                    </p>
                    <Sparkline
                        data={inHistory}
                        height={80}
                        color="#22c55e"
                        strokeWidth={2}
                    />
                </div>
                <div>
                    <p class="text-xs text-muted-foreground mb-1">
                        上传历史 (MB/s)
                    </p>
                    <Sparkline
                        data={outHistory}
                        height={80}
                        color="#3b82f6"
                        strokeWidth={2}
                    />
                </div>
            </div>
        </div>
    {/if}

    {#if network.length === 0}
        <div class="flex items-center justify-center h-64">
            <div class="text-center text-muted-foreground">
                <Network class="w-16 h-16 mx-auto mb-4 opacity-50" />
                <p class="text-lg">未检测到网络接口</p>
            </div>
        </div>
    {:else}
        <div class="space-y-4">
            {#each network as iface, i}
                <div
                    class="glass-card rounded-xl p-6 fade-in"
                    style="animation-delay: {i * 50}ms;"
                >
                    <div class="flex items-center gap-4 mb-4">
                        <div
                            class="w-12 h-12 rounded-lg bg-linear-to-br from-blue-500 to-cyan-600 flex items-center justify-center"
                        >
                            {#if iface.name.startsWith("wl")}
                                <Wifi class="w-6 h-6 text-white" />
                            {:else}
                                <Network class="w-6 h-6 text-white" />
                            {/if}
                        </div>
                        <div class="flex-1">
                            <div class="flex items-center gap-2">
                                <h3 class="font-semibold text-foreground">
                                    {iface.name}
                                </h3>
                                <span
                                    class="px-2 py-0.5 text-xs rounded-full {iface.is_up
                                        ? 'bg-green-500/20 text-green-400'
                                        : 'bg-red-500/20 text-red-400'}"
                                >
                                    {iface.is_up ? "已连接" : "已断开"}
                                </span>
                            </div>
                            {#if iface.link_speed > 0}
                                <p class="text-xs text-muted-foreground">
                                    链路速度: {iface.link_speed} Mbps
                                </p>
                            {/if}
                        </div>
                    </div>

                    <div class="grid grid-cols-2 gap-4">
                        <!-- Download -->
                        <div
                            class="p-3 rounded-lg bg-green-500/10 border border-green-500/20"
                        >
                            <div class="flex items-center gap-2 mb-1">
                                <ArrowDown class="w-4 h-4 text-green-400" />
                                <span class="text-xs text-muted-foreground"
                                    >下载</span
                                >
                            </div>
                            <p class="text-lg font-bold text-green-400">
                                {formatSpeed(iface.speed_in)}
                            </p>
                            <p class="text-xs text-muted-foreground">
                                总计: {formatBytes(iface.bytes_in)}
                            </p>
                        </div>

                        <!-- Upload -->
                        <div
                            class="p-3 rounded-lg bg-blue-500/10 border border-blue-500/20"
                        >
                            <div class="flex items-center gap-2 mb-1">
                                <ArrowUp class="w-4 h-4 text-blue-400" />
                                <span class="text-xs text-muted-foreground"
                                    >上传</span
                                >
                            </div>
                            <p class="text-lg font-bold text-blue-400">
                                {formatSpeed(iface.speed_out)}
                            </p>
                            <p class="text-xs text-muted-foreground">
                                总计: {formatBytes(iface.bytes_out)}
                            </p>
                        </div>
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
