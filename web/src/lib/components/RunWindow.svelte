<script lang="ts">
    import { Play, X } from "@lucide/svelte";
    import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import SwitchView from './SwitchView.svelte';
    import BrowserView from './BrowserView.svelte';
    import OutputView from './OutputView.svelte';
    import TabsView from './TabsView.svelte';
    import EditView from './EditView.svelte';
    import ResultView from './ResultView.svelte';
    import SaveButtons from './SaveButtons.svelte';

    let {
        code,
    }: {
        code: string;
    } = $props();

    let isFullscreenBrowser = $state(false);
    let isFullscreenCode = $state(false);
    let activeTab = $state('run');

    function handleTabChange(event: CustomEvent) {
      activeTab = event.detail;
    }

    function handleSaveClick() {
      // Speichern-Logik hier
      console.log('Speichern geklickt');
    }

    function handleCloseClick() {
      const closeButton = document.querySelector('[data-dialog-close]') as HTMLElement;
      closeButton?.click();
    }

    function handleSplitClick() {
      isFullscreenBrowser = false;
      isFullscreenCode = false;
    }

    function handleMonitorClick() {
      if (isFullscreenBrowser) {
        isFullscreenBrowser = false;
      } else {
        isFullscreenBrowser = true;
        isFullscreenCode = false;
      }
    }

    function handleCodeClick() {
      if (isFullscreenCode) {
        isFullscreenCode = false;
      } else {
        isFullscreenCode = true;
        isFullscreenBrowser = false;
      }
    }
</script>

<Dialog.Root>
  <Dialog.Trigger>
    <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
      <Play class="h-3.5 w-3.5" />
      <span class="text-xs">Ausführen</span>
    </Button>
  </Dialog.Trigger>
  <Dialog.Content class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px] h-[85vh] flex flex-col p-0" showCloseButton={false}>
    <div class="flex flex-row items-center justify-between border-b px-4 py-4">
      <Dialog.Title class="text-lg font-semibold">Button Click Test</Dialog.Title>
      <div class="flex items-center gap-2">
        {#if activeTab === 'edit'}
          <SaveButtons onSaveClick={handleSaveClick} onCloseClick={handleCloseClick} />
        {:else if activeTab === 'run'}
          <SwitchView onSplitClick={handleSplitClick} onMonitorClick={handleMonitorClick} onCodeClick={handleCodeClick} onCloseClick={handleCloseClick} activeView={isFullscreenBrowser ? 'fullscreen' : isFullscreenCode ? 'code' : 'split'} />
        {/if}
      </div>
    </div>
    <Dialog.Close hidden data-dialog-close />
    
    <TabsView bind:activeTab on:tabChange={handleTabChange} />

    {#if activeTab === 'edit'}
      <div class="flex-1 overflow-hidden">
        <EditView code={code} />
      </div>
    {:else if activeTab === 'run'}
      <div class="flex-1 {isFullscreenBrowser || isFullscreenCode ? 'grid grid-cols-1' : 'grid grid-cols-2'} gap-0 overflow-hidden">
        {#if !isFullscreenBrowser && !isFullscreenCode}
          <OutputView />
        {:else if isFullscreenCode}
          <OutputView />
        {/if}
        {#if !isFullscreenCode}
          <BrowserView />
        {/if}
      </div>
    {:else if activeTab === 'result'}
      <div class="flex-1 overflow-hidden">
        <ResultView />
      </div>
    {/if}
  </Dialog.Content>
</Dialog.Root>