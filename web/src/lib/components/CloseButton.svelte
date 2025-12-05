<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { X } from "@lucide/svelte";

    export let onCloseClick: () => void = () => {};
    export let sizeClass = "h-6 w-6 p-0";
    export let variant = "ghost";
    export let btnSize = "sm";
</script>

<Button
    variant={variant}
    size={btnSize}
    class={sizeClass}
    onclick={onCloseClick}
>
    <X class="h-4 w-4" />
</Button>
