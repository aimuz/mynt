<script lang="ts">
    import { onMount } from "svelte";
    import {
        api,
        type StorageSpace,
        type Snapshot,
        type SnapshotPolicy,
    } from "$lib/api";
    import DatasetSelector from "./snapshot/DatasetSelector.svelte";
    import SnapshotListPanel from "./snapshot/SnapshotListPanel.svelte";
    import PolicyListPanel from "./snapshot/PolicyListPanel.svelte";

    let datasets = $state<StorageSpace[]>([]);
    let snapshots = $state<Snapshot[]>([]);
    let policies = $state<SnapshotPolicy[]>([]);
    let selectedDataset = $state<string>("");
    let loading = $state(true);
    let snapshotsLoading = $state(false);
    let policiesLoading = $state(false);
    let activeTab = $state<"snapshots" | "policies">("snapshots");

    onMount(async () => {
        await loadDatasets();
        await loadPolicies();
    });

    async function loadDatasets() {
        try {
            loading = true;
            datasets = (await api.listDatasets().catch(() => [])) || [];
            if (datasets.length > 0 && !selectedDataset) {
                selectedDataset = datasets[0].name;
                await loadSnapshots();
            }
        } catch (err) {
            console.error("Failed to load datasets:", err);
        } finally {
            loading = false;
        }
    }

    async function loadSnapshots() {
        if (!selectedDataset) return;

        try {
            snapshotsLoading = true;
            snapshots =
                (await api.listSnapshots(selectedDataset).catch(() => [])) ||
                [];
        } catch (err) {
            console.error("Failed to load snapshots:", err);
        } finally {
            snapshotsLoading = false;
        }
    }

    async function loadPolicies() {
        try {
            policiesLoading = true;
            policies = (await api.listSnapshotPolicies().catch(() => [])) || [];
        } catch (err) {
            console.error("Failed to load policies:", err);
        } finally {
            policiesLoading = false;
        }
    }

    async function handleDatasetChange(datasetName: string) {
        selectedDataset = datasetName;
        await loadSnapshots();
    }

    function handleTabChange(tab: "snapshots" | "policies") {
        activeTab = tab;
    }
</script>

<div class="flex h-full">
    <!-- Left Sidebar -->
    <DatasetSelector
        {datasets}
        {policies}
        {selectedDataset}
        {activeTab}
        {loading}
        onDatasetChange={handleDatasetChange}
        onTabChange={handleTabChange}
    />

    <!-- Right Panel -->
    {#if activeTab === "snapshots"}
        <SnapshotListPanel
            {selectedDataset}
            {snapshots}
            loading={snapshotsLoading}
            onRefresh={loadSnapshots}
        />
    {:else}
        <PolicyListPanel
            {policies}
            {datasets}
            loading={policiesLoading}
            onRefresh={loadPolicies}
        />
    {/if}
</div>
