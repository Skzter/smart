<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import { Save } from "@lucide/svelte";
    import CloseButton from "./CloseButton.svelte";

    export let onSaveClick: () => void = () => {};
    export let onCloseClick: () => void = () => {};
</script>

<div class="flex items-center gap-2">
    <Button variant="default" size="sm" onclick={onSaveClick}>
        <Save class="h-3 w-4 mr-2" />
        Speichern
    </Button>
    <CloseButton onCloseClick={onCloseClick} />
</div>
