<script lang="ts">
    import { Play } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import SwitchView from "./SwitchView.svelte";
</script>

<Dialog.Root>
    <Dialog.Trigger>
        <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
            <Play class="h-3.5 w-3.5" />
            <span class="text-xs">Ausführen</span>
        </Button>
    </Dialog.Trigger>
    <Dialog.Content
        class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px] h-[85vh] flex flex-col p-0"
    >
        <div
            class="flex flex-row items-center justify-between border-b px-6 py-4"
        >
            <Dialog.Title class="text-lg font-semibold"
                >Button Click Test</Dialog.Title
            >
            <div class="flex items-center gap-2">
                <SwitchView />
                <span class="text-sm text-muted-foreground">⏱ 8.160s</span>
            </div>
        </div>

        <!-- Content Platzhalter -->
        <div class="flex-1 p-6 text-center text-muted-foreground">
            Dialog Inhalt hier
        </div>
    </Dialog.Content>
</Dialog.Root>
