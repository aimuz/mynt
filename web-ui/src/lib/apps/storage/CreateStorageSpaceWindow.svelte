<script lang="ts">
    import { api, type Pool, type CreateDatasetRequest } from "$lib/api";
    import { onMount, getContext } from "svelte";
    import {
        X,
        ChevronRight,
        ChevronLeft,
        Check,
        Database,
        Plus,
    } from "@lucide/svelte";
    import CreatePoolWindow from "$lib/apps/CreatePoolWindow.svelte";
    import EmptyState from "$lib/components/EmptyState.svelte";

    let {
        onRefresh = () => {},
        onClose = () => {},
    }: { onRefresh?: () => void; onClose?: () => void } = $props();

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

    let currentStep = $state(1);
    let pools = $state<Pool[]>([]);
    let loading = $state(true);
    let creating = $state(false);

    function handleCreatePool() {
        if (desktop) {
            desktop.openWindow("create-pool", "创建存储池", Database, () => ({
                component: CreatePoolWindow,
                props: { onRefresh: loadPools },
            }));
        }
    }

    // Form data
    let formData = $state<CreateDatasetRequest>({
        name: "",
        type: "filesystem",
        use_case: "general",
        quota_mode: "flexible",
        quota: 0,
        properties: {},
    });

    let selectedPool = $state("");
    let datasetName = $state("");
    let quotaGB = $state(0);

    // Use case options with Chinese labels
    const useCases = [
        { value: "general", label: "通用", desc: "适用于一般文件存储" },
        { value: "media", label: "影音库", desc: "优化大文件读写性能" },
        { value: "surveillance", label: "监控录像", desc: "视频文件存储优化" },
        { value: "vm", label: "虚拟机", desc: "随机I/O优化" },
        { value: "database", label: "数据库", desc: "数据完整性优先" },
    ];

    onMount(async () => {
        await loadPools();
    });

    async function loadPools() {
        try {
            pools = (await api.listPools().catch(() => [])) || [];
            // Don't auto-select - user chooses in Step 1
        } catch (err) {
            console.error("Failed to load pools:", err);
        } finally {
            loading = false;
        }
    }

    function nextStep() {
        if (currentStep < 4) currentStep++;
    }

    function prevStep() {
        if (currentStep > 1) currentStep--;
    }

    function canProceed(): boolean {
        switch (currentStep) {
            case 1:
                return selectedPool.length > 0;
            case 2:
                return datasetName.trim().length > 0;
            case 3:
                // Volume type requires size (quota) to be set
                if (formData.type === "volume" && quotaGB <= 0) {
                    return false;
                }
                return true;
            case 4:
                return true;
            default:
                return false;
        }
    }

    async function handleCreate() {
        if (creating) return;

        try {
            creating = true;

            // Build full dataset name
            formData.name = `${selectedPool}/${datasetName}`;

            // Convert quota GB to bytes
            if (quotaGB > 0) {
                formData.quota = quotaGB * 1024 * 1024 * 1024;
                // For volumes, size is required - use quota as size
                if (formData.type === "volume") {
                    formData.size = formData.quota;
                }
            } else {
                formData.quota = 0;
            }

            await api.createDataset(formData);

            onRefresh();
            onClose();
        } catch (err) {
            console.error("Failed to create dataset:", err);
            alert("创建失败: " + err);
        } finally {
            creating = false;
        }
    }

    function getUseCaseInfo() {
        return (
            useCases.find((uc) => uc.value === formData.use_case) || useCases[0]
        );
    }
</script>

