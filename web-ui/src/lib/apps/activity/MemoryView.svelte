<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import type { MemStats } from "$lib/api";
    import { MemoryStick } from "@lucide/svelte";
    import Sparkline from "$lib/components/Sparkline.svelte";

    interface Props {
        memory: MemStats;
        history?: number[];
    }

    let { memory, history = [] }: Props = $props();

    // Calculate swap percentage
    let swapPercent = $derived(
        memory.swap_total > 0
            ? (memory.swap_used / memory.swap_total) * 100
            : 0,
    );
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">内存</h2>
        <p class="text-sm text-muted-foreground mt-1">RAM 和交换空间使用情况</p>
    </div>

    <!-- RAM Usage with History Chart -->
    <div class="glass-card rounded-xl p-6 mb-6 fade-in">
        <div class="flex items-center gap-4 mb-4">
            <div
                class="w-16 h-16 rounded-xl bg-linear-to-br from-purple-500 to-pink-600 flex items-center justify-center"
            >
                <MemoryStick class="w-8 h-8 text-white" />
            </div>
            <div class="flex-1">
                <div class="flex justify-between items-center mb-2">
                    <h3 class="text-lg font-semibold text-foreground">内存</h3>
                    <span class="text-3xl font-bold text-foreground"
                        >{memory.percent.toFixed(1)}%</span
                    >
                </div>
                <div class="w-full bg-muted rounded-full h-4">
                    <div
                        class="bg-linear-to-r from-purple-500 to-pink-600 h-4 rounded-full transition-all"
                        style="width: {Math.min(memory.percent, 100)}%"
                    ></div>
                </div>
                <div
                    class="flex justify-between mt-2 text-sm text-muted-foreground"
                >
                    <span>已用: {formatBytes(memory.used)}</span>
                    <span>总计: {formatBytes(memory.total)}</span>
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
                    color="#ec4899"
                    strokeWidth={2}
                />
            </div>
        {/if}
    </div>

    <!-- Memory Breakdown -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div
            class="glass-card rounded-xl p-4 fade-in"
            style="animation-delay: 50ms;"
        >
            <p class="text-xs text-muted-foreground">已用</p>
            <p class="text-xl font-bold text-foreground">
                {formatBytes(memory.used)}
            </p>
        </div>
        <div
            class="glass-card rounded-xl p-4 fade-in"
            style="animation-delay: 100ms;"
        >
            <p class="text-xs text-muted-foreground">可用</p>
            <p class="text-xl font-bold text-foreground">
                {formatBytes(memory.available)}
            </p>
        </div>
        <div
            class="glass-card rounded-xl p-4 fade-in"
            style="animation-delay: 150ms;"
        >
            <p class="text-xs text-muted-foreground">缓存</p>
            <p class="text-xl font-bold text-foreground">
                {formatBytes(memory.cached)}
            </p>
        </div>
        <div
            class="glass-card rounded-xl p-4 fade-in"
            style="animation-delay: 200ms;"
        >
            <p class="text-xs text-muted-foreground">缓冲</p>
            <p class="text-xl font-bold text-foreground">
                {formatBytes(memory.buffers)}
            </p>
        </div>
    </div>

    <!-- Swap Usage -->
    {#if memory.swap_total > 0}
        <div
            class="glass-card rounded-xl p-6 fade-in"
            style="animation-delay: 250ms;"
        >
            <h3 class="font-semibold text-foreground mb-4">交换空间</h3>
            <div class="w-full bg-muted rounded-full h-4 mb-2">
                <div
                    class="bg-linear-to-r from-amber-500 to-orange-600 h-4 rounded-full transition-all"
                    style="width: {Math.min(swapPercent, 100)}%"
                ></div>
            </div>
            <div class="flex justify-between text-sm text-muted-foreground">
                <span>已用: {formatBytes(memory.swap_used)}</span>
                <span>总计: {formatBytes(memory.swap_total)}</span>
                <span>{swapPercent.toFixed(1)}%</span>
            </div>
        </div>
    {/if}
</div>
