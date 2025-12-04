<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
  import { Code, CodeXml, Columns2, Eye, Monitor } from "@lucide/svelte";
  import { Separator } from "$lib/components/ui/separator/index.js";
  import ControlButtons from './ControlButtons.svelte';

  export let onMonitorClick = () => {};
  export let onCodeClick = () => {};
  export let onSplitClick = () => {};
  export let onCloseClick = () => {};
  export let activeView = 'split';
</script>
 
<div class="flex items-center gap-2">
  <ButtonGroup.Root>
    <Button variant={activeView === 'split' ? 'default' : 'secondary'} size="sm" onclick={onSplitClick}>
      <Columns2 class="h-4 w-4" />
    </Button>
    <ButtonGroup.Separator />
    <Button variant={activeView === 'code' ? 'default' : 'secondary'} size="sm" onclick={onCodeClick}>
      <CodeXml class="h-4 w-4" />
    </Button>
    <ButtonGroup.Separator />
    <Button variant={activeView === 'fullscreen' ? 'default' : 'secondary'} size="sm" onclick={onMonitorClick}>
      <Monitor class="h-4 w-4" />
    </Button>
  </ButtonGroup.Root>

  <Separator orientation="vertical" class="h-6" />
  
  <ControlButtons {onCloseClick} />
</div>
