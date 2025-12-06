<script lang="ts">
    import { CircleAlert, RotateCw } from "@lucide/svelte";

    interface WindowProps {
        id: string;
        title: string;
        icon?: any;
        children: any;
        x?: number;
        y?: number;
        width?: number;
        height?: number;
        minWidth?: number;
        minHeight?: number;
        zIndex?: number;
        onClose?: () => void;
        onFocus?: () => void;
    }

    let {
        id,
        title,
        icon,
        children,
        x = 100,
        y = 100,
        width = 800,
        height = 600,
        minWidth = 400,
        minHeight = 300,
        zIndex = 100,
        onClose,
        onFocus,
    }: WindowProps = $props();

    let windowEl: HTMLDivElement;
    let isDragging = $state(false);
    let isResizing = $state(false);
    let resizeDirection = $state("");
    let dragStart = $state({ x: 0, y: 0 });
    let windowPos = $state({ x, y });
    let windowSize = $state({ width, height });
    let resizeStart = $state({ x: 0, y: 0, width: 0, height: 0 });

    let rafId: number | null = null;
    let cachedBounds = { maxX: 0, maxY: 0 };
    let pendingPos: { x: number; y: number } | null = null;
    let pendingSize: { width: number; height: number } | null = null;

    const clamp = (val: number, min: number, max: number) =>
        Math.max(min, Math.min(val, max));

    const scheduleUpdate = (update: () => void) => {
        if (rafId === null) {
            rafId = requestAnimationFrame(() => {
                update();
                rafId = null;
            });
        }
    };

    const addListeners = () => {
        document.addEventListener("pointermove", handlePointerMove);
        document.addEventListener("pointerup", handlePointerUp);
    };

    const removeListeners = () => {
        document.removeEventListener("pointermove", handlePointerMove);
        document.removeEventListener("pointerup", handlePointerUp);
    };

    function handlePointerDownDrag(e: PointerEvent) {
        if ((e.target as HTMLElement).closest(".window-controls")) return;
        e.preventDefault();

        onFocus?.();
        isDragging = true;
        dragStart = { x: e.clientX - windowPos.x, y: e.clientY - windowPos.y };
        cachedBounds = {
            maxX: window.innerWidth - 100,
            maxY: window.innerHeight - 100,
        };
        addListeners();
    }

    function handlePointerDownResize(e: PointerEvent, direction: string) {
        e.preventDefault();
        e.stopPropagation();

        onFocus?.();
        isResizing = true;
        resizeDirection = direction;
        resizeStart = {
            x: e.clientX,
            y: e.clientY,
            width: windowSize.width,
            height: windowSize.height,
        };
        addListeners();
    }

    function handlePointerMove(e: PointerEvent) {
        if (isDragging) {
            const x = clamp(e.clientX - dragStart.x, 0, cachedBounds.maxX);
            const y = clamp(e.clientY - dragStart.y, 0, cachedBounds.maxY);

            pendingPos = { x, y };

            scheduleUpdate(() => {
                if (windowEl && pendingPos) {
                    windowEl.style.transform = `translate(${pendingPos.x}px, ${pendingPos.y}px)`;
                }
            });
        } else if (isResizing) {
            const deltaX = e.clientX - resizeStart.x;
            const deltaY = e.clientY - resizeStart.y;
            const dir = resizeDirection;

            const width = dir.includes("e")
                ? Math.max(minWidth, resizeStart.width + deltaX)
                : dir.includes("w")
                  ? Math.max(minWidth, resizeStart.width - deltaX)
                  : windowSize.width;

            const height = dir.includes("s")
                ? Math.max(minHeight, resizeStart.height + deltaY)
                : dir.includes("n")
                  ? Math.max(minHeight, resizeStart.height - deltaY)
                  : windowSize.height;

            pendingSize = { width, height };

            scheduleUpdate(() => {
                if (windowEl && pendingSize) {
                    windowEl.style.width = `${pendingSize.width}px`;
                    windowEl.style.height = `${pendingSize.height}px`;
                }
            });
        }
    }

    function handlePointerUp() {
        isDragging = false;
        isResizing = false;
        resizeDirection = "";

        if (rafId !== null) {
            cancelAnimationFrame(rafId);
            rafId = null;
        }

        if (pendingPos) {
            windowPos = pendingPos;
            pendingPos = null;
        }
        if (pendingSize) {
            windowSize = pendingSize;
            pendingSize = null;
        }

        removeListeners();
    }

    function handleError(error: unknown, reset: () => void) {
        console.error(`[Window: ${title}] Error caught:`, error);
    }
</script>

