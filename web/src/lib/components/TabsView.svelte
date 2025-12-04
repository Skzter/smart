<script lang="ts">
    export let activeTab = "run";
    import * as UnderlineTabs from "$lib/components/ui/underline-tabs";

    function handleTabClick(tab: string) {
        activeTab = tab;
    }
</script>

<UnderlineTabs.Root
    value={activeTab}
    onValueChange={handleTabClick}
    class="px-6"
>
    <UnderlineTabs.List>
        <UnderlineTabs.Trigger value="edit">Edit</UnderlineTabs.Trigger>
        <UnderlineTabs.Trigger value="run">Run</UnderlineTabs.Trigger>
        <UnderlineTabs.Trigger value="result">Result</UnderlineTabs.Trigger>
    </UnderlineTabs.List>
</UnderlineTabs.Root>
