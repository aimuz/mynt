<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type Disk } from "$lib/api";
    import { formatBytes } from "$lib/utils";
    import {
        HardDrive,
        CircleAlert,
        Plus,
        TriangleAlert,
    } from "@lucide/svelte";

    interface CreatePoolWindowProps {
        onRefreshStorage?: () => void;
    }

    let { onRefreshStorage }: CreatePoolWindowProps = $props();

    // Get desktop context
    const desktop = getContext<{
        closeWindow: (id: string) => void;
    }>("desktop");

    let disks = $state<Disk[]>([]);
    let loading = $state(true);
    let creating = $state(false);

    // Form state
    let poolName = $state("");
    let selectedDisks = $state<string[]>([]);
    let raidType = $state("");
    let error = $state("");
    let success = $state("");

    const raidTypes = [
        {
            value: "",
            label: "Stripe",
            description: "No redundancy, maximum capacity",
            minDisks: 1,
        },
        {
            value: "mirror",
            label: "Mirror",
            description: "Full redundancy, 50% capacity",
            minDisks: 2,
        },
        {
            value: "raidz",
            label: "RAIDZ",
            description: "Single parity, good balance",
            minDisks: 3,
        },
        {
            value: "raidz2",
            label: "RAIDZ2",
            description: "Double parity, high redundancy",
            minDisks: 4,
        },
    ];

    onMount(() => {
        loadDisks();
    });

    async function loadDisks() {
        try {
            disks = (await api.listDisks().catch(() => [])) || [];
            loading = false;
        } catch (err) {
            console.error("Failed to load disks:", err);
            loading = false;
        }
    }

    function toggleDisk(diskPath: string) {
        if (selectedDisks.includes(diskPath)) {
            selectedDisks = selectedDisks.filter((d) => d !== diskPath);
        } else {
            selectedDisks = [...selectedDisks, diskPath];
        }
    }

    function validateForm(): string | null {
        if (!poolName.trim()) {
            return "Pool name is required";
        }

        if (!/^[a-zA-Z0-9_-]+$/.test(poolName)) {
            return "Pool name can only contain letters, numbers, hyphens, and underscores";
        }

        if (selectedDisks.length === 0) {
            return "At least one disk must be selected";
        }

        const selectedRaid = raidTypes.find((r) => r.value === raidType);
        if (selectedRaid && selectedDisks.length < selectedRaid.minDisks) {
            return `${selectedRaid.label} requires at least ${selectedRaid.minDisks} disks`;
        }

        return null;
    }

    async function handleCreatePool() {
        error = "";
        success = "";

        const validationError = validateForm();
        if (validationError) {
            error = validationError;
            return;
        }

        creating = true;

        try {
            await api.createPool(poolName, selectedDisks, raidType);
            success = `Pool "${poolName}" created successfully!`;

            // Refresh storage app if callback provided
            if (onRefreshStorage) {
                onRefreshStorage();
            }

            // Close window after a brief delay to show success message
            setTimeout(() => {
                desktop.closeWindow("create-pool");
            }, 1500);
        } catch (err: any) {
            error = err.message || "Failed to create pool";
            creating = false;
        }
    }

    function isDiskAvailable(disk: Disk): boolean {
        // Show all disks, but warn about in-use ones
        return true;
    }

    function canSelectDisk(disk: Disk): boolean {
        // Allow selecting in-use disks but will show warnings
        return !disk.in_use;
    }

    // Format disk usage for display (ready for future i18n)
    function formatDiskUsage(disk: Disk): string {
        if (!disk.usage) return "";

        switch (disk.usage.type) {
            case "zfs_member":
                return disk.usage.params?.pool
                    ? `ZFS Pool Member (${disk.usage.params.pool})`
                    : "ZFS Pool Member";
            case "formatted":
                return disk.usage.params?.fstype
                    ? `Formatted (${disk.usage.params.fstype})`
                    : "Formatted";
            case "has_partitions":
                return "Has Partitions";
            case "system_disk":
                return disk.usage.params?.fstype
                    ? `System Disk (${disk.usage.params.fstype})`
                    : "System Disk";
            default:
                return "In Use";
        }
    }
