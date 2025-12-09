<script lang="ts">
    let { data, color, autoScale = false } = $props<{
        data: number[],
        color: string,
        autoScale?: boolean
    }>();

    function renderGraph(d: number[]) {
        const height = 60;
        const width = 100;
        let max = 100; // Default Percent

        if (autoScale) {
            max = Math.max(...d, 1); // Avoid division by zero
        }

        const points = d.map((val, i) => {
            const x = (i / (d.length - 1)) * width;
            const y = height - (val / max) * height;
            return `${x},${y}`;
        }).join(" ");

        return `
            <svg viewBox="0 0 ${width} ${height}" class="w-full h-full overflow-visible" preserveAspectRatio="none">
                <path d="M0,${height} ${points} L${width},${height} Z" fill="${color}" fill-opacity="0.2" />
                <polyline points="${points}" fill="none" stroke="${color}" stroke-width="2" vector-effect="non-scaling-stroke" />
            </svg>
        `;
    }
</script>

{@html renderGraph(data)}
