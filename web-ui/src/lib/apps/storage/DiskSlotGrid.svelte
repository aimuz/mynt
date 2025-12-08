<script lang="ts">
    import { type VDevDetail, type DiskDetail } from "$lib/api";
    import {
        HardDrive,
        CircleCheckBig,
        TriangleAlert,
        CircleX,
        MapPin,
        RefreshCw,
    } from "@lucide/svelte";

    interface Props {
        vdevs: VDevDetail[];
        onDiskClick?: (disk: DiskDetail) => void;
        onLocateDisk?: (disk: DiskDetail) => void;
    }

    let { vdevs, onDiskClick, onLocateDisk }: Props = $props();

    let locatingDisk = $state<string | null>(null);

    function getStatusColor(status: string): string {
        switch (status?.toUpperCase()) {
            case "ONLINE":
                return "border-green-500 bg-green-500/10";
            case "DEGRADED":
                return "border-yellow-500 bg-yellow-500/10";
            case "FAULTED":
                return "border-red-500 bg-red-500/10 animate-pulse";
            case "OFFLINE":
                return "border-gray-500 bg-gray-500/10";
            default:
                return "border-muted bg-muted/50";
        }
    }

    function getStatusIcon(status: string) {
        switch (status?.toUpperCase()) {
            case "ONLINE":
                return { icon: CircleCheckBig, class: "text-green-500" };
            case "DEGRADED":
                return { icon: TriangleAlert, class: "text-yellow-500" };
            case "FAULTED":
                return { icon: CircleX, class: "text-red-500" };
            default:
                return { icon: HardDrive, class: "text-muted-foreground" };
        }
    }

    function getVDevTypeLabel(type: string): string {
        switch (type) {
            case "mirror":
                return "镜像";
            case "raidz":
                return "RAID-Z1";
            case "raidz2":
                return "RAID-Z2";
            case "raidz3":
                return "RAID-Z3";
            case "stripe":
                return "条带";
            default:
                return type;
        }
    }

    async function handleLocate(disk: DiskDetail, event: Event) {
        event.stopPropagation();
        locatingDisk = disk.name;
        try {
            await onLocateDisk?.(disk);
            // Keep the indicator on for 3 seconds
            setTimeout(() => {
                if (locatingDisk === disk.name) {
                    locatingDisk = null;
                }
            }, 3000);
        } catch {
            locatingDisk = null;
        }
    }
</script>

