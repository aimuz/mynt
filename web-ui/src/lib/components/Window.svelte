<script lang="ts">
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
        onClose?: () => void;
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
        onClose,
    }: WindowProps = $props();

    let windowEl: HTMLDivElement;
    let isDragging = $state(false);
    let isResizing = $state(false);
    let resizeDirection = $state("");
    let dragStart = $state({ x: 0, y: 0 });
    let windowPos = $state({ x, y });
    let windowSize = $state({ width, height });
    let resizeStart = $state({ x: 0, y: 0, width: 0, height: 0 });

    // 性能优化:缓存边界值和RAF ID
    let rafId: number | null = null;
    let cachedBounds = { maxX: 0, maxY: 0 };
    let pendingPosUpdate: { x: number; y: number } | null = null;
    let pendingSizeUpdate: { width: number; height: number } | null = null;

    // 性能优化:使用被动事件监听器选项
    const passiveListener = { passive: true };

    function updateBoundsCache() {
        cachedBounds = {
            maxX: window.innerWidth - 100,
            maxY: window.innerHeight - 100,
        };
    }

    function handlePointerDownDrag(e: PointerEvent) {
        if ((e.target as HTMLElement).closest(".window-controls")) return;

        // 防止默认行为,避免文本选择
        e.preventDefault();

        isDragging = true;
        dragStart = {
            x: e.clientX - windowPos.x,
            y: e.clientY - windowPos.y,
        };

        // 缓存边界值
        updateBoundsCache();

        // 只在开始拖拽时添加事件监听器
        document.addEventListener("pointermove", handlePointerMove);
        document.addEventListener("pointerup", handlePointerUp);
    }

    function handlePointerDownResize(e: PointerEvent, direction: string) {
        e.preventDefault();
        e.stopPropagation();

        isResizing = true;
        resizeDirection = direction;
        resizeStart = {
            x: e.clientX,
            y: e.clientY,
            width: windowSize.width,
            height: windowSize.height,
        };

        document.addEventListener("pointermove", handlePointerMove);
        document.addEventListener("pointerup", handlePointerUp);
    }

    function handlePointerMove(e: PointerEvent) {
        if (isDragging) {
            // 计算新位置
            const newX = Math.max(
                0,
                Math.min(e.clientX - dragStart.x, cachedBounds.maxX),
            );
            const newY = Math.max(
                0,
                Math.min(e.clientY - dragStart.y, cachedBounds.maxY),
            );

            // 使用 requestAnimationFrame 节流更新
            pendingPosUpdate = { x: newX, y: newY };

            if (rafId === null) {
                rafId = requestAnimationFrame(() => {
                    if (pendingPosUpdate) {
                        windowPos = pendingPosUpdate;
                        pendingPosUpdate = null;
                    }
                    rafId = null;
                });
            }
        } else if (isResizing) {
            const deltaX = e.clientX - resizeStart.x;
            const deltaY = e.clientY - resizeStart.y;

            let newWidth = windowSize.width;
            let newHeight = windowSize.height;

            if (resizeDirection.includes("e")) {
                newWidth = Math.max(minWidth, resizeStart.width + deltaX);
            }
            if (resizeDirection.includes("s")) {
                newHeight = Math.max(minHeight, resizeStart.height + deltaY);
            }
            if (resizeDirection.includes("w")) {
                newWidth = Math.max(minWidth, resizeStart.width - deltaX);
            }
            if (resizeDirection.includes("n")) {
                newHeight = Math.max(minHeight, resizeStart.height - deltaY);
            }

            pendingSizeUpdate = { width: newWidth, height: newHeight };

            if (rafId === null) {
                rafId = requestAnimationFrame(() => {
                    if (pendingSizeUpdate) {
                        windowSize = pendingSizeUpdate;
                        pendingSizeUpdate = null;
                    }
                    rafId = null;
                });
            }
        }
    }

    function handlePointerUp() {
        isDragging = false;
        isResizing = false;
        resizeDirection = "";

        // 取消待处理的RAF
        if (rafId !== null) {
            cancelAnimationFrame(rafId);
            rafId = null;
        }

        // 应用最后的更新
        if (pendingPosUpdate) {
            windowPos = pendingPosUpdate;
            pendingPosUpdate = null;
        }
        if (pendingSizeUpdate) {
            windowSize = pendingSizeUpdate;
            pendingSizeUpdate = null;
        }

        // 移除事件监听器
        document.removeEventListener("pointermove", handlePointerMove);
        document.removeEventListener("pointerup", handlePointerUp);
    }
</script>

<div
    bind:this={windowEl}
    class="fixed window-shadow rounded-xl overflow-hidden flex flex-col"
    style:left="0"
    style:top="0"
    style:transform="translate({windowPos.x}px, {windowPos.y}px)"
    style:width="{windowSize.width}px"
    style:height="{windowSize.height}px"
    style:min-width="{minWidth}px"
    style:min-height="{minHeight}px"
    style:z-index="100"
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
        {@render children()}
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
    /* 性能优化:使用硬件加速 */
    div[style*="transform"] {
        transform: translate3d(0, 0, 0);
        backface-visibility: hidden;
    }

    /* 拖拽时禁用过渡效果,提升性能 */
    .dragging,
    .resizing {
        transition: none !important;
        user-select: none;
        pointer-events: auto;
    }

    /* 调整大小手柄在非hover状态下透明 */
    div[role="button"][aria-label*="Resize"] {
        opacity: 0;
        transition: opacity 0.2s;
    }

    div:hover > div[role="button"][aria-label*="Resize"] {
        opacity: 1;
    }

    /* 性能优化:避免重排 */
    .window-shadow {
        contain: layout style paint;
    }
</style>
