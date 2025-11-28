<script lang="ts">
    import { onMount, getContext } from "svelte";
    import { api, type User } from "$lib/api";
    import {
        Users,
        Plus,
        Trash2,
        Shield,
        User as UserIcon,
    } from "@lucide/svelte";
    import CreateUserWindow from "$lib/apps/CreateUserWindow.svelte";

    let users = $state<User[]>([]);
    let loading = $state(true);
    let currentUser = $state<User | null>(null);

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
            users = await api.listUsers().catch(() => []);
            loading = false;
        } catch (err) {
            console.error("Failed to load users:", err);
            loading = false;
        }
    }

    function handleCreateUser() {
        desktop.openWindow("create-user", "Create User", UserIcon, () => ({
            component: CreateUserWindow,
            props: { onRefresh: loadData },
        }));
    }

    async function handleDeleteUser(username: string) {
        if (!confirm(`Are you sure you want to delete user "${username}"?`)) {
            return;
        }

        try {
            await api.deleteUser(username);
            await loadData();
        } catch (err) {
            alert(`Failed to delete user: ${err}`);
        }
    }

    function getAccountTypeBadge(accountType: string): string {
        return accountType === "system"
            ? "bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400"
            : "bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400";
    }

    function getInitials(user: User): string {
        if (user.full_name) {
            const parts = user.full_name.split(" ");
            if (parts.length >= 2) {
                return (parts[0][0] + parts[1][0]).toUpperCase();
            }
            return parts[0][0].toUpperCase();
        }
        return user.username[0].toUpperCase();
    }

    function formatDate(dateStr?: string): string {
        if (!dateStr) return "Never";
        const date = new Date(dateStr);
        return date.toLocaleDateString() + " " + date.toLocaleTimeString();
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
                    User Management
                </h2>
                <p class="text-sm text-muted-foreground mt-1">
                    Manage system and virtual users
                </p>
            </div>
            {#if currentUser?.is_admin}
                <button
                    onclick={handleCreateUser}
                    class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg hover:shadow-xl"
                >
                    <Plus class="w-4 h-4" />
                    Create User
                </button>
            {/if}
        </div>

        <!-- Users Grid -->
        {#if users.length === 0}
            <div class="glass-card rounded-xl p-12 text-center fade-in">
                <Users
                    class="w-16 h-16 mx-auto mb-4 opacity-50 text-muted-foreground"
                />
                <h3 class="text-lg font-semibold text-foreground mb-2">
                    No Users
                </h3>
                <p class="text-sm text-muted-foreground mb-6">
                    Create your first user to get started
                </p>
                {#if currentUser?.is_admin}
                    <button
                        onclick={handleCreateUser}
                        class="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all"
                    >
                        <Plus class="w-4 h-4" />
                        Create User
                    </button>
                {/if}
            </div>
        {:else}
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
                {#each users as user, i}
                    <div
                        class="glass-card rounded-xl p-6 fade-in hover:bg-white/5 transition-all"
                        style="animation-delay: {i * 50}ms;"
                    >
                        <div class="flex items-start justify-between mb-4">
                            <div class="flex items-center gap-3">
                                <!-- Avatar with initials -->
                                <div
                                    class="w-12 h-12 rounded-full bg-linear-to-br from-purple-500 to-blue-600 flex items-center justify-center shadow-lg text-white font-semibold"
                                >
                                    {getInitials(user)}
                                </div>
                                <div>
                                    <div class="flex items-center gap-2">
                                        <h3
                                            class="font-semibold text-lg text-foreground"
                                        >
                                            {user.username}
                                        </h3>
                                        {#if user.is_admin}
                                            <Shield
                                                class="w-4 h-4 text-amber-500"
                                            />
                                        {/if}
                                    </div>
                                    <p
                                        class="text-sm text-muted-foreground mt-0.5"
                                    >
                                        {user.full_name || "No name set"}
                                    </p>
                                </div>
                            </div>
                            {#if currentUser?.is_admin && user.username !== currentUser.username}
                                <button
                                    onclick={() =>
                                        handleDeleteUser(user.username)}
                                    class="p-2 text-red-500 hover:bg-red-500/10 rounded-lg transition-all"
                                    title="Delete user"
                                >
                                    <Trash2 class="w-4 h-4" />
                                </button>
                            {/if}
                        </div>

                        <!-- User Info -->
                        <div class="space-y-2 mb-3">
                            {#if user.email}
                                <div class="text-sm">
                                    <span class="text-muted-foreground"
                                        >Email:</span
                                    >
                                    <span class="text-foreground ml-2"
                                        >{user.email}</span
                                    >
                                </div>
                            {/if}
                            <div class="flex gap-2 flex-wrap">
                                <span
                                    class="text-xs px-2 py-0.5 rounded-full {getAccountTypeBadge(
                                        user.account_type,
                                    )}"
                                >
                                    {user.account_type}
                                </span>
                                {#if !user.is_active}
                                    <span
                                        class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400"
                                    >
                                        Inactive
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
                                    Created
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {new Date(
                                        user.created_at,
                                    ).toLocaleDateString()}
                                </p>
                            </div>
                            <div>
                                <p class="text-xs text-muted-foreground">
                                    Last Login
                                </p>
                                <p
                                    class="text-sm font-semibold text-foreground mt-0.5"
                                >
                                    {user.last_login
                                        ? new Date(
                                              user.last_login,
                                          ).toLocaleDateString()
                                        : "Never"}
                                </p>
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    {/if}
</div>