<div class="space-y-6">
    {#each vdevs as vdev, vdevIndex}
        <div class="border border-border/50 rounded-lg p-4">
            <!-- VDev Header -->
            <div class="flex items-center gap-2 mb-4">
                <div
                    class="w-8 h-8 rounded-lg flex items-center justify-center {vdev.status ===
                    'ONLINE'
                        ? 'bg-green-500/20'
                        : vdev.status === 'DEGRADED'
                          ? 'bg-yellow-500/20'
                          : 'bg-red-500/20'}"
                >
                    <span class="text-sm font-bold text-foreground"
                        >{vdevIndex + 1}</span
                    >
                </div>
                <div>
                    <span class="font-medium text-foreground">{vdev.name}</span>
                    <span class="text-xs text-muted-foreground ml-2">
                        ({getVDevTypeLabel(vdev.type)})
                    </span>
                </div>
                <span
                    class="ml-auto px-2 py-0.5 rounded text-xs font-medium {vdev.status ===
                    'ONLINE'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                        : vdev.status === 'DEGRADED'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                          : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'}"
                >
                    {vdev.status}
                </span>
            </div>

            <!-- Disk Grid -->
            <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                {#each vdev.children as disk, diskIndex}
                    {@const statusInfo = getStatusIcon(disk.status)}
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <div
                        role="button"
                        tabindex="0"
                        onclick={() => onDiskClick?.(disk)}
                        onkeydown={(e) =>
                            e.key === "Enter" && onDiskClick?.(disk)}
                        class="relative p-3 rounded-lg border-2 transition-all hover:scale-[1.02] cursor-pointer {getStatusColor(
                            disk.status,
                        )} {disk.status !== 'ONLINE'
                            ? 'hover:shadow-lg hover:shadow-red-500/20'
                            : 'hover:shadow-lg hover:shadow-green-500/10'}"
                    >
                        <!-- Slot Number -->
                        <div
                            class="absolute -top-2 -left-2 w-6 h-6 rounded-full bg-background border border-border flex items-center justify-center text-xs font-bold text-foreground"
                        >
                            {disk.slot || diskIndex + 1}
                        </div>

                        <!-- Status Icon -->
                        <div class="flex justify-center mb-2">
                            <div
                                class="w-12 h-12 rounded-lg bg-card flex items-center justify-center"
                            >
                                {#if disk.replacing}
                                    <RefreshCw
                                        class="w-6 h-6 text-blue-500 animate-spin"
                                    />
                                {:else}
                                    {@const Component = statusInfo.icon}
                                    <Component
                                        class="w-6 h-6 {statusInfo.class}"
                                    />
                                {/if}
                            </div>
                        </div>

                        <!-- Disk Info -->
                        <div class="text-center">
                            <p
                                class="text-sm font-medium text-foreground truncate"
                            >
                                {disk.name}
                            </p>
                            <p class="text-xs text-muted-foreground">
                                {disk.status === "ONLINE"
                                    ? "正常"
                                    : disk.status === "DEGRADED"
                                      ? "降级"
                                      : disk.status === "FAULTED"
                                        ? "故障"
                                        : disk.status}
                            </p>
                            {#if disk.replacing}
                                <p class="text-xs text-blue-400 mt-1">
                                    正在替换...
                                </p>
                            {/if}
                        </div>

                        <!-- Error Counts -->
                        {#if disk.read > 0 || disk.write > 0 || disk.checksum > 0}
                            <div
                                class="mt-2 pt-2 border-t border-border/50 grid grid-cols-3 gap-1 text-xs"
                            >
                                <div
                                    class="text-center {disk.read > 0
                                        ? 'text-red-400'
                                        : 'text-muted-foreground'}"
                                >
                                    R:{disk.read}
                                </div>
                                <div
                                    class="text-center {disk.write > 0
                                        ? 'text-red-400'
                                        : 'text-muted-foreground'}"
                                >
                                    W:{disk.write}
                                </div>
                                <div
                                    class="text-center {disk.checksum > 0
                                        ? 'text-red-400'
                                        : 'text-muted-foreground'}"
                                >
                                    C:{disk.checksum}
                                </div>
                            </div>
                        {/if}

                        <!-- Locate Button -->
                        <button
                            onclick={(e) => handleLocate(disk, e)}
                            class="absolute -top-2 -right-2 w-6 h-6 rounded-full bg-primary text-primary-foreground flex items-center justify-center hover:scale-110 transition-transform"
                            title="定位磁盘 LED"
                        >
                            {#if locatingDisk === disk.name}
                                <div
                                    class="w-3 h-3 rounded-full bg-yellow-400 animate-ping"
                                ></div>
                            {:else}
                                <MapPin class="w-3 h-3" />
                            {/if}
                        </button>
                    </div>
                {/each}
            </div>
        </div>
    {/each}

    {#if vdevs.length === 0}
        <div class="text-center py-8 text-muted-foreground">
            <HardDrive class="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>暂无磁盘信息</p>
        </div>
    {/if}
</div>

<!-- Legend -->
<div class="mt-4 flex flex-wrap gap-4 text-xs text-muted-foreground">
    <div class="flex items-center gap-1">
        <div
            class="w-3 h-3 rounded border-2 border-green-500 bg-green-500/10"
        ></div>
        <span>正常</span>
    </div>
    <div class="flex items-center gap-1">
        <div
            class="w-3 h-3 rounded border-2 border-yellow-500 bg-yellow-500/10"
        ></div>
        <span>降级</span>
    </div>
    <div class="flex items-center gap-1">
        <div
            class="w-3 h-3 rounded border-2 border-red-500 bg-red-500/10"
        ></div>
        <span>故障</span>
    </div>
    <div class="flex items-center gap-1">
        <div
            class="w-3 h-3 rounded border-2 border-gray-500 bg-gray-500/10"
        ></div>
        <span>离线</span>
    </div>
</div>
