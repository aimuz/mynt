<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { api, type DiskDetail, type Disk } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        HardDrive,
        MapPin,
        Wrench,
        CircleCheckBig,
        TriangleAlert,
        ArrowRight,
        ArrowLeft,
        Loader,
        RefreshCw,
        Lightbulb,
        CircleX,
    } from "@lucide/svelte";

    interface Props {
        poolName: string;
        faultedDisk: DiskDetail;
        onComplete?: () => void;
        onClose?: () => void;
    }

    let { poolName, faultedDisk, onComplete, onClose }: Props = $props();

    let currentStep = $state(1);
    let locateLedOn = $state(false);
    let availableDisks = $state<Disk[]>([]);
    let selectedNewDisk = $state<string | null>(null);
    let loading = $state(false);
    let error = $state<string | null>(null);

    onMount(() => {
        loadAvailableDisks();
    });

    async function loadAvailableDisks() {
        try {
            const disks = await api.listDisks();
            // Filter to only unused disks
            availableDisks = (disks || []).filter(
                (d) => !d.in_use && d.size > 0 && d.name !== faultedDisk.name,
            );
        } catch (err) {
            console.error("Failed to load disks:", err);
        }
    }

    async function toggleLocateLed() {
        try {
            if (locateLedOn) {
                await api.locateDisk(faultedDisk.name, "off");
            } else {
                await api.locateDisk(faultedDisk.name, "on");
            }
            locateLedOn = !locateLedOn;
        } catch (err) {
            error = err instanceof Error ? err.message : "无法切换定位LED";
            console.error("Failed to toggle LED:", err);
        }
    }

    async function startRebuild() {
        if (!selectedNewDisk) return;

        loading = true;
        error = null;
        try {
            await api.replaceDisk(poolName, faultedDisk.name, selectedNewDisk);
            currentStep = 4;
            onComplete?.();
        } catch (err) {
            error = err instanceof Error ? err.message : "开始重建失败";
        } finally {
            loading = false;
        }
    }

    function nextStep() {
        if (currentStep < 4) {
            currentStep++;
        }
    }

    function prevStep() {
        if (currentStep > 1) {
            currentStep--;
        }
    }

    // Clean up LED when component unmounts
    onDestroy(async () => {
        if (locateLedOn) {
            try {
                await api.locateDisk(faultedDisk.name, "off");
            } catch {
                // Ignore errors
            }
        }
    });
</script>

