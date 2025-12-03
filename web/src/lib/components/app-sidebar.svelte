<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Group from "./Group.svelte";

    let { userId = $bindable() }: { userId: string } = $props();
</script>

<Sidebar.Root>
    <Sidebar.Content>
        <Group bind:userId></Group>
    </Sidebar.Content>
</Sidebar.Root>
