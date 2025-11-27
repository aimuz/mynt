<script lang="ts">
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { currentWallpaper } from "$lib/stores/wallpaper";
    import { LogIn } from "@lucide/svelte";

    let username = $state("");
    let password = $state("");
    let error = $state("");
    let loading = $state(false);

    async function handleLogin() {
        if (!username || !password) {
            error = "Please enter username and password";
            return;
        }

        loading = true;
        error = "";

        try {
            await api.login(username, password);
            goto("/desktop");
        } catch (err) {
            error = err instanceof Error ? err.message : "Login failed";
        } finally {
            loading = false;
        }
    }

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter") {
            handleLogin();
        }
    }

    function getBackgroundStyle(wallpaper: typeof $currentWallpaper) {
        if (wallpaper.type === "image") {
            return `background-image: url('${wallpaper.value}'); background-size: cover; background-position: center; background-repeat: no-repeat;`;
        }
        return `background: ${wallpaper.value};`;
    }
</script>

<div
    class="min-h-screen flex items-center justify-center p-4 transition-all duration-500"
    style={getBackgroundStyle($currentWallpaper)}
>
    <!-- Login Card -->
    <div
        class="glass-strong rounded-3xl p-8 w-full max-w-md window-shadow scale-in"
    >
        <!-- Logo & Title -->
        <div class="text-center mb-8">
            <div
                class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-linear-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg"
            >
                <span class="text-white text-3xl font-bold">M</span>
            </div>
            <h1 class="text-2xl font-bold text-foreground mb-2">
                Welcome to Mynt NAS
            </h1>
            <p class="text-sm text-foreground/60">
                Sign in to access your storage
            </p>
        </div>

        <!-- Error Message -->
        {#if error}
            <div
                class="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/20 fade-in"
            >
                <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
        {/if}

        <!-- Login Form -->
        <div class="space-y-4">
            <div>
                <label
                    for="username"
                    class="block text-sm font-medium text-foreground mb-2"
                >
                    Username
                </label>
                <input
                    id="username"
                    type="text"
                    bind:value={username}
                    onkeypress={handleKeyPress}
                    placeholder="Enter your username"
                    class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    disabled={loading}
                />
            </div>

            <div>
                <label
                    for="password"
                    class="block text-sm font-medium text-foreground mb-2"
                >
                    Password
                </label>
                <input
                    id="password"
                    type="password"
                    bind:value={password}
                    onkeypress={handleKeyPress}
                    placeholder="Enter your password"
                    class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    disabled={loading}
                />
            </div>

            <button
                onclick={handleLogin}
                disabled={loading}
                class="w-full py-3 rounded-xl bg-linear-to-r from-blue-500 to-purple-600 text-white font-semibold hover:shadow-lg hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
                {#if loading}
                    <div
                        class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
                    ></div>
                    <span>Signing in...</span>
                {:else}
                    <LogIn class="w-5 h-5" />
                    <span>Sign In</span>
                {/if}
            </button>
        </div>

        <!-- Footer -->
        <div class="mt-8 text-center text-xs text-foreground/50">
            <p>Mynt NAS - Enterprise-grade Home Storage</p>
        </div>
    </div>
</div>
