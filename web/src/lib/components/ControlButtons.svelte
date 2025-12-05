<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { Pause, RotateCcw } from "@lucide/svelte";
    import CloseButton from "./CloseButton.svelte";

    export let onCloseClick: () => void = () => {};
</script>

<div class="flex items-center gap-2">
    <Button variant="ghost" size="sm" class="h-6 w-6 p-0">
        <Pause class="h-4 w-4" />
    </Button>
    <Button variant="ghost" size="sm" class="h-6 w-6 p-0">
        <RotateCcw class="h-4 w-4" />
    </Button>
    <CloseButton onCloseClick={onCloseClick} />
</div>
