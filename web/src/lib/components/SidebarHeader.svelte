<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar";
    import NewChatButton from "./NewChatButton.svelte";
    import TimeFilter from "./TimeFilterButton.svelte";
</script>

<Sidebar.Header class="border-b">
    <div class="flex flex-col gap-2 p-2">
        <NewChatButton />
        <TimeFilter />
    </div>
</Sidebar.Header>
