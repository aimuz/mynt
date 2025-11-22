<script lang="ts">
    let currentTime = $state(new Date());

    setInterval(() => {
        currentTime = new Date();
    }, 1000);

    const hours = $derived(currentTime.getHours());
    const minutes = $derived(currentTime.getMinutes());
    const seconds = $derived(currentTime.getSeconds());

    const formattedTime = $derived(
        `${hours.toString().padStart(2, "0")}:${minutes.toString().padStart(2, "0")}`,
    );

    const formattedDate = $derived(
        currentTime.toLocaleDateString("en-US", {
            weekday: "short",
            month: "short",
            day: "numeric",
        }),
    );

    const timeOfDay = $derived(
        hours < 12 ? "Morning" : hours < 18 ? "Afternoon" : "Evening",
    );
</script>

<div class="text-center">
    <div class="text-4xl font-bold text-foreground mb-1">{formattedTime}</div>
    <div class="text-sm text-foreground/60">{formattedDate}</div>
    <div class="text-xs text-foreground/40 mt-2">Good {timeOfDay}</div>
</div>
