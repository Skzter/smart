<script lang="ts">
    import * as UnderlineTabs from "$lib/components/ui/underline-tabs";

    let {
        activeTab = $bindable("run"),
        curTest,
    }: { activeTab: string; curTest: string } = $props();

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
        <div class="ml-auto text-sm">{curTest}</div>
    </UnderlineTabs.List>
</UnderlineTabs.Root>
