<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import Graph from "./Graph.svelte";

    let { history, stats } = $props<{
        history: { memory: number[] };
        stats: any;
    }>();
</script>

<div class="h-40 border-t border-border bg-muted/30 p-4 grid grid-cols-3 gap-6">
    <!-- Graph -->
    <div class="col-span-2 bg-background/50 rounded-lg border border-border/50 p-3 flex flex-col relative overflow-hidden">
            <div class="absolute inset-0 opacity-50">
            <Graph data={history.memory} color="#10b981" />
            </div>
            <div class="relative z-10 flex justify-between items-start">
                <span class="text-xs font-medium uppercase text-muted-foreground">
                    Memory Pressure
                </span>
            </div>
    </div>

    <!-- Stats Details -->
    <div class="flex flex-col gap-2 justify-center text-sm">
        <div class="flex justify-between">
            <span class="text-muted-foreground">Used Memory:</span>
            <span class="font-mono font-medium">{formatBytes(stats?.memory.used)}</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Cached Files:</span>
            <span class="font-mono font-medium">{formatBytes(stats?.memory.available)}</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Swap Used:</span>
            <span class="font-mono font-medium">{formatBytes(stats?.swap.used)}</span>
        </div>
        <div class="border-t border-border my-1"></div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Total Physical:</span>
            <span class="font-mono font-medium">{formatBytes(stats?.memory.total)}</span>
        </div>
    </div>
</div>
