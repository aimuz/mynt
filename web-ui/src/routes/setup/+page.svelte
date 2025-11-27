<script lang="ts">
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { Server, User, Mail, Lock, ArrowRight } from "@lucide/svelte";

    let step = $state(1);
    let username = $state("");
    let password = $state("");
    let confirmPassword = $state("");
    let fullName = $state("");
    let email = $state("");
    let error = $state("");
    let loading = $state(false);

    async function handleSetup() {
        // Validation
        if (!username || !password || !confirmPassword) {
            error = "Please fill in all required fields";
            return;
        }

        if (password !== confirmPassword) {
            error = "Passwords do not match";
            return;
        }

        if (password.length < 6) {
            error = "Password must be at least 6 characters";
            return;
        }

        loading = true;
        error = "";

        try {
            await api.setup(username, password, fullName, email);
            goto("/desktop");
        } catch (err) {
            error = err instanceof Error ? err.message : "Setup failed";
        } finally {
            loading = false;
        }
    }

    function nextStep() {
        if (step === 1) {
            if (!username) {
                error = "Please enter a username";
                return;
            }
            error = "";
            step = 2;
        }
    }

    function prevStep() {
        if (step > 1) {
            error = "";
            step--;
        }
    }
</script>

<div class="min-h-screen desktop-bg flex items-center justify-center p-4">
    <!-- Setup Card -->
    <div
        class="glass-strong rounded-3xl p-8 w-full max-w-2xl window-shadow scale-in"
    >
        <!-- Header -->
        <div class="text-center mb-8">
            <div
                class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-linear-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg"
            >
                <Server class="w-10 h-10 text-white" />
            </div>
            <h1 class="text-3xl font-bold text-foreground mb-2">
                Welcome to Mynt NAS
            </h1>
            <p class="text-sm text-foreground/60">
                Let's set up your new storage system
            </p>
        </div>

        <!-- Progress Steps -->
        <div class="flex items-center justify-center mb-8">
            <div class="flex items-center gap-2">
                <div
                    class={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold ${step >= 1 ? "bg-blue-500 text-white" : "bg-foreground/10 text-foreground/50"}`}
                >
                    1
                </div>
                <div
                    class={`w-16 h-0.5 ${step >= 2 ? "bg-blue-500" : "bg-foreground/10"}`}
                ></div>
                <div
                    class={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-semibold ${step >= 2 ? "bg-blue-500 text-white" : "bg-foreground/10 text-foreground/50"}`}
                >
                    2
                </div>
            </div>
        </div>

        <!-- Error Message -->
        {#if error}
            <div
                class="mb-6 p-3 rounded-lg bg-red-500/10 border border-red-500/20 fade-in"
            >
                <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
            </div>
        {/if}

        <!-- Step 1: Account Info -->
        {#if step === 1}
            <div class="space-y-6 fade-in">
                <div class="text-center mb-6">
                    <h2 class="text-xl font-semibold text-foreground mb-2">
                        Create Admin Account
                    </h2>
                    <p class="text-sm text-foreground/60">
                        This will be the main administrator account
                    </p>
                </div>

                <div>
                    <label
                        for="username"
                        class="text-sm font-medium text-foreground mb-2 flex items-center gap-2"
                    >
                        <User class="w-4 h-4" />
                        Username *
                    </label>
                    <input
                        id="username"
                        type="text"
                        bind:value={username}
                        placeholder="admin"
                        class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    />
                </div>

                <div>
                    <label
                        for="fullName"
                        class="text-sm font-medium text-foreground mb-2 flex items-center gap-2"
                    >
                        <User class="w-4 h-4" />
                        Full Name
                    </label>
                    <input
                        id="fullName"
                        type="text"
                        bind:value={fullName}
                        placeholder="John Doe"
                        class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    />
                </div>

                <div>
                    <label
                        for="email"
                        class="text-sm font-medium text-foreground mb-2 flex items-center gap-2"
                    >
                        <Mail class="w-4 h-4" />
                        Email
                    </label>
                    <input
                        id="email"
                        type="email"
                        bind:value={email}
                        placeholder="admin@example.com"
                        class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    />
                </div>

                <button
                    onclick={nextStep}
                    class="w-full py-3 rounded-xl bg-linear-to-r from-blue-500 to-purple-600 text-white font-semibold hover:shadow-lg hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2"
                >
                    <span>Continue</span>
                    <ArrowRight class="w-5 h-5" />
                </button>
            </div>
        {/if}

        <!-- Step 2: Password -->
        {#if step === 2}
            <div class="space-y-6 fade-in">
                <div class="text-center mb-6">
                    <h2 class="text-xl font-semibold text-foreground mb-2">
                        Set Password
                    </h2>
                    <p class="text-sm text-foreground/60">
                        Choose a strong password for your account
                    </p>
                </div>

                <div>
                    <label
                        for="password"
                        class="text-sm font-medium text-foreground mb-2 flex items-center gap-2"
                    >
                        <Lock class="w-4 h-4" />
                        Password *
                    </label>
                    <input
                        id="password"
                        type="password"
                        bind:value={password}
                        placeholder="Enter password (min. 6 characters)"
                        class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    />
                </div>

                <div>
                    <label
                        for="confirmPassword"
                        class="text-sm font-medium text-foreground mb-2 flex items-center gap-2"
                    >
                        <Lock class="w-4 h-4" />
                        Confirm Password *
                    </label>
                    <input
                        id="confirmPassword"
                        type="password"
                        bind:value={confirmPassword}
                        placeholder="Re-enter password"
                        class="w-full px-4 py-3 rounded-xl glass-card border border-border focus:border-primary focus:ring-2 focus:ring-primary/20 outline-none transition-all text-foreground placeholder:text-foreground/40"
                    />
                </div>

                <!-- Password Strength Indicator -->
                {#if password}
                    <div
                        class="glass-card rounded-lg p-3 border border-border/50"
                    >
                        <div
                            class="text-xs font-medium text-foreground/70 mb-2"
                        >
                            Password Strength
                        </div>
                        <div class="flex gap-1">
                            <div
                                class={`h-1 flex-1 rounded-full ${password.length >= 6 ? "bg-green-500" : "bg-foreground/10"}`}
                            ></div>
                            <div
                                class={`h-1 flex-1 rounded-full ${password.length >= 8 ? "bg-green-500" : "bg-foreground/10"}`}
                            ></div>
                            <div
                                class={`h-1 flex-1 rounded-full ${password.length >= 12 ? "bg-green-500" : "bg-foreground/10"}`}
                            ></div>
                        </div>
                    </div>
                {/if}

                <div class="flex gap-3">
                    <button
                        onclick={prevStep}
                        class="flex-1 py-3 rounded-xl glass-card border border-border hover:bg-white/5 transition-all text-foreground font-semibold"
                    >
                        Back
                    </button>
                    <button
                        onclick={handleSetup}
                        disabled={loading}
                        class="flex-1 py-3 rounded-xl bg-linear-to-br from-blue-500 to-purple-600 text-white font-semibold hover:shadow-lg hover:scale-[1.02] active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    >
                        {#if loading}
                            <div
                                class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"
                            ></div>
                            <span>Setting up...</span>
                        {:else}
                            <span>Complete Setup</span>
                        {/if}
                    </button>
                </div>
            </div>
        {/if}

        <!-- Footer -->
        <div class="mt-8 pt-6 border-t border-foreground/10 text-center">
            <p class="text-xs text-foreground/50">
                Mynt NAS v1.0 - Enterprise-grade Home Storage
            </p>
        </div>
    </div>
</div>
