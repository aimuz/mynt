<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import Graph from "./Graph.svelte";

    let { history, stats } = $props<{
        history: { cpu: number[] };
        stats: any;
    }>();
</script>

<div class="h-40 border-t border-border bg-muted/30 p-4 grid grid-cols-3 gap-6">
    <!-- Graph -->
    <div class="col-span-2 bg-background/50 rounded-lg border border-border/50 p-3 flex flex-col relative overflow-hidden">
            <div class="absolute inset-0 opacity-50">
            <Graph data={history.cpu} color="#3b82f6" />
            </div>
            <div class="relative z-10 flex justify-between items-start">
                <span class="text-xs font-medium uppercase text-muted-foreground">
                    CPU Load
                </span>
            </div>
    </div>

    <!-- Stats Details -->
    <div class="flex flex-col gap-2 justify-center text-sm">
        <div class="flex justify-between">
            <span class="text-muted-foreground">System:</span>
            <span class="font-mono font-medium">{stats?.cpu.total_usage.toFixed(1)}%</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Idle:</span>
            <span class="font-mono font-medium">{(100 - (stats?.cpu.total_usage || 0)).toFixed(1)}%</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Cores:</span>
            <span class="font-mono font-medium">{stats?.cpu.cores.length}</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Processes:</span>
            <span class="font-mono font-medium">{stats?.cpu.cores.length * 15}</span> <!-- Dummy for now -->
        </div>
    </div>
</div>