<div class="h-full flex flex-col">
    <!-- Header -->
    <div
        class="flex items-center justify-between p-6 border-b border-border/50"
    >
        <div>
            <h2 class="text-xl font-bold text-foreground">创建存储空间</h2>
            <p class="text-sm text-muted-foreground mt-1">
                步骤 {currentStep} / 4
            </p>
        </div>
    </div>

    <!-- Progress Steps -->
    <div class="px-6 py-4 border-b border-border/50">
        <div class="flex items-center justify-between max-w-2xl mx-auto">
            {#each ["存储池", "名称与用途", "容量配置", "确认"] as step, i}
                <div class="flex items-center {i < 3 ? 'flex-1' : ''}">
                    <div
                        class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-semibold transition-all {i +
                            1 ===
                        currentStep
                            ? 'bg-primary text-primary-foreground'
                            : i + 1 < currentStep
                              ? 'bg-green-500 text-white'
                              : 'bg-muted text-muted-foreground'}"
                    >
                        {i + 1 < currentStep ? "✓" : i + 1}
                    </div>
                    <span
                        class="ml-2 text-sm {i + 1 === currentStep
                            ? 'text-foreground font-medium'
                            : 'text-muted-foreground'}"
                    >
                        {step}
                    </span>
                    {#if i < 3}
                        <div
                            class="flex-1 h-0.5 mx-4 {i + 1 < currentStep
                                ? 'bg-green-500'
                                : 'bg-muted'}"
                        ></div>
                    {/if}
                </div>
            {/each}
        </div>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6">
        {#if loading}
            <div class="flex items-center justify-center h-64">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else}
            <div class="max-w-2xl mx-auto">
                {#if currentStep === 1}
                    <!-- Step 1: Select Pool -->
                    <div class="space-y-4">
                        <div>
                            <label
                                for="datasetPool"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                选择存储池 *
                            </label>
                            <p class="text-sm text-muted-foreground mb-4">
                                选择一个存储池来创建此存储空间
                            </p>
                        </div>

                        {#if pools.length === 0}
                            <EmptyState
                                icon={Database}
                                title="暂无存储池"
                                description="需要先创建一个存储池，才能创建存储空间"
                                actionLabel="创建存储池"
                                onAction={handleCreatePool}
                            />
                        {:else}
                            <div class="grid grid-cols-1 gap-3">
                                {#each pools as pool}
                                    <button
                                        onclick={() =>
                                            (selectedPool = pool.name)}
                                        class="p-4 rounded-lg border-2 transition-all text-left {selectedPool ===
                                        pool.name
                                            ? 'border-primary bg-primary/10'
                                            : 'border-border hover:border-border/80'}"
                                    >
                                        <div
                                            class="flex items-center justify-between"
                                        >
                                            <div>
                                                <div
                                                    class="font-semibold text-foreground"
                                                >
                                                    {pool.name}
                                                </div>
                                                <div
                                                    class="text-sm text-muted-foreground mt-1"
                                                >
                                                    可用空间：{(
                                                        (pool.free /
                                                            pool.size) *
                                                        100
                                                    ).toFixed(1)}% ({(
                                                        pool.free /
                                                        1024 /
                                                        1024 /
                                                        1024
                                                    ).toFixed(1)} GB)
                                                </div>
                                            </div>
                                            {#if selectedPool === pool.name}
                                                <Check
                                                    class="w-5 h-5 text-primary"
                                                />
                                            {/if}
                                        </div>
                                    </button>
                                {/each}
                            </div>
                        {/if}
                    </div>
                {:else if currentStep === 2}
                    <!-- Step 2: Name & Use Case -->
                    <div class="space-y-6">
                        <div>
                            <label
                                for="datasetName"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                存储空间名称 *
                            </label>
                            <input
                                id="datasetName"
                                bind:value={datasetName}
                                placeholder="例如：data、media、backups"
                                class="w-full px-4 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <p class="text-xs text-muted-foreground mt-1">
                                将创建为：{selectedPool}/{datasetName || "..."}
                            </p>
                        </div>

                        <div>
                            <label
                                for="datasetType"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                类型
                            </label>
                            <div class="grid grid-cols-2 gap-3">
                                <button
                                    onclick={() =>
                                        (formData.type = "filesystem")}
                                    class="p-4 rounded-lg border-2 transition-all text-left {formData.type ===
                                    'filesystem'
                                        ? 'border-primary bg-primary/10'
                                        : 'border-border hover:border-border/80'}"
                                >
                                    <div class="font-semibold text-foreground">
                                        文件系统
                                    </div>
                                    <div
                                        class="text-sm text-muted-foreground mt-1"
                                    >
                                        用于 SMB/NFS 共享
                                    </div>
                                </button>
                                <button
                                    onclick={() => (formData.type = "volume")}
                                    class="p-4 rounded-lg border-2 transition-all text-left {formData.type ===
                                    'volume'
                                        ? 'border-primary bg-primary/10'
                                        : 'border-border hover:border-border/80'}"
                                >
                                    <div class="font-semibold text-foreground">
                                        块设备
                                    </div>
                                    <div
                                        class="text-sm text-muted-foreground mt-1"
                                    >
                                        用于 iSCSI/VM 磁盘
                                    </div>
                                </button>
                            </div>
                        </div>

                        <div>
                            <label
                                for="useCase"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                使用场景
                            </label>
                            <div class="grid grid-cols-1 gap-2">
                                {#each useCases as useCase}
                                    <button
                                        onclick={() =>
                                            (formData.use_case = useCase.value)}
                                        class="p-3 rounded-lg border transition-all text-left {formData.use_case ===
                                        useCase.value
                                            ? 'border-primary bg-primary/10'
                                            : 'border-border hover:border-border/80'}"
                                    >
                                        <div
                                            class="flex items-center justify-between"
                                        >
                                            <div>
                                                <div
                                                    class="font-medium text-foreground"
                                                >
                                                    {useCase.label}
                                                </div>
                                                <div
                                                    class="text-sm text-muted-foreground"
                                                >
                                                    {useCase.desc}
                                                </div>
                                            </div>
                                            {#if formData.use_case === useCase.value}
                                                <Check
                                                    class="w-5 h-5 text-primary"
                                                />
                                            {/if}
                                        </div>
                                    </button>
                                {/each}
                            </div>
                        </div>
                    </div>
                {:else if currentStep === 3}
                    <!-- Step 3: Capacity & Quota -->
                    <div class="space-y-6">
                        <div>
                            <label
                                for="quotaMode"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                配额模式
                            </label>
                            <div class="grid grid-cols-2 gap-3">
                                <button
                                    onclick={() =>
                                        (formData.quota_mode = "flexible")}
                                    class="p-4 rounded-lg border-2 transition-all text-left {formData.quota_mode ===
                                    'flexible'
                                        ? 'border-primary bg-primary/10'
                                        : 'border-border hover:border-border/80'}"
                                >
                                    <div class="font-semibold text-foreground">
                                        弹性容量
                                    </div>
                                    <div
                                        class="text-sm text-muted-foreground mt-1"
                                    >
                                        仅设置上限，提高空间利用率
                                    </div>
                                </button>
                                <button
                                    onclick={() =>
                                        (formData.quota_mode = "fixed")}
                                    class="p-4 rounded-lg border-2 transition-all text-left {formData.quota_mode ===
                                    'fixed'
                                        ? 'border-primary bg-primary/10'
                                        : 'border-border hover:border-border/80'}"
                                >
                                    <div class="font-semibold text-foreground">
                                        固定容量
                                    </div>
                                    <div
                                        class="text-sm text-muted-foreground mt-1"
                                    >
                                        预留空间，保证可用
                                    </div>
                                </button>
                            </div>
                        </div>

                        <div>
                            <label
                                for="datasetQuota"
                                class="block text-sm font-medium text-foreground mb-2"
                            >
                                {formData.type === "volume"
                                    ? "容量大小（GB）*"
                                    : "配额大小（GB）"}
                            </label>
                            <input
                                bind:value={quotaGB}
                                type="number"
                                min="0"
                                placeholder={formData.type === "volume"
                                    ? "必填"
                                    : "0 = 无限制"}
                                class="w-full px-4 py-2 bg-background border border-border rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                            />
                            <p class="text-xs text-muted-foreground mt-1">
                                {#if formData.type === "volume"}
                                    块设备必须指定固定大小
                                {:else if quotaGB > 0}
                                    设置为 {quotaGB} GB
                                {:else}
                                    不设置配额限制
                                {/if}
                            </p>
                        </div>
                    </div>
                {:else if currentStep === 4}
                    <!-- Step 4: Confirmation -->
                    <div class="space-y-6">
                        <div class="glass-card rounded-lg p-6">
                            <h3 class="font-semibold text-foreground mb-4">
                                确认信息
                            </h3>
                            <div class="space-y-3 text-sm">
                                <div class="flex justify-between">
                                    <span class="text-muted-foreground"
                                        >名称：</span
                                    >
                                    <span class="text-foreground font-medium"
                                        >{selectedPool}/{datasetName}</span
                                    >
                                </div>
                                <div class="flex justify-between">
                                    <span class="text-muted-foreground"
                                        >类型：</span
                                    >
                                    <span class="text-foreground font-medium">
                                        {formData.type === "filesystem"
                                            ? "文件系统"
                                            : "块设备"}
                                    </span>
                                </div>
                                <div class="flex justify-between">
                                    <span class="text-muted-foreground"
                                        >使用场景：</span
                                    >
                                    <span class="text-foreground font-medium"
                                        >{getUseCaseInfo().label}</span
                                    >
                                </div>
                                <div class="flex justify-between">
                                    <span class="text-muted-foreground"
                                        >配额模式：</span
                                    >
                                    <span class="text-foreground font-medium">
                                        {formData.quota_mode === "flexible"
                                            ? "弹性容量"
                                            : "固定容量"}
                                    </span>
                                </div>
                                <div class="flex justify-between">
                                    <span class="text-muted-foreground"
                                        >配额：</span
                                    >
                                    <span class="text-foreground font-medium">
                                        {quotaGB > 0
                                            ? `${quotaGB} GB`
                                            : "无限制"}
                                    </span>
                                </div>
                            </div>
                        </div>

                        <div
                            class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg"
                        >
                            <p class="text-sm text-blue-900 dark:text-blue-100">
                                <strong>提示：</strong
                                >根据使用场景"{getUseCaseInfo()
                                    .label}"，将自动应用优化的 ZFS 属性配置。
                            </p>
                        </div>
                    </div>
                {/if}
            </div>
        {/if}
    </div>

    <!-- Footer -->
    <div class="p-6 border-t border-border/50 flex justify-between">
        <button
            onclick={prevStep}
            disabled={currentStep === 1}
            class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
            <ChevronLeft class="w-4 h-4" />
            上一步
        </button>

        {#if currentStep < 4}
            <button
                onclick={nextStep}
                disabled={!canProceed()}
                class="flex items-center gap-2 px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
            >
                下一步
                <ChevronRight class="w-4 h-4" />
            </button>
        {:else}
            <button
                onclick={handleCreate}
                disabled={creating}
                class="flex items-center gap-2 px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {creating ? "创建中..." : "创建存储空间"}
            </button>
        {/if}
    </div>
</div>
