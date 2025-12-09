<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import Graph from "./Graph.svelte";

    let { history, stats } = $props<{
        history: { networkRx: number[], networkTx: number[] };
        stats: any;
    }>();
</script>

<div class="h-40 border-t border-border bg-muted/30 p-4 grid grid-cols-3 gap-6">
    <!-- Graph -->
    <div class="col-span-2 bg-background/50 rounded-lg border border-border/50 p-3 flex flex-col relative overflow-hidden">
            <div class="absolute inset-0 opacity-50">
            <Graph data={history.networkRx} color="#8b5cf6" autoScale={true} />
            </div>
            <div class="relative z-10 flex justify-between items-start">
                <span class="text-xs font-medium uppercase text-muted-foreground">
                    Network (RX)
                </span>
            </div>
    </div>

    <!-- Stats Details -->
    <div class="flex flex-col gap-2 justify-center text-sm">
        <div class="flex justify-between">
            <span class="text-muted-foreground">RX Rate:</span>
            <span class="font-mono font-medium">{formatBytes(history.networkRx[history.networkRx.length - 1] || 0)}/s</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">TX Rate:</span>
            <span class="font-mono font-medium">{formatBytes(history.networkTx[history.networkTx.length - 1] || 0)}/s</span>
        </div>
        <div class="border-t border-border my-1"></div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Packets In:</span>
            <span class="font-mono font-medium">{stats?.network?.packets_recv || 0}</span>
        </div>
        <div class="flex justify-between">
            <span class="text-muted-foreground">Packets Out:</span>
            <span class="font-mono font-medium">{stats?.network?.packets_sent || 0}</span>
        </div>
         <div class="flex justify-between">
            <span class="text-muted-foreground">Total Data:</span>
            <span class="font-mono font-medium">{formatBytes((stats?.network?.bytes_recv || 0) + (stats?.network?.bytes_sent || 0))}</span>
        </div>
    </div>
</div>
