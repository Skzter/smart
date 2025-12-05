<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { X } from "@lucide/svelte";

    export let onCloseClick: () => void = () => {};
    export let sizeClass = "h-6 w-6 p-0";
</script>

<Button
    variant={"ghost"}
    size={"sm"}
    class={sizeClass}
    onclick={onCloseClick}
>
    <X class="h-4 w-4" />
</Button>
