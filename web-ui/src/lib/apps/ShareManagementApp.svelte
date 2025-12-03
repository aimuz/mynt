<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type Share } from "$lib/api";
    import {
        FolderOpen,
        Plus,
        Trash2,
        Network,
        Lock,
        Users,
    } from "@lucide/svelte";
    import CreateShareWindow from "$lib/apps/CreateShareWindow.svelte";

    let shares = $state<Share[]>([]);
    let loading = $state(true);
    let currentUser = $state<any>(null);

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

    onMount(() => {
        currentUser = api.getCurrentUser();
        loadData();
    });

    async function loadData() {
        try {
            shares = (await api.listShares().catch(() => [])) || [];
            loading = false;
        } catch (err) {
            console.error("Failed to load shares:", err);
            loading = false;
        }
    }

    function handleCreateShare() {
        desktop.openWindow("create-share", "Create Share", FolderOpen, () => ({
            component: CreateShareWindow,
            props: { onRefresh: loadData },
        }));
    }

    async function handleDeleteShare(id: number, name: string) {
        if (!confirm(`Are you sure you want to delete share "${name}"?`)) {
            return;
        }

        try {
            await api.deleteShare(id);
            await loadData();
        } catch (err) {
            alert(`Failed to delete share: ${err}`);
        }
    }

    function getProtocolBadge(protocol: string): string {
        switch (protocol.toLowerCase()) {
            case "smb":
                return "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400";
            case "nfs":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            default:
                return "bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400";
        }
    }

    function getShareTypeBadge(shareType: string): string {
        switch (shareType) {
            case "public":
                return "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400";
            case "restricted":
                return "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400";
            case "normal":
            default:
                return "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400";
        }
    }

    function getShareIcon(shareType: string) {
        switch (shareType) {
            case "public":
                return Users;
            case "restricted":
                return Lock;
            default:
                return FolderOpen;
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
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
            <div>
                <h2 class="text-2xl font-bold text-foreground">
                    Share Management
                </h2>
                <p class="text-sm text-muted-foreground mt-1">
                    Manage SMB and NFS network shares
                </p>
            </div>
            {#if currentUser?.is_admin}
                <button
                    onclick={handleCreateShare}
                    class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl"
                >
                    <Plus class="w-4 h-4" />
                    Create Share
                </button>
            {/if}
        </div>

        <!-- Shares Grid -->
        {#if shares.length === 0}
            <div class="glass-card rounded-xl p-12 text-center fade-in">
                <Network
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    No Shares
                </h3>
                <p class="text-sm text-muted-foreground mb-6">
                    Create your first network share to get started
                </p>
                {#if currentUser?.is_admin}
                    <button
                        onclick={handleCreateShare}
                        class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                    >
                        <Plus class="w-4 h-4" />
                        Create Share
                    </button>
                {/if}
            </div>
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {#each shares as share, i}
                    {@const SvelteComponent = getShareIcon(share.share_type)}
                    <div
                        class="glass-card rounded-xl p-6 fade-in hover:bg-white/5 transition-all"
                        style="animation-delay: {i * 50}ms;"
                    >
                        <div class="flex items-start justify-between mb-4">
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-12 h-12 rounded-xl bg-linear-to-br from-green-500 to-emerald-600 flex items-center justify-center shadow-lg"
                                >
                                    <SvelteComponent
                                        class="w-6 h-6 text-white"
                                    />
                                </div>
                                <div>
                                    <h3
                                        class="font-semibold text-lg text-foreground"
                                    >
                                        {share.name}
                                    </h3>
                                    <p
                                        class="text-sm text-muted-foreground mt-0.5"
                                    >
                                        {share.path}
                                    </p>
                                </div>
                            </div>
                            {#if currentUser?.is_admin}
                                <button
                                    onclick={() =>
                                        handleDeleteShare(share.id, share.name)}
                                    class="p-2 text-red-500 hover:bg-red-500/10 rounded-lg transition-all"
                                    title="Delete share"
                                >
                                    <Trash2 class="w-4 h-4" />
                                </button>
                            {/if}
                        </div>

                        <!-- Share Info -->
                        <div class="space-y-2 mb-3">
                            {#if share.comment}
                                <div class="text-sm text-muted-foreground">
                                    {share.comment}
                                </div>
                            {/if}
                            <div class="flex gap-2 flex-wrap">
                                <span
                                    class="text-xs px-2 py-0.5 rounded-full {getProtocolBadge(
                                        share.protocol,
                                    )}"
                                >
                                    {share.protocol.toUpperCase()}
                                </span>
                                <span
                                    class="text-xs px-2 py-0.5 rounded-full {getShareTypeBadge(
                                        share.share_type,
                                    )}"
                                >
                                    {share.share_type}
                                </span>
                                {#if share.read_only}
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400"
                                    >
                                        Read-only
                                    </span>
                                {/if}
                                {#if share.guest_ok}
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400"
                                    >
                                        Guest OK
                                    </span>
                                {/if}
                            </div>
                        </div>

                        <!-- Stats -->
                        <div
                            class="grid grid-cols-2 gap-4 mt-4 pt-4 border-t border-border/50"
                        >
                            <div>
                                <p class="text-xs text-muted-foreground">
                                    Browseable
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {share.browseable ? "Yes" : "No"}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-muted-foreground">
                                    Valid Users
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {share.valid_users || "All"}
                                </p>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    {/if}
</div>
