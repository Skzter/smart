<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import { Save, X } from "@lucide/svelte";

  export let onSaveClick = () => {};
  export let onCloseClick = () => {};
</script>

<div class="flex items-center gap-2">
  <Button variant="default" size="sm" onclick={onSaveClick}>
    <Save class="h-3 w-4 mr-2" />
    Speichern
  </Button>
  <Button variant="outline" size="sm" onclick={onCloseClick}>
    <X class="h-4 w-4" />
  </Button>
</div>
