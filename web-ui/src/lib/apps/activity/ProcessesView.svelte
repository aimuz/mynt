<script lang="ts">
    import { formatBytes } from "$lib/utils";
    import type { SysProcess } from "$lib/api";
    import { Search, CircleX, Activity } from "@lucide/svelte";

    interface Props {
        processes: SysProcess[];
        onKill: (pid: number, signal: "TERM" | "KILL") => void;
    }

    let { processes, onKill }: Props = $props();

    let searchQuery = $state("");
    let debouncedQuery = $state("");
    let timer: any;

    // Debounce search input
    function handleSearchInput(e: Event) {
        const val = (e.target as HTMLInputElement).value;
        searchQuery = val;
        clearTimeout(timer);
        timer = setTimeout(() => {
            debouncedQuery = val;
        }, 300);
    }
    let sortBy = $state<keyof SysProcess>("cpu_percent");
    let sortDesc = $state(true);
    let confirmKill = $state<number | null>(null);

    let filteredProcesses = $derived.by(() => {
        const query = debouncedQuery.trim();
        let result: SysProcess[];

        if (query) {
            // Use RegExp for faster case-insensitive matching without repeated .toLowerCase()
            const re = new RegExp(
                query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
                "i",
            );
            result = processes.filter(
                (p) => re.test(p.name) || re.test(p.command) || re.test(p.user),
            );
        } else {
            result = [...processes];
        }

        // Sort
        result.sort((a, b) => {
            const aVal = a[sortBy];
            const bVal = b[sortBy];

            if (typeof aVal === "number" && typeof bVal === "number") {
                return sortDesc ? bVal - aVal : aVal - bVal;
            }

            // Fast string comparison
            const aStr = aVal as string;
            const bStr = bVal as string;
            if (aStr === bStr) return 0;
            const cmp = aStr > bStr ? 1 : -1;
            return sortDesc ? -cmp : cmp;
        });

        return result;
    });

    function handleSort(field: keyof SysProcess) {
        if (sortBy === field) {
            sortDesc = !sortDesc;
        } else {
            sortBy = field;
            sortDesc = true;
        }
    }

    function handleKillClick(pid: number) {
        if (confirmKill === pid) {
            onKill(pid, "TERM");
            confirmKill = null;
        } else {
            confirmKill = pid;
            // Reset after 3 seconds
            setTimeout(() => {
                if (confirmKill === pid) confirmKill = null;
            }, 3000);
        }
    }

    function getStateLabel(state: string): string {
        switch (state) {
            case "R":
                return "运行中";
            case "S":
                return "休眠";
            case "D":
                return "等待IO";
            case "Z":
                return "僵尸";
            case "T":
                return "已停止";
            default:
                return state;
        }
    }
</script>

