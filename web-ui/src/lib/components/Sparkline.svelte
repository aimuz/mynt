<script lang="ts">
    /**
     * A responsive sparkline chart component for displaying historical data trends.
     * Uses SVG with viewBox for width responsiveness.
     */
    import { onMount } from "svelte";

    interface Props {
        /** Array of data points (0-100 range recommended) */
        data: number[];
        /** Height of the chart in pixels */
        height?: number;
        /** Line color (CSS color string) */
        color?: string;
        /** Fill color under the line (CSS color string or 'none') */
        fillColor?: string;
        /** Line width in pixels */
        strokeWidth?: number;
        /** Show a gradient fill under the line */
        showFill?: boolean;
    }

    let {
        data,
        height = 40,
        color = "#3b82f6",
        fillColor,
        strokeWidth = 1.5,
        showFill = true,
    }: Props = $props();

    // Use a fixed viewBox width for consistent path calculations
    const viewBoxWidth = 200;
    const viewBoxHeight = height;

    // Generate unique ID for gradient
    const gradientId = `sparkline-gradient-${Math.random().toString(36).slice(2, 9)}`;

    // Calculate path from data points
    let pathD = $derived(() => {
        if (!data || data.length < 2) return "";

        const padding = 2;
        const chartWidth = viewBoxWidth - padding * 2;
        const chartHeight = viewBoxHeight - padding * 2;

        // Normalize data to chart range
        const max = Math.max(...data, 1);
        const min = Math.min(...data, 0);
        const range = max - min || 1;

        const points = data.map((value, index) => {
            const x = padding + (index / (data.length - 1)) * chartWidth;
            const y =
                padding + chartHeight - ((value - min) / range) * chartHeight;
            return { x, y };
        });

        // Generate smooth path
        const pathParts = points.map((point, i) => {
            if (i === 0) return `M ${point.x},${point.y}`;
            return `L ${point.x},${point.y}`;
        });

        return pathParts.join(" ");
    });

    // Generate fill path (closes the area under the line)
    let fillPathD = $derived(() => {
        if (!data || data.length < 2 || !showFill) return "";

        const padding = 2;
        const chartWidth = viewBoxWidth - padding * 2;
        const chartHeight = viewBoxHeight - padding * 2;

        const max = Math.max(...data, 1);
        const min = Math.min(...data, 0);
        const range = max - min || 1;

        const points = data.map((value, index) => {
            const x = padding + (index / (data.length - 1)) * chartWidth;
            const y =
                padding + chartHeight - ((value - min) / range) * chartHeight;
            return { x, y };
        });

        const linePath = points
            .map((point, i) => {
                if (i === 0) return `M ${point.x},${point.y}`;
                return `L ${point.x},${point.y}`;
            })
            .join(" ");

        // Close the path at the bottom
        const lastX = points[points.length - 1].x;
        const firstX = points[0].x;
        const bottomY = padding + chartHeight;

        return `${linePath} L ${lastX},${bottomY} L ${firstX},${bottomY} Z`;
    });
</script>

<svg
    class="sparkline"
    viewBox="0 0 {viewBoxWidth} {viewBoxHeight}"
    preserveAspectRatio="none"
    style="height: {height}px;"
>
    <defs>
        <linearGradient id={gradientId} x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" style="stop-color:{color};stop-opacity:0.3" />
            <stop offset="100%" style="stop-color:{color};stop-opacity:0.05" />
        </linearGradient>
    </defs>

    {#if showFill && fillPathD()}
        <path
            d={fillPathD()}
            fill={fillColor || `url(#${gradientId})`}
            stroke="none"
        />
    {/if}

    {#if pathD()}
        <path
            d={pathD()}
            fill="none"
            stroke={color}
            stroke-width={strokeWidth}
            stroke-linecap="round"
            stroke-linejoin="round"
            vector-effect="non-scaling-stroke"
        />
    {/if}
</svg>

<style>
    .sparkline {
        display: block;
        width: 100%;
    }
</style>
