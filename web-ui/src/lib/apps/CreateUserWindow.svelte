<script lang="ts">
    import { api } from "$lib/api";
    import { Eye, EyeOff, Plus } from "@lucide/svelte";

    interface Props {
        onRefresh?: () => void;
        onClose?: () => void;
    }

    let { onRefresh, onClose }: Props = $props();

    let username = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let fullName = $state("");
    let email = $state("");
    let accountType = $state<"system" | "virtual">("virtual");
    let isAdmin = $state(false);
    let showPassword = $state(false);
    let showConfirmPassword = $state(false);

    let submitting = $state(false);
    let error = $state("");

    // Validation
    let usernameError = $derived(
        username && !/^[a-z_][a-z0-9_-]*$/.test(username)
            ? "Username must start with lowercase letter or underscore, contain only lowercase letters, numbers, underscores and hyphens"
            : "",
    );

    let emailError = $derived(
        email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
            ? "Invalid email format"
            : "",
    );

    let passwordError = $derived(
        password && password.length < 8
            ? "Password must be at least 8 characters"
            : "",
    );

    let confirmPasswordError = $derived(
        confirmPassword && confirmPassword !== password
            ? "Passwords do not match"
            : "",
    );

    let passwordStrength = $derived(() => {
        if (!password) return 0;
        let strength = 0;
        if (password.length >= 8) strength += 25;
        if (password.length >= 12) strength += 25;
        if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength += 25;
        if (/\d/.test(password)) strength += 15;
        if (/[^a-zA-Z\d]/.test(password)) strength += 10;
        return Math.min(strength, 100);
    });

    let isValid = $derived(
        username &&
            password &&
            confirmPassword &&
            !usernameError &&
            !emailError &&
            !passwordError &&
            !confirmPasswordError,
    );

    async function handleSubmit() {
        if (!isValid || submitting) return;

        submitting = true;
        error = "";

        try {
            await api.createUser({
                username,
                password,
                full_name: fullName || undefined,
                email: email || undefined,
                account_type: accountType,
                is_admin: isAdmin,
            });

            onRefresh?.();
            onClose?.();
        } catch (err) {
            error = String(err);
        } finally {
            submitting = false;
        }
    }

    function getStrengthColor(strength: number): string {
        if (strength < 40) return "bg-red-500";
        if (strength < 70) return "bg-yellow-500";
        return "bg-green-500";
    }

    function getStrengthLabel(strength: number): string {
        if (strength < 40) return "Weak";
        if (strength < 70) return "Medium";
        return "Strong";
    }
</script>

<div class="p-6 max-w-xl mx-auto">
    <div class="mb-6">
        <h2 class="text-2xl font-bold text-foreground">Create New User</h2>
        <p class="text-sm text-muted-foreground mt-1">
            Add a new system or virtual user account
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
        <!-- Username -->
        <div>
            <label
                for="username"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Username <span class="text-red-500">*</span>
            </label>
            <input
                id="username"
                type="text"
                bind:value={username}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="johndoe"
                required
            />
            {#if usernameError}
                <p class="text-xs text-red-500 mt-1">{usernameError}</p>
            {/if}
        </div>

        <!-- Full Name -->
        <div>
            <label
                for="fullName"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Full Name
            </label>
            <input
                id="fullName"
                type="text"
                bind:value={fullName}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="John Doe"
            />
        </div>

        <!-- Email -->
        <div>
            <label
                for="email"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Email
            </label>
            <input
                id="email"
                type="email"
                bind:value={email}
                class="w-full px-4 py-2 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                placeholder="john@example.com"
            />
            {#if emailError}
                <p class="text-xs text-red-500 mt-1">{emailError}</p>
            {/if}
        </div>

        <!-- Password -->
        <div>
            <label
                for="password"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Password <span class="text-red-500">*</span>
            </label>
            <div class="relative">
                <input
                    id="password"
                    type={showPassword ? "text" : "password"}
                    bind:value={password}
                    class="w-full px-4 py-2 pr-10 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                    placeholder="••••••••"
                    required
                />
                <button
                    type="button"
                    onclick={() => (showPassword = !showPassword)}
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                    {#if showPassword}
                        <EyeOff class="w-4 h-4" />
                    {:else}
                        <Eye class="w-4 h-4" />
                    {/if}
                </button>
            </div>
            {#if password}
                <div class="mt-2">
                    <div class="flex justify-between items-center mb-1">
                        <span class="text-xs text-muted-foreground"
                            >Password Strength:</span
                        >
                        <span
                            class="text-xs font-medium {passwordStrength() < 40
                                ? 'text-red-500'
                                : passwordStrength() < 70
                                  ? 'text-yellow-500'
                                  : 'text-green-500'}"
                        >
                            {getStrengthLabel(passwordStrength())}
                        </span>
                    </div>
                    <div class="w-full bg-muted rounded-full h-1.5">
                        <div
                            class="h-1.5 rounded-full transition-all {getStrengthColor(
                                passwordStrength(),
                            )}"
                            style="width: {passwordStrength()}%"
                        ></div>
                    </div>
                </div>
            {/if}
            {#if passwordError}
                <p class="text-xs text-red-500 mt-1">{passwordError}</p>
            {/if}
        </div>

        <!-- Confirm Password -->
        <div>
            <label
                for="confirmPassword"
                class="block text-sm font-medium text-foreground mb-1"
            >
                Confirm Password <span class="text-red-500">*</span>
            </label>
            <div class="relative">
                <input
                    id="confirmPassword"
                    type={showConfirmPassword ? "text" : "password"}
                    bind:value={confirmPassword}
                    class="w-full px-4 py-2 pr-10 bg-background/50 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-foreground"
                    placeholder="••••••••"
                    required
                />
                <button
                    type="button"
                    onclick={() => (showConfirmPassword = !showConfirmPassword)}
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                    {#if showConfirmPassword}
                        <EyeOff class="w-4 h-4" />
                    {:else}
                        <Eye class="w-4 h-4" />
                    {/if}
                </button>
            </div>
            {#if confirmPasswordError}
                <p class="text-xs text-red-500 mt-1">{confirmPasswordError}</p>
            {/if}
        </div>

        <!-- Account Type -->
        <div>
            <label class="block text-sm font-medium text-foreground mb-2">
                Account Type
            </label>
            <div class="flex gap-3">
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={accountType}
                        value="virtual"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">Virtual</span>
                    <span class="text-xs text-muted-foreground"
                        >(SMB/NFS only)</span
                    >
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                    <input
                        type="radio"
                        bind:group={accountType}
                        value="system"
                        class="text-primary focus:ring-primary"
                    />
                    <span class="text-sm text-foreground">System</span>
                    <span class="text-xs text-muted-foreground"
                        >(Full Linux user)</span
                    >
                </label>
            </div>
        </div>

        <!-- Is Admin -->
        <div class="flex items-center gap-2">
            <input
                id="isAdmin"
                type="checkbox"
                bind:checked={isAdmin}
                class="rounded text-primary focus:ring-primary"
            />
            <label for="isAdmin" class="text-sm text-foreground cursor-pointer">
                Grant administrator privileges
            </label>
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
                    Create User
                {/if}
            </button>
        </div>
    </form>
</div>
