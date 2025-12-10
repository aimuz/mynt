<script lang="ts">
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

    // Use a fixed viewBox for consistent sizing
    const viewBoxWidth = 200;
    const viewBoxHeight = height;
    const padding = 2;
    const chartWidth = viewBoxWidth - padding * 2;
    const chartHeight = viewBoxHeight - padding * 2;

    // Generate unique ID for gradient
    const gradientId = `sparkline-gradient-${Math.random().toString(36).slice(2, 9)}`;

    // Combine path calculations into a single pass to generate both paths
    let paths = $derived.by(() => {
        if (!data || data.length < 2) return { linePath: "", fillPath: "" };

        // Calculate data range in single pass
        let max = data[0];
        let min = data[0];
        for (let i = 1; i < data.length; i++) {
            if (data[i] > max) max = data[i];
            if (data[i] < min) min = data[i];
        }
        max = Math.max(max, 1);
        min = Math.min(min, 0);

        // Pre-calculate constants outside loop
        const dataLength = data.length;
        const xStep = chartWidth / (dataLength - 1);
        const baseY = padding + chartHeight;
        const yScale = chartHeight / (max - min || 1);

        // Build path string directly (avoid array allocation + join overhead)
        const firstX = padding;
        const lastX = padding + (dataLength - 1) * xStep;
        let linePath = `M ${firstX},${baseY - (data[0] - min) * yScale}`;

        for (let i = 1; i < dataLength; i++) {
            linePath += ` L ${padding + i * xStep},${baseY - (data[i] - min) * yScale}`;
        }

        // Generate fill path (only when needed)
        const fillPath = showFill
            ? `${linePath} L ${lastX},${baseY} L ${firstX},${baseY} Z`
            : "";

        return { linePath, fillPath };
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

    {#if showFill && paths.fillPath}
        <path
            d={paths.fillPath}
            fill={fillColor || `url(#${gradientId})`}
            stroke="none"
        />
    {/if}

    {#if paths.linePath}
        <path
            d={paths.linePath}
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
