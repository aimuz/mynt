<script lang="ts">
    import { HardDrive } from "@lucide/svelte";
</script>

<script module lang="ts">
    import { HardDrive as AppIcon } from "@lucide/svelte";

    export function launch(api: any, component: any) {
        api.openWindow({
            id: "disks",
            title: "Disk Manager",
            icon: AppIcon,
            component: component,
            width: 800,
            height: 600
        });
    }
</script>

<div class="flex items-center justify-center h-full">
    <div class="text-center text-muted-foreground">
        <HardDrive class="w-16 h-16 mx-auto mb-4 opacity-50" />
        <p class="text-lg">Coming Soon</p>
        <p class="text-sm mt-2">Disk Manager is under development</p>
    </div>
</div>
