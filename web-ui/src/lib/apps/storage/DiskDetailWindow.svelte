<script lang="ts">
    import { onMount } from "svelte";
    import {
        api,
        type Disk,
        type DetailedSmartReport,
        type SmartTestStatus,
    } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        HardDrive,
        Thermometer,
        Clock,
        Power,
        TriangleAlert,
        CircleCheck,
        CircleX,
        RefreshCw,
        Flashlight,
        Play,
        LoaderCircle,
        Database,
        Info,
    } from "@lucide/svelte";

    interface Props {
        disk: Disk;
    }

    let { disk }: Props = $props();

    let smartReport = $state<DetailedSmartReport | null>(null);
    let testStatus = $state<SmartTestStatus | null>(null);
    let loading = $state(true);
    let refreshing = $state(false);
    let testRunning = $state(false);
    let locateActive = $state(false);
    let error = $state<string | null>(null);

    onMount(() => {
        loadSmartData();
        // Poll for test status if running
        const interval = setInterval(() => {
            if (testRunning) {
                checkTestStatus();
            }
        }, 5000);
        return () => clearInterval(interval);
    });

    async function loadSmartData() {
        loading = true;
        error = null;
        try {
            smartReport = await api.getDiskSmartDetails(disk.name);
            testStatus = await api.getSmartTestStatus(disk.name);
            testRunning = testStatus?.running || false;
        } catch (err) {
            console.error("Failed to load SMART data:", err);
            error = err instanceof Error ? err.message : "加载 SMART 数据失败";
        } finally {
            loading = false;
        }
    }

    async function refreshSmartNow() {
        refreshing = true;
        error = null;
        try {
            // Force fresh SMART data fetch (bypasses cache)
            smartReport = await api.refreshSmartData(disk.name);
            testStatus = await api.getSmartTestStatus(disk.name);
            testRunning = testStatus?.running || false;
        } catch (err) {
            console.error("Failed to refresh SMART data:", err);
            error = err instanceof Error ? err.message : "刷新 SMART 数据失败";
        } finally {
            refreshing = false;
        }
    }

    async function runTest(type: "short" | "long") {
        try {
            testRunning = true;
            await api.runSmartTest(disk.name, type);
            // Start polling for status
            checkTestStatus();
        } catch (err) {
            console.error("Failed to start SMART test:", err);
            testRunning = false;
        }
    }

    async function checkTestStatus() {
        try {
            testStatus = await api.getSmartTestStatus(disk.name);
            testRunning = testStatus?.running || false;
        } catch (err) {
            console.error("Failed to get test status:", err);
        }
    }

    async function toggleLocate() {
        try {
            const action = locateActive ? "off" : "on";
            await api.locateDisk(disk.name, action);
            locateActive = !locateActive;
        } catch (err) {
            console.error("Failed to toggle locate LED:", err);
        }
    }

    function getAttributeStatus(attr: { status: string; id: number }): string {
        if (attr.status === "FAILING") return "text-red-500";
        // Highlight critical attributes
        const criticalIds = [5, 10, 187, 188, 196, 197, 198, 201];
        if (criticalIds.includes(attr.id))
            return "text-yellow-600 dark:text-yellow-400";
        return "text-foreground";
    }

    function isCriticalAttribute(id: number): boolean {
        return [5, 10, 187, 188, 196, 197, 198, 201].includes(id);
    }

    function formatPowerOnTime(hours: number): string {
        if (hours < 24) return `${hours} 小时`;
        const days = Math.floor(hours / 24);
        if (days < 365) return `${days} 天`;
        const years = Math.floor(days / 365);
        const remainDays = days % 365;
        return `${years} 年 ${remainDays} 天`;
    }
</script>

