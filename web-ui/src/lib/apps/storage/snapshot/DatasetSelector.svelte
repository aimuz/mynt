<script lang="ts">
    import { type StorageSpace, type SnapshotPolicy } from "$lib/api";
    import { Camera } from "@lucide/svelte";
    import { formatBytes } from "$lib/utils";

    interface Props {
        datasets: StorageSpace[];
        policies: SnapshotPolicy[];
        selectedDataset: string;
        activeTab: "snapshots" | "policies";
        loading: boolean;
        onDatasetChange: (name: string) => void;
        onTabChange: (tab: "snapshots" | "policies") => void;
    }

    let {
        datasets,
        policies,
        selectedDataset,
        activeTab,
        loading,
        onDatasetChange,
        onTabChange,
    }: Props = $props();
</script>

<div class="w-64 glass-card border-r border-border/50 flex flex-col">
    <!-- Tab Switcher -->
    <div class="p-2 border-b border-border/50">
        <div class="flex bg-muted/50 p-1 rounded-lg">
            <button
                onclick={() => onTabChange("snapshots")}
                class="flex-1 text-xs font-medium py-1.5 rounded-md transition-all {activeTab ===
                'snapshots'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'}"
            >
                快照列表
            </button>
            <button
                onclick={() => onTabChange("policies")}
                class="flex-1 text-xs font-medium py-1.5 rounded-md transition-all {activeTab ===
                'policies'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'}"
            >
                策略管理
            </button>
        </div>
    </div>

    {#if activeTab === "snapshots"}
        <!-- Snapshot Tab: Dataset List -->
        {#if loading}
            <div class="p-4 text-center">
                <div
                    class="animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent mx-auto"
                ></div>
            </div>
        {:else if datasets.length === 0}
            <div class="flex-1 flex items-center justify-center p-4">
                <div class="text-center">
                    <Camera
                        class="w-10 h-10 mx-auto mb-2 text-muted-foreground opacity-50"
                    />
                    <p class="text-sm text-muted-foreground">暂无存储空间</p>
                </div>
            </div>
        {:else}
            <div class="flex-1 overflow-y-auto p-2">
                {#each datasets as dataset}
                    <button
                        onclick={() => onDatasetChange(dataset.name)}
                        class="w-full text-left px-3 py-2 rounded-lg text-sm transition-all mb-1 {selectedDataset ===
                        dataset.name
                            ? 'bg-primary/10 text-primary font-medium'
                            : 'text-muted-foreground hover:bg-white/5 hover:text-foreground'}"
                    >
                        <div class="font-medium truncate">{dataset.name}</div>
                        <div class="text-xs opacity-70 mt-0.5">
                            {formatBytes(dataset.used)} / {dataset.quota
                                ? formatBytes(dataset.quota)
                                : "无限制"}
                        </div>
                    </button>
                {/each}
            </div>
        {/if}
    {:else}
        <!-- Policy Tab: Summary Stats -->
        <div class="flex-1 overflow-y-auto p-4">
            <div class="text-xs text-muted-foreground mb-4">
                快照策略用于自动创建和管理快照的生命周期。
            </div>
            <div class="space-y-3">
                <div class="flex items-center justify-between text-sm">
                    <span class="text-muted-foreground">活跃策略</span>
                    <span class="font-medium text-green-500"
                        >{policies.filter((p) => p.enabled).length}</span
                    >
                </div>
                <div class="flex items-center justify-between text-sm">
                    <span class="text-muted-foreground">策略总数</span>
                    <span class="font-medium">{policies.length}</span>
                </div>
            </div>
        </div>
    {/if}
</div>
