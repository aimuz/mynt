<script lang="ts">
    import { api, type SnapshotPolicy, type StorageSpace } from "$lib/api";
    import { Plus, Trash2, CalendarClock } from "@lucide/svelte";

    interface Props {
        policies: SnapshotPolicy[];
        datasets: StorageSpace[];
        loading: boolean;
        onRefresh: () => void;
    }

    let { policies, datasets, loading, onRefresh }: Props = $props();

    let creating = $state(false);
    let showPolicyForm = $state(false);
    let nameError = $state("");
    let newPolicy = $state<Partial<SnapshotPolicy>>({
        name: "",
        schedule: "@daily",
        retention: "7d",
        datasets: [],
        enabled: true,
    });

    // Validate policy name: must start with letter, only letters, numbers, underscores, hyphens
    const policyNameRegex = /^[a-zA-Z][a-zA-Z0-9_-]*$/;

    function validateName(name: string): boolean {
        if (!name) {
            nameError = "";
            return false;
        }
        if (!policyNameRegex.test(name)) {
            nameError =
                "名称只能包含英文字母、数字、下划线和连字符，且必须以字母开头";
            return false;
        }
        nameError = "";
        return true;
    }

    function toggleDatasetSelection(datasetName: string) {
        if (!newPolicy.datasets) newPolicy.datasets = [];
        const index = newPolicy.datasets.indexOf(datasetName);
        if (index === -1) {
            newPolicy.datasets.push(datasetName);
        } else {
            newPolicy.datasets.splice(index, 1);
        }
        newPolicy.datasets = [...newPolicy.datasets];
    }

    async function handleCreatePolicy() {
        if (
            !newPolicy.name ||
            !newPolicy.schedule ||
            !newPolicy.retention ||
            !newPolicy.datasets?.length
        ) {
            alert("请填写所有必填字段");
            return;
        }

        if (!validateName(newPolicy.name)) {
            return;
        }

        try {
            creating = true;
            await api.createSnapshotPolicy(newPolicy);
            showPolicyForm = false;
            newPolicy = {
                name: "",
                schedule: "@daily",
                retention: "7d",
                datasets: [],
                enabled: true,
            };
            onRefresh();
        } catch (err) {
            console.error("Failed to create policy:", err);
            alert("创建策略失败: " + err);
        } finally {
            creating = false;
        }
    }

    async function handleDeletePolicy(id: number) {
        if (!confirm("确定要删除此策略吗？")) return;

        try {
            await api.deleteSnapshotPolicy(id);
            onRefresh();
        } catch (err) {
            console.error("Failed to delete policy:", err);
            alert("删除策略失败: " + err);
        }
    }

    async function handleTogglePolicy(policy: SnapshotPolicy) {
        try {
            await api.updateSnapshotPolicy(policy.id, {
                enabled: !policy.enabled,
            });
            onRefresh();
        } catch (err) {
            console.error("Failed to toggle policy:", err);
            alert("更新策略失败: " + err);
        }
    }
</script>