</script>

<div class="p-6 h-full overflow-auto">
    {#if loading}
        <div class="flex items-center justify-center h-64">
            <div
                class="animate-spin rounded-full h-12 w-12 border-4 border-primary border-t-transparent"
            ></div>
        </div>
    {:else}
        <!-- Error/Success Messages -->
        {#if error}
            <div
                class="mb-4 p-3 rounded-lg bg-red-100 dark:bg-red-900/30 border border-red-200 dark:border-red-800 flex items-start gap-2"
            >
                <CircleAlert
                    class="w-4 h-4 text-red-600 dark:text-red-400 mt-0.5"
                />
                <p class="text-sm text-red-800 dark:text-red-400">
                    {error}
                </p>
            </div>
        {/if}

        {#if success}
            <div
                class="mb-4 p-3 rounded-lg bg-green-100 dark:bg-green-900/30 border border-green-200 dark:border-green-800"
            >
                <p class="text-sm text-green-800 dark:text-green-400">
                    {success}
                </p>
            </div>
        {/if}

        <!-- Pool Name -->
        <div class="mb-6">
            <label
                for="poolName"
                class="block text-sm font-semibold dark:text-white text-gray-900 mb-2"
            >
                Pool Name
            </label>
            <input
                id="poolName"
                type="text"
                bind:value={poolName}
                placeholder="e.g., tank, storage, backup"
                class="w-full px-4 py-2 rounded-lg dark:bg-gray-800/50 bg-gray-100/50 border dark:border-gray-700 border-gray-300 dark:text-white text-gray-900 dark:placeholder-gray-500 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 transition-all"
            />
            <p class="text-xs dark:text-gray-400 text-gray-600 mt-1">
                Use only letters, numbers, hyphens, and underscores
            </p>
        </div>

        <!-- RAID Type -->
        <fieldset class="mb-6">
            <legend
                class="block text-sm font-semibold dark:text-white text-gray-900 mb-3"
            >
                RAID Type
            </legend>
            <div class="grid grid-cols-2 gap-3">
                {#each raidTypes as raid}
                    <button
                        onclick={() => (raidType = raid.value)}
                        class="p-4 rounded-lg border-2 transition-all text-left {raidType ===
                        raid.value
                            ? 'dark:border-blue-500 border-blue-600 dark:bg-blue-500/20 bg-blue-100/80'
                            : 'dark:border-gray-700 border-gray-300 dark:hover:border-gray-600 hover:border-gray-400 dark:hover:bg-gray-800/50 hover:bg-gray-100/50'}"
                    >
                        <div
                            class="font-semibold text-sm dark:text-white text-gray-900 mb-1"
                        >
                            {raid.label}
                        </div>
                        <div
                            class="text-xs dark:text-gray-400 text-gray-600 mb-2"
                        >
                            {raid.description}
                        </div>
                        <div class="text-xs dark:text-gray-500 text-gray-500">
                            Min: {raid.minDisks} disk{raid.minDisks > 1
                                ? "s"
                                : ""}
                        </div>
                    </button>
                {/each}
            </div>
        </fieldset>

        <!-- Disk Selection -->
        <fieldset class="mb-6">
            <legend
                class="block text-sm font-semibold dark:text-white text-gray-900 mb-3"
            >
                Select Disks ({selectedDisks.length} selected)
            </legend>
            {#if disks.length === 0}
                <div
                    class="p-8 rounded-lg border-2 border-dashed dark:border-gray-700 border-gray-300 text-center"
                >
                    <HardDrive
                        class="w-12 h-12 mx-auto mb-2 opacity-50 dark:text-gray-500 text-gray-400"
                    />
                    <p class="text-sm dark:text-gray-400 text-gray-600">
                        No available disks found
                    </p>
                </div>
            {:else}
                <div class="space-y-2 max-h-64 overflow-y-auto pr-2">
                    {#each disks as disk}
                        {#if isDiskAvailable(disk)}
                            <button
                                onclick={() =>
                                    canSelectDisk(disk) &&
                                    toggleDisk(disk.path)}
                                disabled={disk.in_use}
                                class="w-full p-3 rounded-lg border transition-all text-left flex items-center gap-3 {disk.in_use
                                    ? 'dark:border-orange-500/50 border-orange-400/50 dark:bg-orange-500/10 bg-orange-50/50 opacity-75 cursor-not-allowed'
                                    : selectedDisks.includes(disk.path)
                                      ? 'dark:border-blue-500 border-blue-600 dark:bg-blue-500/20 bg-blue-100/80'
                                      : 'dark:border-gray-700 border-gray-300 dark:hover:border-gray-600 hover:border-gray-400 dark:hover:bg-gray-800/50 hover:bg-gray-100/50'}"
                            >
                                <div
                                    class="w-10 h-10 rounded-lg {disk.in_use
                                        ? 'bg-linear-to-br from-orange-500 to-red-500'
                                        : 'bg-linear-to-br from-blue-500 to-cyan-500'} flex items-center justify-center shrink-0"
                                >
                                    {#if disk.in_use}
                                        <TriangleAlert
                                            class="w-5 h-5 text-white"
                                        />
                                    {:else}
                                        <HardDrive class="w-5 h-5 text-white" />
                                    {/if}
                                </div>
                                <div class="flex-1 min-w-0">
                                    <div
                                        class="font-semibold text-sm dark:text-white text-gray-900 flex items-center gap-2"
                                    >
                                        {disk.name}
                                        {#if disk.in_use}
                                            <span
                                                class="text-[10px] px-1.5 py-0.5 rounded bg-orange-500/20 text-orange-600 dark:text-orange-400 uppercase tracking-wide font-bold"
                                            >
                                                In Use
                                            </span>
                                        {/if}
                                    </div>
                                    <div
                                        class="text-xs dark:text-gray-400 text-gray-600 truncate"
                                    >
                                        {disk.path}
                                        {#if disk.model}
                                            · {disk.model}
                                        {/if}
                                    </div>
                                    {#if disk.in_use && disk.usage}
                                        <div
                                            class="text-xs text-orange-600 dark:text-orange-400 mt-1 flex items-center gap-1"
                                        >
                                            <CircleAlert class="w-3 h-3" />
                                            {formatDiskUsage(disk)}
                                        </div>
                                    {/if}
                                </div>
                                <div class="text-right shrink-0">
                                    <div
                                        class="text-sm font-semibold dark:text-white text-gray-900"
                                    >
                                        {formatBytes(disk.size)}
                                    </div>
                                    <div
                                        class="text-xs dark:text-gray-400 text-gray-600"
                                    >
                                        {disk.type}
                                    </div>
                                </div>
                            </button>
                        {/if}
                    {/each}
                </div>
            {/if}
        </fieldset>

        <!-- Footer Buttons -->
        <div
            class="flex items-center justify-end gap-3 pt-4 border-t border-border/50"
        >
            <button
                onclick={() => desktop.closeWindow("create-pool")}
                disabled={creating}
                class="px-4 py-2 rounded-lg dark:hover:bg-gray-800/50 hover:bg-gray-200/50 transition-colors dark:text-white text-gray-900 disabled:opacity-50"
            >
                Cancel
            </button>
            <button
                onclick={handleCreatePool}
                disabled={creating}
                class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-all shadow-lg disabled:opacity-50 flex items-center gap-2"
            >
                {#if creating}
                    <div
                        class="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"
                    ></div>
                    Creating...
                {:else}
                    <Plus class="w-4 h-4" />
                    Create Pool
                {/if}
            </button>
        </div>
    {/if}
</div>
