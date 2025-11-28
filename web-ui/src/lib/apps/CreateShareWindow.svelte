<script lang="ts">
    import { api } from "$lib/api";
    import { Plus } from "@lucide/svelte";

    interface Props {
        onRefresh?: () => void;
        onClose?: () => void;
    }

    let { onRefresh, onClose }: Props = $props();

    let name = $state("");
    let path = $state("");
    let protocol = $state<"smb" | "nfs">("smb");
    let shareType = $state<"normal" | "public" | "restricted">("normal");
    let readOnly = $state(false);
    let browseable = $state(true);
    let guestOk = $state(false);
    let validUsers = $state("");
    let comment = $state("");

    let submitting = $state(false);
    let error = $state("");

    // Validation
    let nameError = $derived(
        name && !/^[a-zA-Z0-9_-]+$/.test(name)
            ? "Share name can only contain letters, numbers, underscores and hyphens"
            : "",
    );

    let pathError = $derived(
        path && !path.startsWith("/") ? "Path must start with /" : "",
    );

    let isValid = $derived(name && path && !nameError && !pathError);

    async function handleSubmit() {
        if (!isValid || submitting) return;

        submitting = true;
        error = "";

        try {
            await api.createShare({
                name,
                path,
                protocol,
                share_type: shareType,
                read_only: readOnly,
                browseable,
                guest_ok: guestOk,
                valid_users: validUsers,
                comment,
            });

            onRefresh?.();
            onClose?.();
        } catch (err) {
            error = String(err);
        } finally {
            submitting = false;
        }
    }
</script>

<div class="p-6 max-w-xl mx-auto">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">Create New Share</h2>
        <p class="text-sm text-muted-foreground mt-1">
            Add a new SMB or NFS network share
        </p>
    </div>

    {#if error}
        <div
            class="mb-4 p-4 bg-red-500/10 border border-red-500/30 rounded-lg text-red-500 text-sm"
        >
            {error}
        </div>
    {/if}

    <form
        onsubmit={(e) => {
            e.preventDefault();
            handleSubmit();
        }}
        class="space-y-4"
    >
        <!-- Share Name -->
        <div>
            <label
                for="name"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Share Name <span class="text-red-500">*</span>
            </label>
            <input
                id="name"
                type="text"
                bind:value={name}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="myshare"
                required
            />
            {#if nameError}
                <p class="text-xs text-red-500 mt-1">{nameError}</p>
            {/if}
        </div>

        <!-- Path -->
        <div>
            <label
                for="path"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Path <span class="text-red-500">*</span>
            </label>
            <input
                id="path"
                type="text"
                bind:value={path}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="/mnt/pool/myshare"
                required
            />
            {#if pathError}
                <p class="text-xs text-red-500 mt-1">{pathError}</p>
            {/if}
        </div>

        <!-- Protocol -->
        <div>
            <label class="block text-sm font-medium text-foreground mb-2">
                Protocol
            </label>
            <div class="flex gap-3">
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={protocol}
                        value="smb"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">SMB</span>
                    <span class="text-xs text-muted-foreground"
                        >(Windows/Mac/Linux)</span
                    >
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={protocol}
                        value="nfs"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">NFS</span>
                    <span class="text-xs text-muted-foreground"
                        >(Unix/Linux)</span
                    >
                </label>
            </div>
        </div>

        <!-- Share Type -->
        <div>
            <label class="block text-sm font-medium text-foreground mb-2">
                Share Type
            </label>
            <div class="flex gap-3">
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={shareType}
                        value="normal"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">Normal</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={shareType}
                        value="public"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">Public</span>
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={shareType}
                        value="restricted"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">Restricted</span>
                </label>
            </div>
        </div>

        <!-- Permissions -->
        <div class="space-y-2">
            <label class="block text-sm font-medium text-foreground mb-2">
                Permissions
            </label>
            <div class="flex items-center gap-2">
                <input
                    id="readOnly"
                    type="checkbox"
                    bind:checked={readOnly}
                    class="rounded text-primary focus:ring-primary"
                />
                <label
                    for="readOnly"
                    class="text-sm text-foreground cursor-pointer"
                >
                    Read-only access
                </label>
            </div>
            <div class="flex items-center gap-2">
                <input
                    id="browseable"
                    type="checkbox"
                    bind:checked={browseable}
                    class="rounded text-primary focus:ring-primary"
                />
                <label
                    for="browseable"
                    class="text-sm text-foreground cursor-pointer"
                >
                    Browseable in network browser
                </label>
            </div>
            <div class="flex items-center gap-2">
                <input
                    id="guestOk"
                    type="checkbox"
                    bind:checked={guestOk}
                    class="rounded text-primary focus:ring-primary"
                />
                <label
                    for="guestOk"
                    class="text-sm text-foreground cursor-pointer"
                >
                    Allow guest access
                </label>
            </div>
        </div>

        <!-- Valid Users -->
        <div>
            <label
                for="validUsers"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Valid Users
            </label>
            <input
                id="validUsers"
                type="text"
                bind:value={validUsers}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="user1, user2 (leave empty for all users)"
            />
            <p class="text-xs text-muted-foreground mt-1">
                Comma-separated list of usernames who can access this share
            </p>
        </div>

        <!-- Comment -->
        <div>
            <label
                for="comment"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Comment
            </label>
            <textarea
                id="comment"
                bind:value={comment}
                rows="2"
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground resize-none"
                placeholder="Optional description for this share"
            ></textarea>
        </div>

        <!-- Actions -->
        <div
            class="flex items-center justify-end gap-3 pt-4 border-t border-border/50"
        >
            <button
                type="button"
                onclick={onClose}
                disabled={submitting}
                class="px-4 py-2 rounded-lg dark:hover:bg-gray-800/50 hover:bg-gray-200/50 transition-colors dark:text-white text-gray-900 disabled:opacity-50"
            >
                Cancel
            </button>
            <button
                type="submit"
                disabled={!isValid || submitting}
                class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-all shadow-lg disabled:opacity-50 flex items-center gap-2"
            >
                {#if submitting}
                    <div
                        class="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"
                    ></div>
                    Creating...
                {:else}
                    <Plus class="w-4 h-4" />
                    Create Share
                {/if}
            </button>
        </div>
    </form>
</div>
