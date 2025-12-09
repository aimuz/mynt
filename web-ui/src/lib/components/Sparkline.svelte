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

    // 固定 viewBox 尺寸
    const viewBoxWidth = 200;
    const viewBoxHeight = height;
    const padding = 2;
    const chartWidth = viewBoxWidth - padding * 2;
    const chartHeight = viewBoxHeight - padding * 2;

    // 生成唯一 ID（只生成一次）
    const gradientId = `sparkline-gradient-${Math.random().toString(36).slice(2, 9)}`;

    // 合并路径计算，一次遍历生成两个路径
    let paths = $derived(() => {
        if (!data || data.length < 2) return { linePath: "", fillPath: "" };

        // 计算数据范围
        let max = data[0];
        let min = data[0];
        for (let i = 1; i < data.length; i++) {
            if (data[i] > max) max = data[i];
            if (data[i] < min) min = data[i];
        }
        max = Math.max(max, 1);
        min = Math.min(min, 0);
        const range = max - min || 1;

        // 预计算常量
        const dataLength = data.length;
        const xStep = chartWidth / (dataLength - 1);
        const bottomY = padding + chartHeight;

        // 构建路径字符串（使用数组 join 比字符串拼接快）
        const lineSegments: string[] = [];
        let firstX = 0;
        let lastX = 0;

        for (let i = 0; i < dataLength; i++) {
            const x = padding + i * xStep;
            const y =
                padding + chartHeight - ((data[i] - min) / range) * chartHeight;

            if (i === 0) {
                lineSegments.push(`M ${x},${y}`);
                firstX = x;
            } else {
                lineSegments.push(`L ${x},${y}`);
            }

            if (i === dataLength - 1) {
                lastX = x;
            }
        }

        const linePath = lineSegments.join(" ");

        // 生成填充路径（仅在需要时）
        const fillPath = showFill
            ? `${linePath} L ${lastX},${bottomY} L ${firstX},${bottomY} Z`
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

    {#if showFill && paths().fillPath}
        <path
            d={paths().fillPath}
            fill={fillColor || `url(#${gradientId})`}
            stroke="none"
        />
    {/if}

    {#if paths().linePath}
        <path
            d={paths().linePath}
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
