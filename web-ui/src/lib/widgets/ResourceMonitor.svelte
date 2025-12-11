<script lang="ts">
    import { onMount } from "svelte";
    import { Cpu, MemoryStick, Clock } from "@lucide/svelte";
    import { api, type SystemStats } from "$lib/api";
    import { formatBytes, formatUptime } from "$lib/utils";

    let stats = $state<SystemStats | null>(null);

    async function refreshStats() {
        try {
            stats = await api.getSystemStats();
        } catch (error) {
            console.error("Failed to load system stats:", error);
        }
    }

    onMount(() => {
        refreshStats();
        const interval = setInterval(refreshStats, 3000);
        return () => clearInterval(interval);
    });
</script>

<div class="space-y-4">
    {#if stats}
        <!-- Uptime -->
        <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-blue-500/20 flex items-center justify-center text-blue-500">
                <Clock class="w-4 h-4" />
            </div>
            <div>
                <div class="text-xs text-foreground/60">Uptime</div>
                <div class="font-medium text-sm">{formatUptime(stats.uptime)}</div>
            </div>
        </div>

        <!-- CPU -->
        <div>
            <div class="flex justify-between text-xs mb-1">
                <div class="flex items-center gap-1">
                    <Cpu class="w-3 h-3 text-foreground/70" />
                    <span class="text-foreground/70">CPU</span>
                </div>
                <div class="font-medium">
                    {Math.round(stats.cpu.total)}%
                    {#if stats.cpu.temperature > 0}
                        <span class="text-foreground/50 ml-1">({stats.cpu.temperature}°C)</span>
                    {/if}
                </div>
            </div>
            <div class="w-full bg-foreground/10 rounded-full h-1.5 overflow-hidden">
                <div
                    class="h-full transition-all duration-500 rounded-full"
                    class:bg-blue-500={stats.cpu.total < 60}
                    class:bg-yellow-500={stats.cpu.total >= 60 && stats.cpu.total < 85}
                    class:bg-red-500={stats.cpu.total >= 85}
                    style="width: {stats.cpu.total}%"
                ></div>
            </div>
        </div>

        <!-- Memory -->
        <div>
            <div class="flex justify-between text-xs mb-1">
                <div class="flex items-center gap-1">
                    <MemoryStick class="w-3 h-3 text-foreground/70" />
                    <span class="text-foreground/70">Memory</span>
                </div>
                <div class="font-medium">{Math.round(stats.memory.percent)}%</div>
            </div>
            <div class="w-full bg-foreground/10 rounded-full h-1.5 overflow-hidden">
                <div
                    class="h-full transition-all duration-500 rounded-full"
                    class:bg-purple-500={stats.memory.percent < 70}
                    class:bg-yellow-500={stats.memory.percent >= 70 && stats.memory.percent < 90}
                    class:bg-red-500={stats.memory.percent >= 90}
                    style="width: {stats.memory.percent}%"
                ></div>
            </div>
            <div class="flex justify-between text-xs mt-1 text-foreground/60">
                <span>{formatBytes(stats.memory.used)}</span>
                <span>{formatBytes(stats.memory.total)}</span>
            </div>
        </div>
    {:else}
        <div class="flex items-center justify-center py-8 text-foreground/50 text-sm">
            Loading stats...
        </div>
    {/if}
</div>
