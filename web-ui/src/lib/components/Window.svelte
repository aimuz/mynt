<script lang="ts">
    import { onMount } from "svelte";

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
    let dragStart = $state({ x: 0, y: 0 });
    let windowPos = $state({ x, y });
    let windowSize = $state({ width, height });

    function handleMouseDownDrag(e: MouseEvent) {
        if ((e.target as HTMLElement).closest(".window-controls")) return;
        isDragging = true;
        dragStart = {
            x: e.clientX - windowPos.x,
            y: e.clientY - windowPos.y,
        };
    }

    function handleMouseMove(e: MouseEvent) {
        if (isDragging) {
            windowPos = {
                x: Math.max(
                    0,
                    Math.min(e.clientX - dragStart.x, window.innerWidth - 100),
                ),
                y: Math.max(
                    0,
                    Math.min(e.clientY - dragStart.y, window.innerHeight - 100),
                ),
            };
        }
    }

    function handleMouseUp() {
        isDragging = false;
        isResizing = false;
    }

    onMount(() => {
        document.addEventListener("mousemove", handleMouseMove);
        document.addEventListener("mouseup", handleMouseUp);

        return () => {
            document.removeEventListener("mousemove", handleMouseMove);
            document.removeEventListener("mouseup", handleMouseUp);
        };
    });
</script>

<div
    bind:this={windowEl}
    class="fixed window-shadow rounded-xl overflow-hidden"
    style="
    left: {windowPos.x}px;
    top: {windowPos.y}px;
    width: {windowSize.width}px;
    height: {windowSize.height}px;
    min-width: {minWidth}px;
    min-height: {minHeight}px;
    z-index: 100;
  "
>
    <!-- Window Header -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
        class="glass-strong h-12 flex items-center justify-between px-4 cursor-move border-b border-white/10"
        onmousedown={handleMouseDownDrag}
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
                class="w-3 h-3 rounded-full bg-yellow-400 hover:bg-yellow-500 transition-colors"
                title="Minimize"
            ></button>
            <button
                class="w-3 h-3 rounded-full bg-green-400 hover:bg-green-500 transition-colors"
                title="Maximize"
            ></button>
            <button
                onclick={onClose}
                class="w-3 h-3 rounded-full bg-red-400 hover:bg-red-500 transition-colors"
                title="Close"
            ></button>
        </div>
    </div>

    <!-- Window Content -->
    <div class="glass h-full overflow-auto">
        {@render children()}
    </div>
</div>
