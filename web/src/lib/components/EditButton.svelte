<script lang="ts">
    import { SquarePen } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";

    let { activeTab = $bindable() }: { activeTab: string } = $props();
</script>

<Button
    variant="outline"
    size="sm"
    class="h-7 gap-1.5 px-2 cursor-pointer"
    onclick={() => {
        activeTab = "edit";
    }}
>
    <SquarePen class="h-3.5 w-3.5" />
    <p>Bearbeiten</p>
</Button>
