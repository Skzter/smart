<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import { Pause, RotateCcw, X } from "@lucide/svelte";

  export let onCloseClick = () => {};
</script>

<div class="flex items-center gap-2">
  <Button variant="ghost" size="sm" class="h-6 w-6 p-0">
    <Pause class="h-4 w-4" />
  </Button>
  <Button variant="ghost" size="sm" class="h-6 w-6 p-0">
    <RotateCcw class="h-4 w-4" />
  </Button>
  <Button variant="ghost" size="sm" class="h-6 w-6 p-0" onclick={onCloseClick}>
    <X class="h-4 w-4" />
  </Button>
</div>
