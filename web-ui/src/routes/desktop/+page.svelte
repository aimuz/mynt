<script lang="ts">
    import {
        Database,
        FolderOpen,
        HardDrive,
        Settings,
        Users,
        Bell,
        Activity,
    } from "@lucide/svelte";
    import Dock, { type DesktopApp } from "$lib/components/Dock.svelte";
    import Window from "$lib/components/Window.svelte";
    import Widget from "$lib/components/Widget.svelte";
    import DashboardApp from "$lib/apps/DashboardApp.svelte";
    import SystemStatus from "$lib/widgets/SystemStatus.svelte";
    import Clock from "$lib/widgets/Clock.svelte";
    import RecentNotifications from "$lib/widgets/RecentNotifications.svelte";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";

    // Check authentication
    onMount(() => {
        if (!api.isAuthenticated()) {
            goto("/login");
        }
    });

    let activeWindows = $state<
        Array<{ id: string; title: string; icon: any; component: any }>
    >([]);
    let currentTime = $state(new Date());

    // Update time every second
    setInterval(() => {
        currentTime = new Date();
    }, 1000);

    const apps: DesktopApp[] = [
        {
            id: "dashboard",
            name: "Dashboard",
            icon: Activity,
            color: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
            onClick: () =>
                openWindow("dashboard", "Dashboard", Activity, DashboardApp),
        },
        {
            id: "storage",
            name: "Storage",
            icon: Database,
            color: "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
            onClick: () =>
                openWindow("storage", "Storage Manager", Database, null),
        },
        {
            id: "shares",
            name: "Shares",
            icon: FolderOpen,
            color: "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
            onClick: () =>
                openWindow("shares", "File Shares", FolderOpen, null),
        },
        {
            id: "disks",
            name: "Disks",
            icon: HardDrive,
            color: "linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)",
            onClick: () => openWindow("disks", "Disk Manager", HardDrive, null),
        },
        {
            id: "users",
            name: "Users",
            icon: Users,
            color: "linear-gradient(135deg, #fa709a 0%, #fee140 100%)",
            onClick: () => openWindow("users", "User Management", Users, null),
        },
        {
            id: "settings",
            name: "Settings",
            icon: Settings,
            color: "linear-gradient(135deg, #30cfd0 0%, #330867 100%)",
            onClick: () => openWindow("settings", "Settings", Settings, null),
        },
    ];

    function openWindow(id: string, title: string, icon: any, component: any) {
        // Check if window is already open
        if (activeWindows.some((w) => w.id === id)) return;

        activeWindows = [...activeWindows, { id, title, icon, component }];
    }

    function closeWindow(id: string) {
        activeWindows = activeWindows.filter((w) => w.id !== id);
    }

    function handleLogout() {
        api.logout();
        goto("/login");
    }
</script>

<div class="relative w-full h-screen desktop-bg overflow-hidden">
    <!-- Menu Bar (macOS style) -->
    <div
        class="fixed top-0 left-0 right-0 glass-strong h-8 flex items-center justify-between px-4 z-50 border-b border-white/10"
    >
        <div class="flex items-center gap-6">
            <div class="flex items-center gap-2">
                <div
                    class="w-5 h-5 rounded-md bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center"
                >
                    <span class="text-white text-xs font-bold">M</span>
                </div>
                <span class="text-sm font-semibold text-foreground"
                    >Mynt NAS</span
                >
            </div>
            <div class="flex items-center gap-4 text-sm text-foreground/70">
                <button class="hover:text-foreground transition-colors"
                    >File</button
                >
                <button class="hover:text-foreground transition-colors"
                    >View</button
                >
                <button class="hover:text-foreground transition-colors"
                    >Tools</button
                >
                <button class="hover:text-foreground transition-colors"
                    >Help</button
                >
            </div>
        </div>

        <div class="flex items-center gap-4">
            <button class="hover:text-foreground transition-colors">
                <Bell class="w-4 h-4" />
            </button>
            <button
                onclick={handleLogout}
                class="text-xs font-medium text-foreground/70 hover:text-foreground transition-colors"
            >
                Logout
            </button>
            <span class="text-xs font-medium text-foreground/70">
                {currentTime.toLocaleTimeString("en-US", {
                    hour: "2-digit",
                    minute: "2-digit",
                })}
            </span>
        </div>
    </div>

    <!-- Desktop Content -->
    <div class="pt-12 p-6 flex gap-6 h-[calc(100vh-8rem)]">
        <!-- Left: Widgets Area - Fixed width on large screens -->
        <div class="w-80 flex-shrink-0 space-y-4 overflow-y-auto">
            <Widget title="System Status" icon={Activity} size="medium">
                {#snippet children()}
                    <SystemStatus />
                {/snippet}
            </Widget>

            <Widget title="Clock" size="small">
                {#snippet children()}
                    <Clock />
                {/snippet}
            </Widget>

            <Widget title="Notifications" icon={Bell} size="medium">
                {#snippet children()}
                    <RecentNotifications />
                {/snippet}
            </Widget>
        </div>

        <!-- Right: App Icons - Flexible grid -->
        <div
            class="flex-1 grid grid-cols-4 xl:grid-cols-8 2xl:grid-cols-12 gap-4 auto-rows-min content-start overflow-y-auto"
        >
            {#each apps as app, i (app.id)}
                <div class="flex items-center justify-center">
                    <button
                        onclick={app.onClick}
                        class="flex flex-col w-28 h-28 items-center justify-center gap-1 p-3 rounded-lg hover:bg-white/10 active:bg-white/20 transition-all group fade-in"
                        style="animation-delay: {i * 50}ms;"
                    >
                        <div
                            class="w-16 h-16 rounded-xl flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform"
                            style="background: {app.color};"
                        >
                            <app.icon
                                class="w-8 h-8 text-white"
                            />
                        </div>
                        <span
                            class="text-xs font-medium text-foreground text-center leading-tight"
                        >
                            {app.name}
                        </span>
                    </button>
                </div>
            {/each}
        </div>
    </div>

    <!-- Active Windows -->
    {#each activeWindows as window (window.id)}
        <Window
            id={window.id}
            title={window.title}
            icon={window.icon}
            onClose={() => closeWindow(window.id)}
            x={100 + activeWindows.indexOf(window) * 30}
            y={100 + activeWindows.indexOf(window) * 30}
        >
            {#snippet children()}
                {#if window.component}
                    <window.component />
                {:else}
                    <div class="flex items-center justify-center h-full">
                        <div class="text-center text-muted-foreground">
                            <window.icon
                                class="w-16 h-16 mx-auto mb-4 opacity-50"
                            />
                            <p class="text-lg">Coming Soon</p>
                            <p class="text-sm mt-2">
                                This app is under development
                            </p>
                        </div>
                    </div>
                {/if}
            {/snippet}
        </Window>
    {/each}

    <!-- Dock -->
    <Dock {apps} />
</div>
