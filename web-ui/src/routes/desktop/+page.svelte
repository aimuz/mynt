<script lang="ts">
    import {
        Database,
        FolderOpen,
        HardDrive,
        Settings,
        Users,
        Bell,
        Activity,
        Image,
    } from "@lucide/svelte";
    import { setContext } from "svelte";
    import Dock, { type DesktopApp } from "$lib/components/Dock.svelte";
    import Window from "$lib/components/Window.svelte";
    import Widget from "$lib/components/Widget.svelte";
    // Apps are now lazy loaded
    // import DashboardApp from "$lib/apps/DashboardApp.svelte";
    // import StorageApp from "$lib/apps/StorageApp.svelte";
    import SystemStatus from "$lib/widgets/SystemStatus.svelte";
    import Clock from "$lib/widgets/Clock.svelte";
    import RecentNotifications from "$lib/widgets/RecentNotifications.svelte";
    import WallpaperSelector from "$lib/components/WallpaperSelector.svelte";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { currentWallpaper } from "$lib/stores/wallpaper";

    // Check authentication
    onMount(() => {
        if (!api.isAuthenticated()) {
            goto("/login");
        }
    });

    const DEFAULT_WINDOW_WIDTH = 800;
    const DEFAULT_WINDOW_HEIGHT = 600;

    let activeWindows = $state<
        Array<{
            id: string;
            title: string;
            icon: any;
            component: any;
            props?: any;
            zIndex: number;
            x: number;
            y: number;
        }>
    >([]);
    let currentTime = $state(new Date());
    let showContextMenu = $state(false);
    let contextMenuX = $state(0);
    let contextMenuY = $state(0);
    let nextZIndex = $state(100);

    // Update time every second
    onMount(() => {
        const interval = setInterval(() => {
            currentTime = new Date();
        }, 1000);

        return () => {
            clearInterval(interval);
        };
    });

    const apps: DesktopApp[] = [
        {
            id: "dashboard",
            name: "Dashboard",
            icon: Activity,
            color: "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
            onClick: async () => {
                const module = await import("$lib/apps/DashboardApp.svelte");
                openWindow("dashboard", "Dashboard", Activity, module.default);
            },
        },
        {
            id: "storage",
            name: "Storage",
            icon: Database,
            color: "linear-gradient(135deg, #f093fb 0%, #f5576c 100%)",
            onClick: async () => {
                const module = await import("$lib/apps/StorageApp.svelte");
                openWindow(
                    "storage",
                    "Storage Manager",
                    Database,
                    module.default,
                );
            },
        },
        {
            id: "shares",
            name: "Shares",
            icon: FolderOpen,
            color: "linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)",
            onClick: async () => {
                const module = await import(
                    "$lib/apps/ShareManagementApp.svelte"
                );
                openWindow(
                    "shares",
                    "Share Management",
                    FolderOpen,
                    module.default,
                );
            },
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
            onClick: async () => {
                const module = await import(
                    "$lib/apps/UserManagementApp.svelte"
                );
                openWindow("users", "User Management", Users, module.default);
            },
        },
        {
            id: "settings",
            name: "Settings",
            icon: Settings,
            color: "linear-gradient(135deg, #30cfd0 0%, #330867 100%)",
            onClick: () => openWindow("settings", "Settings", Settings, null),
        },
    ];

    function bringToFront(id: string) {
        const window = activeWindows.find((w) => w.id === id);
        if (window) {
            // Optimization: O(1) update instead of O(N) search
            // Only update if not already on top
            if (window.zIndex < nextZIndex - 1) {
                window.zIndex = nextZIndex++;
            }
        }
    }

    function openWindow(
        id: string,
        title: string,
        icon: any,
        component: any | (() => { component: any; props?: any }),
    ) {
        // Check if window is already open
        const existingWindow = activeWindows.find((w) => w.id === id);
        if (existingWindow) {
            bringToFront(id);
            return;
        }

        const zIndex = nextZIndex++;

        // Calculate position based on existing windows to cascade them
        const offset = activeWindows.length * 30;

        let x = 100 + offset;
        let y = 100 + offset;

        if (typeof window !== "undefined") {
            x = Math.max(0, (window.innerWidth - DEFAULT_WINDOW_WIDTH) / 2) + offset;
            y = Math.max(0, (window.innerHeight - DEFAULT_WINDOW_HEIGHT) / 2) + offset;
        }

        // Check if component is a factory function
        if (typeof component === "function" && component.length === 0) {
            try {
                const result = component();
                if (
                    result &&
                    typeof result === "object" &&
                    "component" in result
                ) {
                    activeWindows = [
                        ...activeWindows,
                        {
                            id,
                            title,
                            icon,
                            component: result.component,
                            props: result.props,
                            zIndex,
                            x,
                            y,
                        },
                    ];
                    return;
                }
            } catch (e) {
                // Not a factory, fall through to direct component
            }
        }

        // Direct component
        activeWindows = [
            ...activeWindows,
            { id, title, icon, component, zIndex, x, y },
        ];
    }

    function closeWindow(id: string) {
        activeWindows = activeWindows.filter((w) => w.id !== id);
    }

    // Expose window management to child components via context
    setContext("desktop", {
        openWindow,
        closeWindow,
    });

    function handleLogout() {
        api.logout();
        goto("/login");
    }

    function handleContextMenu(e: MouseEvent) {
        // Check if target is part of excluded UI
        const target = e.target as HTMLElement;
        if (
            target.closest(
                ".desktop-window, .desktop-dock, .desktop-widget, .desktop-menubar, .desktop-icon",
            )
        ) {
            return;
        }

        e.preventDefault();
        contextMenuX = e.clientX;
        contextMenuY = e.clientY;
        showContextMenu = true;
    }

    function closeContextMenu() {
        showContextMenu = false;
    }

    function openWallpaperSelector() {
        openWindow("wallpaper", "Change Wallpaper", Image, () => ({
            component: WallpaperSelector,
            props: {
                onClose: () => closeWindow("wallpaper"),
            },
        }));
        closeContextMenu();
    }

    function getBackgroundStyle(wallpaper: typeof $currentWallpaper) {
        if (wallpaper.type === "image") {
            return `background-image: url('${wallpaper.value}'); background-size: cover; background-position: center; background-repeat: no-repeat;`;
        }
        return `background: ${wallpaper.value};`;
    }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
    class="relative w-full h-screen overflow-hidden transition-all duration-500"
    style={getBackgroundStyle($currentWallpaper)}
    oncontextmenu={handleContextMenu}
    onclick={closeContextMenu}
