<script lang="ts">
    import { onMount } from "svelte";
    import { HardDrive, Database, Activity } from "@lucide/svelte";
    import { api } from "$lib/api";
    import { formatBytes } from "$lib/utils";

    let stats = $state({
        disks: 0,
        pools: 0,
        shares: 0,
        totalCapacity: 0,
        usedCapacity: 0,
    });

    onMount(async () => {
        try {
            const [disks, pools, shares] = await Promise.all([
                api.listDisks().catch(() => []),
                api.listPools().catch(() => []),
                api.listShares().catch(() => []),
            ]);

            stats = {
                disks: disks.length,
                pools: pools.length,
                shares: shares.length,
                totalCapacity: pools.reduce((sum, p) => sum + p.size, 0),
                usedCapacity: pools.reduce((sum, p) => sum + p.allocated, 0),
            };
        } catch (error) {
            console.error("Failed to load stats:", error);
        }
    });

    $effect(() => {
        const interval = setInterval(async () => {
            // Refresh every 30s
            try {
                const [disks, pools] = await Promise.all([
                    api.listDisks().catch(() => []),
                    api.listPools().catch(() => []),
                ]);

                stats.disks = disks.length;
                stats.pools = pools.length;
                stats.totalCapacity = pools.reduce((sum, p) => sum + p.size, 0);
                stats.usedCapacity = pools.reduce(
                    (sum, p) => sum + p.allocated,
                    0,
                );
            } catch (error) {
                console.error("Failed to refresh stats:", error);
            }
        }, 30000);

        return () => clearInterval(interval);
    });

    const usagePercent = $derived(
        stats.totalCapacity > 0
            ? Math.round((stats.usedCapacity / stats.totalCapacity) * 100)
            : 0,
    );
</script>

<div class="space-y-3">
    <!-- Capacity Bar -->
    <div>
        <div class="flex justify-between text-xs mb-1">
            <span class="text-foreground/70">Storage</span>
            <span class="font-medium">{usagePercent}%</span>
        </div>
        <div class="w-full bg-foreground/10 rounded-full h-1.5 overflow-hidden">
            <div
                class="gradient-to-r from-blue-500 to-purple-600 h-full transition-all duration-500 rounded-full"
                style="width: {usagePercent}%"
            ></div>
        </div>
        <div class="flex justify-between text-xs mt-1 text-foreground/60">
            <span>{formatBytes(stats.usedCapacity)}</span>
            <span>{formatBytes(stats.totalCapacity)}</span>
        </div>
    </div>

    <!-- Quick Stats -->
    <div class="grid grid-cols-3 gap-2 pt-2 border-t border-foreground/10">
        <div class="text-center">
            <HardDrive class="w-4 h-4 mx-auto mb-1 text-blue-500" />
            <div class="text-xl font-bold">{stats.disks}</div>
            <div class="text-xs text-foreground/60">Disks</div>
        </div>

        <div class="text-center">
            <Database class="w-4 h-4 mx-auto mb-1 text-purple-500" />
            <div class="text-xl font-bold">{stats.pools}</div>
            <div class="text-xs text-foreground/60">Pools</div>
        </div>

        <div class="text-center">
            <Activity class="w-4 h-4 mx-auto mb-1 text-green-500" />
            <div class="text-xl font-bold">{stats.shares}</div>
            <div class="text-xs text-foreground/60">Shares</div>
        </div>
    </div>
</div>
