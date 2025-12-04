<script lang="ts">
    import { Play } from "@lucide/svelte";
    import { Button, buttonVariants } from "$lib/components/ui/button/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import * as Terminal from '$lib/components/ui/terminal';
</script>

<Dialog.Root>
  <Dialog.Trigger>
    <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
      <Play class="h-3.5 w-3.5" />
      <span class="text-xs">Ausführen</span>
    </Button>
  </Dialog.Trigger>
  <Dialog.Content class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px] h-[85vh] flex flex-col p-0">
    <Dialog.Header class="flex-row items-center justify-between border-b px-6 py-4 space-y-0">
      <Dialog.Title class="text-lg font-semibold">Button Click Test</Dialog.Title>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">⏱ 8.160s</span>
      </div>
    </Dialog.Header>
    
    <!-- Tabs -->
    <div class="flex border-b px-6">
      <button class="px-4 py-3 text-sm font-medium border-b-2 border-transparent hover:border-gray-300">
        Edit
      </button>
      <button class="px-4 py-3 text-sm font-medium border-b-2 border-black">
        Run
      </button>
      <button class="px-4 py-3 text-sm font-medium border-b-2 border-transparent hover:border-gray-300">
        Result
      </button>
    </div>

    <div class="flex-1 grid grid-cols-2 gap-0 overflow-hidden">
      <div class="flex flex-col border-r">
        <div class="px-4 py-2 border-b bg-muted/50 text-sm font-medium">
          Test Output
        </div>
        <div class="flex-1 overflow-auto">
          <Terminal.Root class="m-0 max-w-none h-full" delay={0}>
            <Terminal.TypingAnimation>&gt; Test wird ausgeführt...</Terminal.TypingAnimation>
          </Terminal.Root>
        </div>
      </div>

      <div class="flex flex-col">
        <div class="px-4 py-2 border-b bg-muted/50 text-sm font-medium flex items-center gap-2">
          <span>🖥️</span>
          Browser Vorschau
        </div>
        <div class="flex-1 overflow-auto bg-gray-50 flex items-center justify-center">
          <div class="text-center text-muted-foreground">
            <div class="text-6xl mb-4">🖥️</div>
            <p class="text-sm">Browser Vorschau</p>
          </div>
        </div>
      </div>
    </div>

    <div class="flex items-center justify-end gap-3 border-t px-6 py-3">
      <Button variant="default">Schließen</Button>
    </div>
  </Dialog.Content>
</Dialog.Root>