<script lang="ts">
    import {
        Database,
        House,
        HardDrive,
        Disc,
        Camera,
        Activity,
    } from "@lucide/svelte";
    import SnapshotView from "$lib/apps/storage/SnapshotView.svelte";
    import OverviewView from "$lib/apps/storage/OverviewView.svelte";
    import StorageSpaceListView from "$lib/apps/storage/StorageSpaceListView.svelte";
    import DiskListView from "$lib/apps/storage/DiskListView.svelte";
    import TaskLogView from "$lib/apps/storage/TaskLogView.svelte";
    import PoolListView from "$lib/apps/storage/PoolListView.svelte";

    // View state
    let currentView = $state<string>("overview");

    // Navigation menu items
    const navItems = [
        { id: "overview", name: "总览", icon: House },
        { id: "pools", name: "存储池", icon: Database },
        { id: "spaces", name: "存储空间", icon: Disc },
        { id: "disks", name: "磁盘", icon: HardDrive },
        { id: "snapshots", name: "快照", icon: Camera },
        { id: "tasks", name: "任务", icon: Activity },
    ];
</script>

<div class="flex h-full">
    <!-- Left Navigation Sidebar -->
    <nav class="w-48 glass-card border-r border-border/50 flex flex-col">
        <div class="flex-1 overflow-y-auto p-2">
            {#each navItems as item}
                <button
                    onclick={() => (currentView = item.id)}
                    class="w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm transition-all {currentView ===
                    item.id
                        ? 'bg-primary/10 text-primary font-medium'
                        : 'text-muted-foreground hover:bg-white/5 hover:text-foreground'}"
                >
                    <item.icon class="w-4 h-4" />
                    {item.name}
                </button>
            {/each}
        </div>
    </nav>

    <!-- Main Content Area -->
    <div class="flex-1 overflow-hidden flex flex-col">
        {#if currentView === "pools"}
            <PoolListView />
        {:else if currentView === "overview"}
            <!-- Overview View -->
            <OverviewView />
        {:else if currentView === "spaces"}
            <!-- Storage Spaces View -->
            <StorageSpaceListView />
        {:else if currentView === "disks"}
            <!-- Disks View -->
            <DiskListView />
        {:else if currentView === "snapshots"}
            <!-- Snapshots View -->
            <SnapshotView />
        {:else if currentView === "tasks"}
            <!-- Tasks View -->
            <TaskLogView />
        {/if}
    </div>
</div>