<div class="p-6 h-full flex flex-col">
    <!-- Progress Steps -->
    <div class="flex items-center justify-center mb-8">
        {#each [1, 2, 3, 4] as step}
            <div class="flex items-center">
                <div
                    class="w-10 h-10 rounded-full flex items-center justify-center font-medium transition-colors {currentStep >=
                    step
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-muted-foreground'}"
                >
                    {#if currentStep > step}
                        <CircleCheckBig class="w-5 h-5" />
                    {:else}
                        {step}
                    {/if}
                </div>
                {#if step < 4}
                    <div
                        class="w-16 h-1 mx-2 rounded transition-colors {currentStep >
                        step
                            ? 'bg-primary'
                            : 'bg-muted'}"
                    ></div>
                {/if}
            </div>
        {/each}
    </div>

    <!-- Step Labels -->
    <div class="flex justify-between text-xs text-muted-foreground mb-6 px-2">
        <span class={currentStep >= 1 ? "text-foreground" : ""}>定位磁盘</span>
        <span class={currentStep >= 2 ? "text-foreground" : ""}>物理更换</span>
        <span class={currentStep >= 3 ? "text-foreground" : ""}>开始重建</span>
        <span class={currentStep >= 4 ? "text-foreground" : ""}>完成</span>
    </div>

    <!-- Step Content -->
    <div class="flex-1 overflow-auto">
        {#if currentStep === 1}
            <!-- Step 1: Locate Disk -->
            <div class="space-y-6">
                <div class="text-center">
                    <MapPin class="w-16 h-16 mx-auto mb-4 text-primary" />
                    <h2 class="text-xl font-bold text-foreground mb-2">
                        📍 定位故障磁盘
                    </h2>
                    <p class="text-muted-foreground">
                        点击下方按钮让故障磁盘的 LED 闪烁
                    </p>
                </div>

                <div class="glass-card rounded-xl p-6">
                    <h3 class="font-semibold text-foreground mb-4">磁盘信息</h3>
                    <div class="grid grid-cols-2 gap-4">
                        <div>
                            <span class="text-muted-foreground text-sm"
                                >磁盘名称</span
                            >
                            <p class="font-medium text-foreground">
                                {faultedDisk.name}
                            </p>
                        </div>
                        <div>
                            <span class="text-muted-foreground text-sm"
                                >槽位</span
                            >
                            <p class="font-medium text-foreground">
                                {faultedDisk.slot || "未知"}
                            </p>
                        </div>
                        <div>
                            <span class="text-muted-foreground text-sm"
                                >路径</span
                            >
                            <p class="font-medium text-foreground">
                                {faultedDisk.path}
                            </p>
                        </div>
                        <div>
                            <span class="text-muted-foreground text-sm"
                                >状态</span
                            >
                            <p class="font-medium text-red-500">
                                {faultedDisk.status}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="flex justify-center">
                    <button
                        onclick={toggleLocateLed}
                        class="px-6 py-3 rounded-xl font-medium flex items-center gap-2 transition-all {locateLedOn
                            ? 'bg-yellow-500 text-black animate-pulse'
                            : 'bg-primary text-primary-foreground'}"
                    >
                        <Lightbulb class="w-5 h-5" />
                        {locateLedOn
                            ? "LED 正在闪烁（点击关闭）"
                            : "点亮定位 LED"}
                    </button>
                </div>

                {#if locateLedOn}
                    <div
                        class="text-center text-sm text-yellow-500 animate-pulse"
                    >
                        ⚡ LED 定位灯正在闪烁，请找到机箱中闪烁的硬盘
                    </div>
                {/if}
            </div>
        {:else if currentStep === 2}
            <!-- Step 2: Physical Replacement -->
            <div class="space-y-6">
                <div class="text-center">
                    <Wrench class="w-16 h-16 mx-auto mb-4 text-primary" />
                    <h2 class="text-xl font-bold text-foreground mb-2">
                        🔧 物理更换硬盘
                    </h2>
                    <p class="text-muted-foreground">请按以下步骤操作</p>
                </div>

                <div class="glass-card rounded-xl p-6 space-y-4">
                    <div class="flex items-start gap-3">
                        <div
                            class="w-8 h-8 rounded-full bg-green-500/20 flex items-center justify-center shrink-0"
                        >
                            <CircleCheckBig class="w-5 h-5 text-green-500" />
                        </div>
                        <div>
                            <p class="font-medium text-foreground">
                                1. 确认已定位到故障磁盘
                            </p>
                            <p class="text-sm text-muted-foreground">
                                您已在上一步定位了磁盘 {faultedDisk.name}
                            </p>
                        </div>
                    </div>

                    <div class="flex items-start gap-3">
                        <div
                            class="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0"
                        >
                            <span class="text-sm font-medium">2</span>
                        </div>
                        <div>
                            <p class="font-medium text-foreground">
                                安全拔出故障磁盘
                            </p>
                            <p class="text-sm text-muted-foreground">
                                如果支持热插拔，可直接拔出；否则请先关机
                            </p>
                        </div>
                    </div>

                    <div class="flex items-start gap-3">
                        <div
                            class="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0"
                        >
                            <span class="text-sm font-medium">3</span>
                        </div>
                        <div>
                            <p class="font-medium text-foreground">
                                将新硬盘插入同一槽位
                            </p>
                            <p class="text-sm text-muted-foreground">
                                确保新硬盘容量不小于故障磁盘
                            </p>
                        </div>
                    </div>

                    <div class="flex items-start gap-3">
                        <div
                            class="w-8 h-8 rounded-full bg-muted flex items-center justify-center shrink-0"
                        >
                            <span class="text-sm font-medium">4</span>
                        </div>
                        <div>
                            <p class="font-medium text-foreground">
                                确认新硬盘已被系统识别
                            </p>
                            <p class="text-sm text-muted-foreground">
                                点击下一步查看可用磁盘列表
                            </p>
                        </div>
                    </div>
                </div>

                <div
                    class="p-4 rounded-xl bg-yellow-500/10 border border-yellow-500/30"
                >
                    <div class="flex items-start gap-2">
                        <TriangleAlert
                            class="w-5 h-5 text-yellow-500 shrink-0"
                        />
                        <p class="text-sm text-yellow-200">
                            <strong>注意：</strong>更换期间请勿关闭 NAS
                            电源。如不支持热插拔，请在关机状态下更换。
                        </p>
                    </div>
                </div>
            </div>
        {:else if currentStep === 3}
            <!-- Step 3: Select New Disk and Start Rebuild -->
            <div class="space-y-6">
                <div class="text-center">
                    <RefreshCw class="w-16 h-16 mx-auto mb-4 text-primary" />
                    <h2 class="text-xl font-bold text-foreground mb-2">
                        🔄 开始重建
                    </h2>
                    <p class="text-muted-foreground">
                        选择新磁盘并开始数据重建
                    </p>
                </div>

                <div class="glass-card rounded-xl p-6">
                    <div class="flex items-center justify-between mb-4">
                        <h3 class="font-semibold text-foreground">
                            选择新磁盘
                        </h3>
                        <button
                            onclick={loadAvailableDisks}
                            class="text-sm text-primary hover:underline flex items-center gap-1"
                        >
                            <RefreshCw class="w-4 h-4" />
                            刷新列表
                        </button>
                    </div>

                    {#if availableDisks.length === 0}
                        <div class="text-center py-8 text-muted-foreground">
                            <HardDrive
                                class="w-12 h-12 mx-auto mb-2 opacity-50"
                            />
                            <p>未检测到可用的新磁盘</p>
                            <p class="text-sm mt-1">
                                请确保新磁盘已插入并被系统识别
                            </p>
                        </div>
                    {:else}
                        <div class="space-y-2">
                            {#each availableDisks as disk}
                                <button
                                    onclick={() =>
                                        (selectedNewDisk = disk.name)}
                                    class="w-full p-4 rounded-lg border-2 transition-all text-left {selectedNewDisk ===
                                    disk.name
                                        ? 'border-primary bg-primary/10'
                                        : 'border-border hover:border-primary/50'}"
                                >
                                    <div class="flex items-center gap-3">
                                        <div
                                            class="w-10 h-10 rounded-lg bg-muted flex items-center justify-center"
                                        >
                                            <HardDrive
                                                class="w-5 h-5 text-muted-foreground"
                                            />
                                        </div>
                                        <div class="flex-1">
                                            <p
                                                class="font-medium text-foreground"
                                            >
                                                {disk.name}
                                            </p>
                                            <p
                                                class="text-sm text-muted-foreground"
                                            >
                                                {disk.model || "未知型号"} • {formatBytes(
                                                    disk.size,
                                                )}
                                            </p>
                                        </div>
                                        {#if selectedNewDisk === disk.name}
                                            <CircleCheckBig
                                                class="w-6 h-6 text-primary"
                                            />
                                        {/if}
                                    </div>
                                </button>
                            {/each}
                        </div>
                    {/if}
                </div>

                {#if error}
                    <div
                        class="p-4 rounded-xl bg-red-500/10 border border-red-500/30"
                    >
                        <div class="flex items-start gap-2">
                            <CircleX class="w-5 h-5 text-red-500 shrink-0" />
                            <p class="text-sm text-red-400">{error}</p>
                        </div>
                    </div>
                {/if}

                <div class="glass-card rounded-xl p-4">
                    <h4 class="font-medium text-foreground mb-2">
                        即将执行的操作：
                    </h4>
                    <div class="text-sm text-muted-foreground space-y-1">
                        <p>
                            • 旧磁盘：<span class="text-red-400"
                                >{faultedDisk.name}</span
                            >（故障）
                        </p>
                        <p>
                            • 新磁盘：<span class="text-green-400"
                                >{selectedNewDisk || "未选择"}</span
                            >
                        </p>
                        <p>• 存储池：{poolName}</p>
                    </div>
                </div>

                <div
                    class="p-4 rounded-xl bg-blue-500/10 border border-blue-500/30"
                >
                    <div class="flex items-start gap-2">
                        <Lightbulb class="w-5 h-5 text-blue-400 shrink-0" />
                        <div class="text-sm text-blue-300">
                            <p><strong>重建过程中请注意：</strong></p>
                            <ul class="list-disc list-inside mt-1 space-y-1">
                                <li>请勿关闭 NAS 电源</li>
                                <li>重建期间性能会有所下降</li>
                                <li>
                                    预计重建时间：约 4-8 小时（取决于数据量）
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        {:else if currentStep === 4}
            <!-- Step 4: Complete -->
            <div class="space-y-6">
                <div class="text-center">
                    <CircleCheckBig
                        class="w-20 h-20 mx-auto mb-4 text-green-500"
                    />
                    <h2 class="text-xl font-bold text-foreground mb-2">
                        🎉 重建已开始
                    </h2>
                    <p class="text-muted-foreground">
                        数据正在后台重建中，您可以关闭此窗口
                    </p>
                </div>

                <div class="glass-card rounded-xl p-6 text-center">
                    <p class="text-sm text-muted-foreground mb-4">
                        您可以在存储池详情页查看重建进度
                    </p>
                    <button
                        onclick={onClose}
                        class="px-6 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90"
                    >
                        关闭窗口
                    </button>
                </div>
            </div>
        {/if}
    </div>

    <!-- Navigation Buttons -->
    {#if currentStep < 4}
        <div class="flex justify-between mt-6 pt-4 border-t border-border">
            <button
                onclick={prevStep}
                disabled={currentStep === 1}
                class="px-4 py-2 rounded-lg border border-border hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
                <ArrowLeft class="w-4 h-4" />
                上一步
            </button>

            {#if currentStep === 3}
                <button
                    onclick={startRebuild}
                    disabled={!selectedNewDisk || loading}
                    class="px-6 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
                >
                    {#if loading}
                        <Loader class="w-4 h-4 animate-spin" />
                        正在启动...
                    {:else}
                        开始重建 🚀
                    {/if}
                </button>
            {:else}
                <button
                    onclick={nextStep}
                    class="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:opacity-90 flex items-center gap-2"
                >
                    下一步
                    <ArrowRight class="w-4 h-4" />
                </button>
            {/if}
        </div>
    {/if}
</div>
