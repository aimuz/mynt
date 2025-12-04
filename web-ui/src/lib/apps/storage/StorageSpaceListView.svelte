<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type StorageSpace } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        HardDrive,
        Plus,
        Trash2,
        Settings,
        RefreshCw,
    } from "@lucide/svelte";
    import CreateStorageSpaceWindow from "./CreateStorageSpaceWindow.svelte";

    let datasets = $state<StorageSpace[]>([]);
    let loading = $state(true);
    let filterPool = $state<string>("");
    let filterType = $state<string>("");

    // Get desktop context for window management
    const desktop = getContext<{
        openWindow: (
            id: string,
            title: string,
            icon: any,
            component: any,
        ) => void;
        closeWindow: (id: string) => void;
    }>("desktop");

    // Get unique pool names for filter
    const poolOptions = $derived(
        Array.from(new Set(datasets.map((d) => d.pool))).sort(),
    );

    // Filtered datasets
    const filteredDatasets = $derived(
        datasets.filter((d) => {
            if (filterPool && d.pool !== filterPool) return false;
            if (filterType && d.type !== filterType) return false;
            return true;
        }),
    );

    onMount(() => {
        loadData();
    });

    async function loadData() {
        try {
            loading = true;
            datasets = (await api.listDatasets().catch(() => [])) || [];
        } catch (err) {
            console.error("Failed to load datasets:", err);
        } finally {
            loading = false;
        }
    }

    function handleCreateSpace() {
        desktop.openWindow(
            "create-storage-space",
            "创建存储空间",
            HardDrive,
            () => ({
                component: CreateStorageSpaceWindow,
                props: {
                    onRefresh: loadData,
                    onClose: () => desktop.closeWindow("create-storage-space"),
                },
            }),
        );
    }

    async function handleDelete(datasetName: string) {
        if (
            !confirm(
                `确定要删除存储空间 "${datasetName}" 吗？\n\n此操作将删除所有数据和快照，无法恢复！`,
            )
        ) {
            return;
        }

        try {
            await api.deleteDataset(datasetName);
            await loadData();
        } catch (err) {
            console.error("Failed to delete dataset:", err);
            alert("删除失败: " + err);
        }
    }

    function getTypeLabel(type: string): string {
        return type === "filesystem" ? "文件系统" : "块设备";
    }

    function getUsagePercent(dataset: StorageSpace): number {
        if (!dataset.quota) return 0;
        return (dataset.used / dataset.quota) * 100;
    }
</script>

<div class="p-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
        <div>
            <h2 class="text-2xl font-bold text-foreground">存储空间</h2>
            <p class="text-sm text-muted-foreground mt-1">
                管理 ZFS 数据集和卷
            </p>
        </div>
        <div class="flex gap-2">
            <button
                onclick={() => loadData()}
                class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all"
            >
                <RefreshCw class="w-4 h-4" />
                刷新
            </button>
            <button
                onclick={handleCreateSpace}
                class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl"
            >
                <Plus class="w-4 h-4" />
                创建存储空间
            </button>
        </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-3 mb-4">
        <select
            bind:value={filterPool}
            class="px-3 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        >
            <option value="">所有存储池</option>
            {#each poolOptions as pool}
                <option value={pool}>{pool}</option>
            {/each}
        </select>

        <select
            bind:value={filterType}
            class="px-3 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        >
            <option value="">所有类型</option>
            <option value="filesystem">文件系统</option>
            <option value="volume">块设备</option>
        </select>

        {#if filterPool || filterType}
            <button
                onclick={() => {
                    filterPool = "";
                    filterType = "";
                }}
                class="px-3 py-2 text-sm text-muted-foreground hover:text-foreground transition-all"
            >
                清除筛选
            </button>
        {/if}
    </div>

    <!-- Dataset List -->
    <div class="flex-1 overflow-auto">
        {#if loading}
            <div class="flex items-center justify-center h-64">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else if filteredDatasets.length === 0}
            <div class="glass-card rounded-xl p-12 text-center">
                <HardDrive
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    {datasets.length === 0 ? "暂无存储空间" : "无匹配结果"}
                </h3>
                <p class="text-sm text-muted-foreground mb-6">
                    {datasets.length === 0
                        ? "创建第一个存储空间以开始使用"
                        : "尝试调整筛选条件"}
                </p>
                {#if datasets.length === 0}
                    <button
                        onclick={handleCreateSpace}
                        class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                    >
                        <Plus class="w-4 h-4" />
                        创建存储空间
                    </button>
                {/if}
            </div>
        {:else}
            <div class="space-y-3">
                {#each filteredDatasets as dataset, i}
                    <div
                        class="glass-card rounded-lg p-4 fade-in hover:bg-white/5 transition-all"
                        style="animation-delay: {i * 30}ms;"
                    >
                        <div class="flex items-start justify-between">
                            <div class="flex-1">
                                <div class="flex items-center gap-3 mb-2">
                                    <HardDrive class="w-5 h-5 text-primary" />
                                    <h4 class="font-semibold text-foreground">
                                        {dataset.name}
                                    </h4>
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
                                    >
                                        {getTypeLabel(dataset.type)}
                                    </span>
                                </div>

                                <div class="grid grid-cols-4 gap-4 text-sm">
                                    <div>
                                        <span class="text-muted-foreground"
                                            >存储池：</span
                                        >
                                        <span
                                            class="text-foreground font-medium"
                                            >{dataset.pool}</span
                                        >
                                    </div>
                                    <div>
                                        <span class="text-muted-foreground"
                                            >已用：</span
                                        >
                                        <span
                                            class="text-foreground font-medium"
                                            >{formatBytes(dataset.used)}</span
                                        >
                                    </div>
                                    <div>
                                        <span class="text-muted-foreground"
                                            >可用：</span
                                        >
                                        <span
                                            class="text-foreground font-medium"
                                            >{formatBytes(
                                                dataset.available,
                                            )}</span
                                        >
                                    </div>
                                    <div>
                                        <span class="text-muted-foreground"
                                            >配额：</span
                                        >
                                        <span
                                            class="text-foreground font-medium"
                                        >
                                            {dataset.quota
                                                ? formatBytes(dataset.quota)
                                                : "无限制"}
                                        </span>
                                    </div>
                                </div>

                                {#if dataset.quota}
                                    <div class="mt-3">
                                        <div
                                            class="flex justify-between text-xs text-muted-foreground mb-1"
                                        >
                                            <span>配额使用</span>
                                            <span
                                                >{getUsagePercent(
                                                    dataset,
                                                ).toFixed(1)}%</span
                                            >
                                        </div>
                                        <div
                                            class="w-full bg-muted rounded-full h-1.5"
                                        >
                                            <div
                                                class="bg-linear-to-r from-blue-500 to-purple-600 h-1.5 rounded-full transition-all"
                                                style="width: {getUsagePercent(
                                                    dataset,
                                                )}%"
                                            ></div>
                                        </div>
                                    </div>
                                {/if}
                            </div>

                            <div class="flex gap-2 ml-4">
                                <button
                                    onclick={() => handleDelete(dataset.name)}
                                    class="p-2 rounded-lg border border-border hover:bg-red-500/10 hover:border-red-500 hover:text-red-500 transition-all"
                                    title="删除"
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
