<script lang="ts">
    import { onMount } from "svelte";
    import { api, type StorageSpace, type Snapshot } from "$lib/api";
    import { Camera, Plus, RotateCcw, Trash2, RefreshCw } from "@lucide/svelte";
    import { formatBytes } from "$lib/utils";

    let datasets = $state<StorageSpace[]>([]);
    let snapshots = $state<Snapshot[]>([]);
    let selectedDataset = $state<string>("");
    let loading = $state(true);
    let snapshotsLoading = $state(false);
    let creating = $state(false);
    let newSnapshotName = $state("");
    let showCreateForm = $state(false);

    onMount(async () => {
        await loadDatasets();
    });

    async function loadDatasets() {
        try {
            loading = true;
            datasets = (await api.listDatasets().catch(() => [])) || [];
            if (datasets.length > 0 && !selectedDataset) {
                selectedDataset = datasets[0].name;
                await loadSnapshots();
            }
        } catch (err) {
            console.error("Failed to load datasets:", err);
        } finally {
            loading = false;
        }
    }

    async function loadSnapshots() {
        if (!selectedDataset) return;

        try {
            snapshotsLoading = true;
            snapshots =
                (await api.listSnapshots(selectedDataset).catch(() => [])) ||
                [];
        } catch (err) {
            console.error("Failed to load snapshots:", err);
        } finally {
            snapshotsLoading = false;
        }
    }

    async function handleDatasetChange(datasetName: string) {
        selectedDataset = datasetName;
        await loadSnapshots();
    }

    async function handleCreateSnapshot() {
        if (!newSnapshotName.trim() || !selectedDataset) return;

        try {
            creating = true;
            await api.createSnapshot(selectedDataset, newSnapshotName);
            newSnapshotName = "";
            showCreateForm = false;
            await loadSnapshots();
        } catch (err) {
            console.error("Failed to create snapshot:", err);
            alert("创建快照失败: " + err);
        } finally {
            creating = false;
        }
    }

    async function handleDeleteSnapshot(snapshotName: string) {
        if (!confirm(`确定要删除快照 ${snapshotName} 吗？此操作不可恢复。`)) {
            return;
        }

        try {
            await api.deleteSnapshot(snapshotName);
            await loadSnapshots();
        } catch (err) {
            console.error("Failed to delete snapshot:", err);
            alert("删除快照失败: " + err);
        }
    }

    async function handleRollback(snapshotName: string) {
        if (
            !confirm(
                `警告：回滚到快照 ${snapshotName} 将丢失该快照之后的所有数据更改。\n\n确定要继续吗？`,
            )
        ) {
            return;
        }

        try {
            await api.rollbackSnapshot(snapshotName);
            alert("回滚成功！");
            await loadSnapshots();
        } catch (err) {
            console.error("Failed to rollback snapshot:", err);
            alert("回滚失败: " + err);
        }
    }

    function formatDate(dateStr: string): string {
        try {
            const date = new Date(dateStr);
            return date.toLocaleString("zh-CN", {
                year: "numeric",
                month: "2-digit",
                day: "2-digit",
                hour: "2-digit",
                minute: "2-digit",
            });
        } catch {
            return dateStr;
        }
    }

    function getSourceLabel(source: string): string {
        if (source === "manual") return "手动";
        if (source.startsWith("policy:"))
            return "策略: " + source.replace("policy:", "");
        return source;
    }
</script>