<div
    bind:this={windowEl}
    class="fixed window-shadow rounded-xl overflow-hidden flex flex-col desktop-window"
    onpointerdown={() => onFocus?.()}
    style:left="0"
    style:top="0"
    style:transform="translate({windowPos.x}px, {windowPos.y}px)"
    style:width="{windowSize.width}px"
    style:height="{windowSize.height}px"
    style:min-width="{minWidth}px"
    style:min-height="{minHeight}px"
    style:z-index={zIndex}
    style:will-change={isDragging || isResizing ? "transform" : "auto"}
    class:dragging={isDragging}
    class:resizing={isResizing}
>
    <!-- Window Header -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
        class="glass-strong h-12 flex items-center justify-between px-4 cursor-move border-b border-white/10 shrink-0"
        onpointerdown={handlePointerDownDrag}
        role="group"
        aria-label="Window Header"
    >
        <div class="flex items-center gap-3">
            {#if icon}
                {@const SvelteComponent = icon}
                <SvelteComponent class="w-5 h-5 text-foreground/70" />
            {/if}
            <span class="font-semibold text-sm text-foreground">{title}</span>
        </div>

        <!-- Window Controls (macOS style) -->
        <div class="window-controls flex items-center gap-2">
            <button
                class="w-3 h-3 rounded-full bg-yellow-400 hover:bg-yellow-500 transition-colors border-0 outline-none shadow-none"
                title="Minimize"
                aria-label="Minimize Window"
            ></button>
            <button
                class="w-3 h-3 rounded-full bg-green-400 hover:bg-green-500 transition-colors border-0 outline-none shadow-none"
                title="Maximize"
                aria-label="Maximize Window"
            ></button>
            <button
                onclick={onClose}
                class="w-3 h-3 rounded-full bg-red-400 hover:bg-red-500 transition-colors border-0 outline-none shadow-none"
                title="Close"
                aria-label="Close Window"
            ></button>
        </div>
    </div>

    <!-- Window Content -->
    <div class="glass flex-1 overflow-auto">
        <svelte:boundary onerror={handleError}>
            {@render children()}
            {#snippet failed(error, reset)}
                <div
                    class="flex flex-col items-center justify-center h-full text-center p-6"
                >
                    <CircleAlert class="w-12 h-12 text-red-400 mb-4" />
                    <h3 class="text-lg font-semibold text-foreground mb-2">
                        Application Error
                    </h3>
                    <p class="text-sm text-foreground/70 mb-1">
                        Something went wrong in this window
                    </p>
                    <p
                        class="text-xs text-foreground/50 mb-6 max-w-md break-all bg-black/10 p-2 rounded-md font-mono"
                        title={error instanceof Error
                            ? error.message
                            : String(error)}
                    >
                        {error instanceof Error ? error.message : String(error)}
                    </p>
                    <div class="flex gap-3">
                        <button
                            onclick={onClose}
                            class="px-4 py-2 text-sm text-foreground/70 hover:text-foreground transition-colors"
                        >
                            Close
                        </button>
                        <button
                            onclick={reset}
                            class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-all shadow-lg"
                        >
                            <RotateCw class="w-4 h-4" />
                            Try Again
                        </button>
                    </div>
                </div>
            {/snippet}
        </svelte:boundary>
    </div>

    <!-- Resize Handles -->
    <div
        class="absolute right-0 bottom-0 w-4 h-4 cursor-se-resize"
        onpointerdown={(e) => handlePointerDownResize(e, "se")}
        role="button"
        tabindex="-1"
        aria-label="Resize Window"
    ></div>
    <div
        class="absolute right-0 top-12 bottom-0 w-1 cursor-e-resize"
        onpointerdown={(e) => handlePointerDownResize(e, "e")}
        role="button"
        tabindex="-1"
        aria-label="Resize Window Horizontally"
    ></div>
    <div
        class="absolute left-0 right-0 bottom-0 h-1 cursor-s-resize"
        onpointerdown={(e) => handlePointerDownResize(e, "s")}
        role="button"
        tabindex="-1"
        aria-label="Resize Window Vertically"
    ></div>
</div>

<style>
    div[style*="transform"] {
        transform: translate3d(0, 0, 0);
        backface-visibility: hidden;
    }

    .dragging,
    .resizing {
        transition: none !important;
        user-select: none;
        pointer-events: auto;
    }

    div[role="button"][aria-label*="Resize"] {
        opacity: 0;
        transition: opacity 0.2s;
    }

    div:hover > div[role="button"][aria-label*="Resize"] {
        opacity: 1;
    }

    .window-shadow {
        contain: layout style paint;
    }
</style>
