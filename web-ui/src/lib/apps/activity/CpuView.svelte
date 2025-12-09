<script lang="ts">
    import type { CPUStats } from "$lib/api";
    import { Cpu } from "@lucide/svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    interface Props {
        cpu: CPUStats;
        history?: number[];
    }

    let { cpu, history = [] }: Props = $props();
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">CPU</h2>
        <p class="text-sm text-muted-foreground mt-1">处理器使用情况</p>
    </div>

    <!-- Aggregate Usage with History Chart -->
    <div class="glass-card rounded-xl p-6 mb-6 fade-in">
        <div class="flex items-center gap-4 mb-4">
            <div
                class="w-16 h-16 rounded-xl bg-linear-to-br from-green-500 to-teal-600 flex items-center justify-center"
            >
                <Cpu class="w-8 h-8 text-white" />
            </div>
            <div class="flex-1">
                <div class="flex justify-between items-center mb-2">
                    <h3 class="text-lg font-semibold text-foreground">
                        总使用率
                    </h3>
                    <span class="text-3xl font-bold text-foreground"
                        >{cpu.total.toFixed(1)}%</span
                    >
                </div>
                <div class="w-full bg-muted rounded-full h-4">
                    <div
                        class="bg-linear-to-r from-green-500 to-teal-600 h-4 rounded-full transition-all"
                        style="width: {Math.min(cpu.total, 100)}%"
                    ></div>
                </div>
            </div>
        </div>

        <!-- History Chart -->
        {#if history.length > 1}
            <div class="mt-4 pt-4 border-t border-border/50">
                <p class="text-xs text-muted-foreground mb-2">
                    历史趋势 (最近 2 分钟)
                </p>
                <Sparkline
                    data={history}
                    height={100}
                    color="#10b981"
                    strokeWidth={2}
                />
            </div>
        {/if}

        <div class="grid grid-cols-3 gap-4 mt-4 pt-4 border-t border-border/50">
            <div>
                <p class="text-xs text-muted-foreground">核心数</p>
                <p class="text-lg font-semibold text-foreground">
                    {cpu.core_count}
                </p>
            </div>
            <div>
                <p class="text-xs text-muted-foreground">频率</p>
                <p class="text-lg font-semibold text-foreground">
                    {cpu.frequency.toFixed(0)} MHz
                </p>
            </div>
            <div>
                <p class="text-xs text-muted-foreground">温度</p>
                <p class="text-lg font-semibold text-foreground">
                    {cpu.temperature > 0
                        ? `${cpu.temperature.toFixed(0)}°C`
                        : "-"}
                </p>
            </div>
        </div>
    </div>

    <!-- Per-Core Usage -->
    {#if cpu.cores && cpu.cores.length > 0}
        <div
            class="glass-card rounded-xl p-6 fade-in"
            style="animation-delay: 50ms;"
        >
            <h3 class="font-semibold text-foreground mb-4">每核心使用率</h3>
            <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-3">
                {#each cpu.cores as usage, i}
                    <div class="text-center">
                        <div
                            class="relative w-full h-24 bg-muted rounded-lg overflow-hidden"
                        >
                            <div
                                class="absolute bottom-0 left-0 right-0 bg-linear-to-t from-green-500 to-teal-400 transition-all"
                                style="height: {Math.min(usage, 100)}%"
                            ></div>
                            <div
                                class="absolute inset-0 flex items-center justify-center"
                            >
                                <span
                                    class="text-sm font-medium text-foreground drop-shadow"
                                >
                                    {usage.toFixed(0)}%
                                </span>
                            </div>
                        </div>
                        <p class="text-xs text-muted-foreground mt-1">
                            Core {i}
                        </p>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>
