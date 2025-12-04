<script lang="ts">
    import { onMount } from "svelte";
    import { api } from "$lib/api";
    import {
        Activity,
        RefreshCw,
        Clock,
        CircleCheckBig,
        CircleX,
        Loader,
    } from "@lucide/svelte";

    // Task interface - matches potential backend task info
    interface Task {
        id: string;
        type: string; // "scrub", "resilver", "backup", "snapshot"
        pool?: string;
        dataset?: string;
        status: string; // "running", "completed", "failed", "pending"
        progress?: number; // 0-100
        startTime: Date;
        endTime?: Date;
        message?: string;
    }

    let tasks = $state<Task[]>([]);
    let loading = $state(true);

    onMount(() => {
        loadData();
    });

    async function loadData() {
        loading = true;
        try {
            // TODO: API endpoint for task listing
            tasks = [];
        } catch (err) {
            console.error("Failed to load tasks:", err);
        } finally {
            loading = false;
        }
    }

    function getStatusInfo(status: string) {
        switch (status) {
            case "running":
                return {
                    icon: Loader,
                    color: "text-blue-500",
                    label: "进行中",
                };
            case "completed":
                return {
                    icon: CircleCheckBig,
                    color: "text-green-500",
                    label: "已完成",
                };
            case "failed":
                return { icon: CircleX, color: "text-red-500", label: "失败" };
            case "pending":
                return {
                    icon: Clock,
                    color: "text-yellow-500",
                    label: "等待中",
                };
            default:
                return {
                    icon: Activity,
                    color: "text-muted-foreground",
                    label: status,
                };
        }
    }

    function getTaskTypeLabel(type: string): string {
        switch (type) {
            case "scrub":
                return "数据校验";
            case "resilver":
                return "磁盘修复";
            case "backup":
                return "备份任务";
            case "snapshot":
                return "快照任务";
            default:
                return type;
        }
    }

    function formatDuration(start: Date, end?: Date): string {
        const endTime = end || new Date();
        const duration = endTime.getTime() - start.getTime();
        const hours = Math.floor(duration / 3600000);
        const minutes = Math.floor((duration % 3600000) / 60000);
        const seconds = Math.floor((duration % 60000) / 1000);

        if (hours > 0) {
            return `${hours}小时${minutes}分钟`;
        } else if (minutes > 0) {
            return `${minutes}分钟${seconds}秒`;
        }
        return `${seconds}秒`;
    }
</script>

<div class="p-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
        <div>
            <h2 class="text-2xl font-bold text-foreground">任务与日志</h2>
            <p class="text-sm text-muted-foreground mt-1">
                查看存储相关的后台任务和活动日志
            </p>
        </div>
        <button
            onclick={() => loadData()}
            class="flex items-center gap-2 px-4 py-2 rounded-lg border border-border hover:bg-white/5 transition-all"
        >
            <RefreshCw class="w-4 h-4" />
            刷新
        </button>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto">
        {#if loading}
            <div class="flex items-center justify-center h-64">
                <div
                    class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
                ></div>
            </div>
        {:else if tasks.length === 0}
            <!-- Placeholder -->
            <div class="glass-card rounded-xl p-12 text-center">
                <Activity
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    任务日志
                </h3>
                <p class="text-sm text-muted-foreground mb-4">
                    暂无进行中的任务
                </p>
                <div class="text-xs text-muted-foreground">
                    <p>将显示：</p>
                    <ul class="mt-2 space-y-1">
                        <li>• 数据校验（Scrub）任务进度</li>
                        <li>• 磁盘修复（Resilver）状态</li>
                        <li>• 快照策略执行记录</li>
                        <li>• 备份任务历史</li>
                    </ul>
                </div>
            </div>
        {:else}
            <!-- Task List -->
            <div class="space-y-3">
                {#each tasks as task, i}
                    {@const statusInfo = getStatusInfo(task.status)}
                    <div
                        class="glass-card rounded-lg p-4 fade-in"
                        style="animation-delay: {i * 30}ms;"
                    >
                        <div class="flex items-start justify-between">
                            <div class="flex-1">
                                <div class="flex items-center gap-2 mb-2">
                                    <statusInfo.icon
                                        class="w-5 h-5 {statusInfo.color}"
                                    />
                                    <h4 class="font-semibold text-foreground">
                                        {getTaskTypeLabel(task.type)}
                                    </h4>
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full bg-muted text-muted-foreground"
                                    >
                                        {statusInfo.label}
                                    </span>
                                </div>

                                <p class="text-sm text-muted-foreground mb-2">
                                    {#if task.pool}
                                        存储池：{task.pool}
                                    {/if}
                                    {#if task.dataset}
                                        <span class="mx-2">•</span>
                                        数据集：{task.dataset}
                                    {/if}
                                </p>

                                {#if task.progress !== undefined && task.status === "running"}
                                    <div class="mb-2">
                                        <div
                                            class="flex justify-between text-xs text-muted-foreground mb-1"
                                        >
                                            <span>进度</span>
                                            <span>{task.progress}%</span>
                                        </div>
                                        <div
                                            class="w-full bg-muted rounded-full h-2"
                                        >
                                            <div
                                                class="bg-primary h-2 rounded-full transition-all"
                                                style="width: {task.progress}%"
                                            ></div>
                                        </div>
                                    </div>
                                {/if}

                                {#if task.message}
                                    <p class="text-xs text-muted-foreground">
                                        {task.message}
                                    </p>
                                {/if}
                            </div>

                            <div
                                class="text-right text-sm text-muted-foreground"
                            >
                                <p>
                                    开始：{task.startTime.toLocaleString(
                                        "zh-CN",
                                    )}
                                </p>
                                <p>
                                    耗时：{formatDuration(
                                        task.startTime,
                                        task.endTime,
                                    )}
                                </p>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
