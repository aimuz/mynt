<script lang="ts">
    import { Settings } from "@lucide/svelte";
</script>

<script module lang="ts">
    import { Settings as AppIcon } from "@lucide/svelte";

    export function launch(api: any, component: any) {
        api.openWindow({
            id: "settings",
            title: "Settings",
            icon: AppIcon,
            component: component,
            width: 700,
            height: 500
        });
    }
</script>

<div class="flex items-center justify-center h-full">
    <div class="text-center text-muted-foreground">
        <Settings class="w-16 h-16 mx-auto mb-4 opacity-50" />
        <p class="text-lg">Coming Soon</p>
        <p class="text-sm mt-2">Settings is under development</p>
    </div>
</div>