<div class="flex h-full">
    <!-- Left: Dataset Selector -->
    <div class="w-64 glass-card border-r border-border/50 flex flex-col">
        <div class="p-4 border-b border-border/50">
            <h3
                class="text-sm font-semibold text-foreground flex items-center gap-2"
            >
                <Camera class="w-4 h-4" />
                选择存储空间
            </h3>
        </div>

        {#if loading}
            <div class="p-4 text-center">
                <div
                    class="animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent mx-auto"
                ></div>
            </div>
        {:else if datasets.length === 0}
            <div class="p-4 text-center">
                <p class="text-sm text-muted-foreground">暂无存储空间</p>
            </div>
        {:else}
            <div class="flex-1 overflow-y-auto p-2">
                {#each datasets as dataset}
                    <button
                        onclick={() => handleDatasetChange(dataset.name)}
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
    </div>

    <!-- Right: Snapshots List -->
    <div class="flex-1 flex flex-col">
        <div class="p-6 border-b border-border/50">
            <div class="flex items-center justify-between">
                <div>
                    <h2 class="text-2xl font-bold text-foreground">快照管理</h2>
                    <p class="text-sm text-muted-foreground mt-1">
                        {selectedDataset || "请选择存储空间"}
                    </p>
                </div>
                <div class="flex gap-2">
                    <button
                        onclick={() => loadSnapshots()}
                        disabled={!selectedDataset}
                        class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <RefreshCw class="w-4 h-4" />
                        刷新
                    </button>
                    <button
                        onclick={() => (showCreateForm = !showCreateForm)}
                        disabled={!selectedDataset}
                        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <Plus class="w-4 h-4" />
                        创建快照
                    </button>
                </div>
            </div>

            {#if showCreateForm}
                <div class="mt-4 p-4 glass-card rounded-lg">
                    <div class="flex gap-2">
                        <input
                            bind:value={newSnapshotName}
                            placeholder="快照名称（如：before-update）"
                            class="flex-1 px-3 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                            onkeydown={(e) =>
                                e.key === "Enter" && handleCreateSnapshot()}
                        />
                        <button
                            onclick={handleCreateSnapshot}
                            disabled={creating || !newSnapshotName.trim()}
                            class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {creating ? "创建中..." : "创建"}
                        </button>
                        <button
                            onclick={() => {
                                showCreateForm = false;
                                newSnapshotName = "";
                            }}
                            class="px-4 py-2 border border-border rounded-lg hover:bg-white/5 transition-all"
                        >
                            取消
                        </button>
                    </div>
                </div>
            {/if}
        </div>

        <div class="flex-1 overflow-auto p-6">
            {#if snapshotsLoading}
                <div class="flex items-center justify-center h-64">
                    <div
                        class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                    ></div>
                </div>
            {:else if !selectedDataset}
                <div class="flex items-center justify-center h-64">
                    <div class="text-center">
                        <Camera
                            class="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50"
                        />
                        <p class="text-sm text-muted-foreground">
                            请从左侧选择一个存储空间
                        </p>
                    </div>
                </div>
            {:else if snapshots.length === 0}
                <div class="glass-card rounded-xl p-12 text-center">
                    <Camera
                        class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                    />
                    <h3 class="text-lg font-semibold text-foreground mb-2">
                        暂无快照
                    </h3>
                    <p class="text-sm text-muted-foreground mb-6">
                        为此存储空间创建第一个快照
                    </p>
                    <button
                        onclick={() => (showCreateForm = true)}
                        class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                    >
                        <Plus class="w-4 h-4" />
                        创建快照
                    </button>
                </div>
            {:else}
                <div class="space-y-2">
                    {#each snapshots as snapshot, i}
                        <div
                            class="glass-card rounded-lg p-4 fade-in hover:bg-white/5 transition-all"
                            style="animation-delay: {i * 30}ms;"
                        >
                            <div class="flex items-start justify-between">
                                <div class="flex-1">
                                    <div class="flex items-center gap-3 mb-2">
                                        <Camera class="w-5 h-5 text-primary" />
                                        <h4
                                            class="font-semibold text-foreground"
                                        >
                                            {snapshot.name.split("@")[1] ||
                                                snapshot.name}
                                        </h4>
                                        <span
                                            class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                                        >
                                            {getSourceLabel(snapshot.source)}
                                        </span>
                                    </div>
                                    <div class="grid grid-cols-3 gap-4 text-sm">
                                        <div>
                                            <span class="text-muted-foreground"
                                                >创建时间：</span
                                            >
                                            <span class="text-foreground"
                                                >{formatDate(
                                                    snapshot.created_at,
                                                )}</span
                                            >
                                        </div>
                                        <div>
                                            <span class="text-muted-foreground"
                                                >已用空间：</span
                                            >
                                            <span class="text-foreground"
                                                >{formatBytes(
                                                    snapshot.used,
                                                )}</span
                                            >
                                        </div>
                                        <div>
                                            <span class="text-muted-foreground"
                                                >引用空间：</span
                                            >
                                            <span class="text-foreground"
                                                >{formatBytes(
                                                    snapshot.referenced,
                                                )}</span
                                            >
                                        </div>
                                    </div>
                                </div>
                                <div class="flex gap-2 ml-4">
                                    <button
                                        onclick={() =>
                                            handleRollback(snapshot.name)}
                                        class="p-2 rounded-lg border border-border hover:bg-yellow-500/10 hover:border-yellow-500 hover:text-yellow-500 transition-all"
                                        title="回滚到此快照"
                                    >
                                        <RotateCcw class="w-4 h-4" />
                                    </button>
                                    <button
                                        onclick={() =>
                                            handleDeleteSnapshot(snapshot.name)}
                                        class="p-2 rounded-lg border border-border hover:bg-red-500/10 hover:border-red-500 hover:text-red-500 transition-all"
                                        title="删除快照"
                                    >
                                        <Trash2 class="w-4 h-4" />
                                    </button>
                                </div>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </div>
    </div>
</div>