<div class="p-6 h-full overflow-auto">
    {#if loading}
        <div class="flex items-center justify-center h-64">
            <div
                class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else if error}
        <div class="glass-card rounded-xl p-8 text-center">
            <CircleX class="w-12 h-12 mx-auto mb-4 text-red-500" />
            <h3 class="text-lg font-semibold text-foreground mb-2">加载失败</h3>
            <p class="text-sm text-muted-foreground mb-4">{error}</p>
            <button
                onclick={() => loadSmartData()}
                class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90"
            >
                重试
            </button>
        </div>
    {:else}
        <!-- Basic Info Section -->
        <div class="glass-card rounded-xl p-6 mb-6">
            <div class="flex items-start gap-4">
                <div
                    class="w-16 h-16 rounded-xl bg-linear-to-br from-slate-500 to-slate-700 flex items-center justify-center shrink-0"
                >
                    <HardDrive class="w-8 h-8 text-white" />
                </div>
                <div class="flex-1 min-w-0">
                    <h2 class="text-xl font-bold text-foreground">
                        {disk.model || disk.name}
                    </h2>
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-3">
                        <div>
                            <p class="text-xs text-muted-foreground">设备名</p>
                            <p class="text-sm font-medium text-foreground">
                                {disk.name}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">序列号</p>
                            <p class="text-sm font-medium text-foreground">
                                {disk.serial}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">容量</p>
                            <p class="text-sm font-medium text-foreground">
                                {formatBytes(disk.size)}
                            </p>
                        </div>
                        <div>
                            <p class="text-xs text-muted-foreground">类型</p>
                            <p class="text-sm font-medium text-foreground">
                                {disk.type}
                            </p>
                        </div>
                    </div>
                    {#if disk.pool}
                        <div class="mt-3 flex items-center gap-2">
                            <Database class="w-4 h-4 text-primary" />
                            <span class="text-sm text-foreground"
                                >所属存储池: <span class="font-medium"
                                    >{disk.pool}</span
                                ></span
                            >
                        </div>
                    {/if}
                </div>
            </div>
        </div>

        <!-- Health Status Cards -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <!-- SMART Status -->
            <div class="glass-card rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                    {#if smartReport?.passed}
                        <CircleCheck class="w-5 h-5 text-green-500" />
                    {:else}
                        <CircleX class="w-5 h-5 text-red-500" />
                    {/if}
                    <span class="text-sm font-medium text-foreground"
                        >S.M.A.R.T. 状态</span
                    >
                </div>
                <p
                    class="text-lg font-bold {smartReport?.passed
                        ? 'text-green-500'
                        : 'text-red-500'}"
                >
                    {smartReport?.passed ? "通过" : "失败"}
                </p>
            </div>

            <!-- Temperature -->
            <div class="glass-card rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                    <Thermometer class="w-5 h-5 text-muted-foreground" />
                    <span class="text-sm font-medium text-foreground">温度</span
                    >
                </div>
                <p class="text-lg font-bold text-foreground">
                    {smartReport?.temperature ?? "-"}°C
                </p>
            </div>

            <!-- Power On Hours -->
            <div class="glass-card rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                    <Clock class="w-5 h-5 text-muted-foreground" />
                    <span class="text-sm font-medium text-foreground"
                        >运行时间</span
                    >
                </div>
                <p class="text-lg font-bold text-foreground">
                    {smartReport?.power_on_hours
                        ? formatPowerOnTime(smartReport.power_on_hours)
                        : "-"}
                </p>
            </div>

            <!-- Power Cycle Count -->
            <div class="glass-card rounded-xl p-4">
                <div class="flex items-center gap-2 mb-2">
                    <Power class="w-5 h-5 text-muted-foreground" />
                    <span class="text-sm font-medium text-foreground"
                        >开关机次数</span
                    >
                </div>
                <p class="text-lg font-bold text-foreground">
                    {smartReport?.power_cycle_count ?? "-"}
                </p>
            </div>
        </div>

        <!-- Error Counters -->
        <div class="glass-card rounded-xl p-6 mb-6">
            <h3
                class="text-lg font-semibold text-foreground mb-4 flex items-center gap-2"
            >
                <TriangleAlert class="w-5 h-5" />
                错误计数
            </h3>
            <div class="grid grid-cols-3 gap-6">
                <div>
                    <p class="text-sm text-muted-foreground">重新分配扇区</p>
                    <p
                        class="text-2xl font-bold {smartReport?.reallocated_sectors
                            ? 'text-yellow-500'
                            : 'text-foreground'}"
                    >
                        {smartReport?.reallocated_sectors ?? 0}
                    </p>
                </div>
                <div>
                    <p class="text-sm text-muted-foreground">待处理扇区</p>
                    <p
                        class="text-2xl font-bold {smartReport?.pending_sectors
                            ? 'text-yellow-500'
                            : 'text-foreground'}"
                    >
                        {smartReport?.pending_sectors ?? 0}
                    </p>
                </div>
                <div>
                    <p class="text-sm text-muted-foreground">不可修正错误</p>
                    <p
                        class="text-2xl font-bold {smartReport?.uncorrectable_errors
                            ? 'text-red-500'
                            : 'text-foreground'}"
                    >
                        {smartReport?.uncorrectable_errors ?? 0}
                    </p>
                </div>
            </div>
        </div>

        <!-- Actions -->
        <div class="glass-card rounded-xl p-6 mb-6">
            <h3 class="text-lg font-semibold text-foreground mb-4">操作</h3>
            <div class="flex flex-wrap gap-3">
                <button
                    onclick={() => runTest("short")}
                    disabled={testRunning}
                    class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {#if testRunning}
                        <LoaderCircle class="w-4 h-4 animate-spin" />
                    {:else}
                        <Play class="w-4 h-4" />
                    {/if}
                    S.M.A.R.T. 短测
                </button>
                <button
                    onclick={() => runTest("long")}
                    disabled={testRunning}
                    class="flex items-center gap-2 px-4 py-2 border border-border rounded-lg hover:bg-white/5 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    {#if testRunning}
                        <LoaderCircle class="w-4 h-4 animate-spin" />
                    {:else}
                        <Play class="w-4 h-4" />
                    {/if}
                    S.M.A.R.T. 长测
                </button>
                <button
                    onclick={toggleLocate}
                    class="flex items-center gap-2 px-4 py-2 border border-border rounded-lg hover:bg-white/5 transition-all {locateActive
                        ? 'bg-yellow-500/20 border-yellow-500'
                        : ''}"
                >
                    <Flashlight
                        class="w-4 h-4 {locateActive ? 'text-yellow-500' : ''}"
                    />
                    {locateActive ? "停止定位" : "定位硬盘"}
                </button>
                <button
                    onclick={refreshSmartNow}
                    disabled={refreshing}
                    class="flex items-center gap-2 px-4 py-2 border border-border rounded-lg hover:bg-white/5 transition-all disabled:opacity-50"
                    title="强制从硬盘读取最新 S.M.A.R.T. 数据"
                >
                    <RefreshCw
                        class="w-4 h-4 {refreshing ? 'animate-spin' : ''}"
                    />
                    {refreshing ? "刷新中..." : "刷新 S.M.A.R.T."}
                </button>
            </div>

            {#if testRunning && testStatus}
                <div class="mt-4 p-4 bg-primary/10 rounded-lg">
                    <div class="flex items-center gap-2 mb-2">
                        <LoaderCircle
                            class="w-4 h-4 animate-spin text-primary"
                        />
                        <span class="text-sm font-medium text-foreground"
                            >测试进行中</span
                        >
                    </div>
                    {#if testStatus.progress !== undefined}
                        <div class="w-full bg-muted rounded-full h-2 mb-2">
                            <div
                                class="bg-primary h-2 rounded-full transition-all"
                                style="width: {testStatus.progress}%"
                            ></div>
                        </div>
                        <p class="text-xs text-muted-foreground">
                            进度: {testStatus.progress}%
                        </p>
                    {/if}
                </div>
            {:else if testStatus?.last_result}
                <div
                    class="mt-4 p-4 bg-muted/50 rounded-lg flex items-center gap-2"
                >
                    <Info class="w-4 h-4 text-muted-foreground" />
                    <span class="text-sm text-muted-foreground"
                        >上次测试结果: {testStatus.last_result}</span
                    >
                </div>
            {/if}
        </div>

        <!-- S.M.A.R.T. Attributes Table -->
        <div class="glass-card rounded-xl overflow-hidden">
            <div class="p-4 border-b border-border">
                <h3 class="text-lg font-semibold text-foreground">
                    S.M.A.R.T. 属性
                </h3>
                <p class="text-xs text-muted-foreground mt-1">
                    显示所有 S.M.A.R.T. 属性，关键属性以黄色高亮显示
                </p>
            </div>
            <div class="overflow-x-auto">
                <table class="w-full">
                    <thead class="bg-muted/50 border-b border-border">
                        <tr>
                            <th
                                class="text-left px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >ID</th
                            >
                            <th
                                class="text-left px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >属性名称</th
                            >
                            <th
                                class="text-right px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >当前值</th
                            >
                            <th
                                class="text-right px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >最差值</th
                            >
                            <th
                                class="text-right px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >阈值</th
                            >
                            <th
                                class="text-left px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >原始值</th
                            >
                            <th
                                class="text-left px-4 py-2 text-xs font-semibold text-muted-foreground"
                                >状态</th
                            >
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-border">
                        {#each smartReport?.attributes ?? [] as attr}
                            <tr
                                class="hover:bg-white/5 {isCriticalAttribute(
                                    attr.id,
                                )
                                    ? 'bg-yellow-500/5'
                                    : ''}"
                            >
                                <td
                                    class="px-4 py-2 text-sm {getAttributeStatus(
                                        attr,
                                    )}">{attr.id}</td
                                >
                                <td
                                    class="px-4 py-2 text-sm {getAttributeStatus(
                                        attr,
                                    )} font-mono">{attr.name}</td
                                >
                                <td
                                    class="px-4 py-2 text-sm text-right {getAttributeStatus(
                                        attr,
                                    )}">{attr.value}</td
                                >
                                <td
                                    class="px-4 py-2 text-sm text-right text-muted-foreground"
                                    >{attr.worst}</td
                                >
                                <td
                                    class="px-4 py-2 text-sm text-right text-muted-foreground"
                                    >{attr.thresh}</td
                                >
                                <td
                                    class="px-4 py-2 text-sm text-foreground font-mono"
                                    >{attr.raw}</td
                                >
                                <td class="px-4 py-2">
                                    {#if attr.status === "OK"}
                                        <span
                                            class="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400"
                                            >OK</span
                                        >
                                    {:else}
                                        <span
                                            class="text-xs px-2 py-0.5 rounded-full bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400"
                                            >{attr.status}</span
                                        >
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>