<div class="flex-1 flex flex-col">
    <div class="p-6 border-b border-border/50">
        <div class="flex items-center justify-between">
            <div>
                <h2 class="text-2xl font-bold text-foreground">策略管理</h2>
                <p class="text-sm text-muted-foreground mt-1">
                    配置自动快照计划和保留规则
                </p>
            </div>
            <button
                onclick={() => (showPolicyForm = !showPolicyForm)}
                class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl"
            >
                <Plus class="w-4 h-4" />
                新建策略
            </button>
        </div>

        {#if showPolicyForm}
            <div
                class="mt-6 p-6 glass-card rounded-xl border border-primary/20"
            >
                <h3 class="text-lg font-semibold mb-4">
                    {newPolicy.id ? "编辑策略" : "新建快照策略"}
                </h3>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div class="space-y-4">
                        <div>
                            <label
                                for="policy-name"
                                class="block text-sm font-medium text-foreground mb-1"
                                >策略名称</label
                            >
                            <input
                                id="policy-name"
                                bind:value={newPolicy.name}
                                oninput={() =>
                                    validateName(newPolicy.name || "")}
                                placeholder="例如：daily-backup"
                                pattern="^[a-zA-Z][a-zA-Z0-9_-]*$"
                                class="w-full px-3 py-2 bg-background border rounded-lg {nameError
                                    ? 'border-red-500'
                                    : 'border-border'}"
                            />
                            {#if nameError}
                                <p class="text-xs text-red-500 mt-1">
                                    {nameError}
                                </p>
                            {:else}
                                <p class="text-xs text-muted-foreground mt-1">
                                    只能包含英文字母、数字、下划线、连字符
                                </p>
                            {/if}
                        </div>
                        <div>
                            <label
                                for="policy-schedule"
                                class="block text-sm font-medium text-foreground mb-1"
                                >执行计划</label
                            >
                            <select
                                id="policy-schedule"
                                bind:value={newPolicy.schedule}
                                class="w-full px-3 py-2 bg-background border border-border rounded-lg"
                            >
                                <option value="@hourly">每小时 (@hourly)</option
                                >
                                <option value="@daily">每天 (@daily)</option>
                                <option value="@weekly">每周 (@weekly)</option>
                                <option value="@monthly">每月 (@monthly)</option
                                >
                                <option value="custom">自定义...</option>
                            </select>
                            {#if newPolicy.schedule === "custom"}
                                <input
                                    placeholder="0 * * * *"
                                    class="w-full mt-2 px-3 py-2 bg-background border border-border rounded-lg"
                                />
                            {/if}
                        </div>
                        <div>
                            <label
                                for="policy-retention"
                                class="block text-sm font-medium text-foreground mb-1"
                                >保留时间</label
                            >
                            <select
                                id="policy-retention"
                                bind:value={newPolicy.retention}
                                class="w-full px-3 py-2 bg-background border border-border rounded-lg"
                            >
                                <option value="24h">24 小时</option>
                                <option value="7d">7 天</option>
                                <option value="30d">30 天</option>
                                <option value="365d">1 年</option>
                                <option value="forever">永久保留</option>
                            </select>
                        </div>
                    </div>
                    <div class="space-y-4">
                        <label
                            for="policy-datasets"
                            class="block text-sm font-medium text-foreground mb-1"
                            >应用到存储空间</label
                        >
                        <div
                            class="h-48 overflow-y-auto border border-border rounded-lg p-2 space-y-1"
                        >
                            {#each datasets as dataset}
                                <label
                                    class="flex items-center gap-2 p-2 hover:bg-white/5 rounded cursor-pointer"
                                >
                                    <input
                                        type="checkbox"
                                        checked={newPolicy.datasets?.includes(
                                            dataset.name,
                                        )}
                                        onchange={() =>
                                            toggleDatasetSelection(
                                                dataset.name,
                                            )}
                                        class="rounded border-border bg-background text-primary focus:ring-primary"
                                    />
                                    <span class="text-sm">{dataset.name}</span>
                                </label>
                            {/each}
                        </div>
                    </div>
                </div>
                <div class="flex justify-end gap-3 mt-6">
                    <button
                        onclick={() => (showPolicyForm = false)}
                        class="px-4 py-2 border border-border rounded-lg hover:bg-white/5"
                    >
                        取消
                    </button>
                    <button
                        onclick={handleCreatePolicy}
                        disabled={creating}
                        class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 disabled:opacity-50"
                    >
                        {creating ? "保存中..." : "保存策略"}
                    </button>
                </div>
            </div>
        {/if}
    </div>

    <div class="flex-1 overflow-auto p-6">
        {#if loading}
            <div class="flex items-center justify-center h-64">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else if policies.length === 0}
            <div class="glass-card rounded-xl p-12 text-center">
                <CalendarClock
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    暂无快照策略
                </h3>
                <p class="text-sm text-muted-foreground mb-6">
                    创建策略以自动保护您的数据
                </p>
                <button
                    onclick={() => (showPolicyForm = true)}
                    class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                >
                    <Plus class="w-4 h-4" />
                    新建策略
                </button>
            </div>
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {#each policies as policy}
                    <div
                        class="glass-card rounded-xl p-6 hover:bg-white/5 transition-all"
                    >
                        <div class="flex justify-between items-start mb-4">
                            <div>
                                <h3 class="font-semibold text-lg">
                                    {policy.name}
                                </h3>
                                <div
                                    class="flex items-center gap-2 mt-1 text-sm text-muted-foreground"
                                >
                                    <span>{policy.schedule}</span>
                                </div>
                            </div>
                            <div class="flex items-center gap-2">
                                <!-- Toggle Switch -->
                                <button
                                    onclick={() => handleTogglePolicy(policy)}
                                    class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors {policy.enabled
                                        ? 'bg-green-500'
                                        : 'bg-muted'}"
                                    title={policy.enabled
                                        ? "点击禁用"
                                        : "点击启用"}
                                >
                                    <span
                                        class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform {policy.enabled
                                            ? 'translate-x-6'
                                            : 'translate-x-1'}"
                                    ></span>
                                </button>
                                <button
                                    onclick={() =>
                                        handleDeletePolicy(policy.id)}
                                    class="p-2 text-muted-foreground hover:text-red-500 transition-colors"
                                >
                                    <Trash2 class="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                        <div class="space-y-2 text-sm">
                            <div class="flex justify-between">
                                <span class="text-muted-foreground"
                                    >保留时间:</span
                                >
                                <span>{policy.retention}</span>
                            </div>
                            <div class="flex justify-between">
                                <span class="text-muted-foreground"
                                    >应用对象:</span
                                >
                                <span>{policy.datasets.length} 个存储空间</span>
                            </div>
                        </div>
                        {#if policy.datasets.length > 0}
                            <div class="mt-4 flex flex-wrap gap-1">
                                {#each policy.datasets.slice(0, 3) as ds}
                                    <span
                                        class="text-xs px-2 py-1 rounded bg-secondary text-secondary-foreground"
                                    >
                                        {ds}
                                    </span>
                                {/each}
                                {#if policy.datasets.length > 3}
                                    <span
                                        class="text-xs px-2 py-1 rounded bg-secondary text-secondary-foreground"
                                    >
                                        +{policy.datasets.length - 3}
                                    </span>
                                {/if}
                            </div>
                        {/if}
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