<div class="p-6 overflow-auto flex-1">
    <div class="flex items-center justify-between mb-6">
        <div>
            <h2 class="text-2xl font-bold text-foreground">进程</h2>
            <p class="text-sm text-muted-foreground mt-1">
                {processes.length} 个进程
            </p>
        </div>

        <!-- Search -->
        <div class="relative">
            <Search
                class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground"
            />
            <input
                type="text"
                placeholder="搜索进程..."
                value={searchQuery}
                oninput={handleSearchInput}
                class="pl-10 pr-4 py-2 rounded-lg bg-white/5 border border-border/50 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
        </div>
    </div>

    {#if filteredProcesses.length === 0}
        <div class="flex items-center justify-center h-64">
            <div class="text-center text-muted-foreground">
                <Activity class="w-16 h-16 mx-auto mb-4 opacity-50" />
                <p class="text-lg">未找到进程</p>
            </div>
        </div>
    {:else}
        <div class="glass-card rounded-xl overflow-hidden">
            <div class="overflow-x-auto">
                <table class="w-full table-fixed">
                    <thead>
                        <tr class="border-b border-border/50">
                            <th
                                class="text-left p-3 text-xs text-muted-foreground font-medium w-24"
                            >
                                <button
                                    onclick={() => handleSort("pid")}
                                    class="hover:text-foreground"
                                >
                                    PID {sortBy === "pid"
                                        ? sortDesc
                                            ? "↓"
                                            : "↑"
                                        : ""}
                                </button>
                            </th>
                            <th
                                class="text-left p-3 text-xs text-muted-foreground font-medium w-auto"
                            >
                                <button
                                    onclick={() => handleSort("name")}
                                    class="hover:text-foreground"
                                >
                                    名称 {sortBy === "name"
                                        ? sortDesc
                                            ? "↓"
                                            : "↑"
                                        : ""}
                                </button>
                            </th>
                            <th
                                class="text-left p-3 text-xs text-muted-foreground font-medium w-32"
                            >
                                <button
                                    onclick={() => handleSort("user")}
                                    class="hover:text-foreground"
                                >
                                    用户 {sortBy === "user"
                                        ? sortDesc
                                            ? "↓"
                                            : "↑"
                                        : ""}
                                </button>
                            </th>
                            <th
                                class="text-right p-3 text-xs text-muted-foreground font-medium w-24"
                            >
                                <button
                                    onclick={() => handleSort("cpu_percent")}
                                    class="hover:text-foreground"
                                >
                                    CPU {sortBy === "cpu_percent"
                                        ? sortDesc
                                            ? "↓"
                                            : "↑"
                                        : ""}
                                </button>
                            </th>
                            <th
                                class="text-right p-3 text-xs text-muted-foreground font-medium w-24"
                            >
                                <button
                                    onclick={() => handleSort("mem_rss")}
                                    class="hover:text-foreground"
                                >
                                    内存 {sortBy === "mem_rss"
                                        ? sortDesc
                                            ? "↓"
                                            : "↑"
                                        : ""}
                                </button>
                            </th>
                            <th
                                class="text-center p-3 text-xs text-muted-foreground font-medium w-24"
                                >状态</th
                            >
                            <th
                                class="text-center p-3 text-xs text-muted-foreground font-medium w-16"
                                >操作</th
                            >
                        </tr>
                    </thead>
                    <tbody>
                        {#each filteredProcesses as proc (proc.pid)}
                            <tr
                                class="border-b border-border/30 hover:bg-white/5 transition-colors"
                            >
                                <td class="p-3 text-sm text-muted-foreground"
                                    >{proc.pid}</td
                                >
                                <td class="p-3">
                                    <p
                                        class="text-sm font-medium text-foreground"
                                    >
                                        {proc.name}
                                    </p>
                                    <p
                                        class="text-xs text-muted-foreground truncate max-w-xs"
                                        title={proc.command}
                                    >
                                        {proc.command}
                                    </p>
                                </td>
                                <td class="p-3 text-sm text-muted-foreground"
                                    >{proc.user}</td
                                >
                                <td class="p-3 text-sm text-right">
                                    <span
                                        class={proc.cpu_percent > 80
                                            ? "text-red-400"
                                            : proc.cpu_percent > 50
                                              ? "text-orange-400"
                                              : "text-foreground"}
                                    >
                                        {proc.cpu_percent.toFixed(1)}%
                                    </span>
                                </td>
                                <td
                                    class="p-3 text-sm text-right text-foreground"
                                >
                                    {formatBytes(proc.mem_rss)}
                                </td>
                                <td class="p-3 text-center">
                                    <span
                                        class="px-2 py-0.5 text-xs rounded-full bg-white/10 text-muted-foreground"
                                    >
                                        {getStateLabel(proc.state)}
                                    </span>
                                </td>
                                <td class="p-3 text-center">
                                    <button
                                        onclick={() =>
                                            handleKillClick(proc.pid)}
                                        class="p-1.5 rounded-lg transition-colors {confirmKill ===
                                        proc.pid
                                            ? 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
                                            : 'hover:bg-white/10 text-muted-foreground hover:text-foreground'}"
                                        title={confirmKill === proc.pid
                                            ? "再次点击确认终止"
                                            : "终止进程"}
                                    >
                                        <CircleX class="w-4 h-4" />
                                    </button>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </div>
    {/if}
</div>