>
    <!-- Menu Bar (macOS style) -->
    <div
        class="fixed top-0 left-0 right-0 glass-strong h-8 flex items-center justify-between px-4 z-50 border-b border-white/10 desktop-menubar"
    >
        <div class="flex items-center gap-6">
            <div class="flex items-center gap-2">
                <div
                    class="w-5 h-5 rounded-md bg-linear-to-br from-blue-500 to-purple-600 flex items-center justify-center"
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
        <div class="w-80 shrink-0 space-y-4 overflow-y-auto">
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
                        class="flex flex-col w-28 h-28 items-center justify-center gap-1 p-3 rounded-lg hover:bg-white/10 active:bg-white/20 transition-all group fade-in desktop-icon"
                        style="animation-delay: {i * 50}ms;"
                    >
                        <div
                            class="w-16 h-16 rounded-xl flex items-center justify-center shadow-lg group-hover:scale-110 transition-transform"
                            style="background: {app.color};"
                        >
                            <app.icon class="w-8 h-8 text-white" />
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
            zIndex={window.zIndex}
            onClose={() => closeWindow(window.id)}
            onFocus={() => bringToFront(window.id)}
            x={window.x}
            y={window.y}
        >
            {#snippet children()}
                {#if window.component}
                    {#if window.props}
                        <window.component {...window.props} />
                    {:else}
                        <window.component />
                    {/if}
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

    <!-- Context Menu -->
    {#if showContextMenu}
        <div
            class="fixed glass-strong rounded-lg shadow-xl py-1 min-w-48 z-200 border border-white/20 overflow-hidden"
            style="left: {contextMenuX}px; top: {contextMenuY}px;"
        >
            <button
                onclick={openWallpaperSelector}
                class="w-full px-4 py-2 text-left text-sm hover:bg-white/10 transition-colors flex items-center gap-2"
            >
                <div
                    class="w-4 h-4 rounded bg-linear-to-br from-blue-500 to-purple-600"
                ></div>
                Change Wallpaper
            </button>
        </div>
    {/if}

</div>
