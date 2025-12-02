<script lang="ts">
    import { Copy, Edit, Play, Save, Bot } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";

    let {
        message,
    }: {
        message: string;
    } = $props();
</script>

<div class="flex justify-start gap-2 items-start">
    <div
        class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
    >
        <Bot class="h-5 w-5" />
    </div>
    <div
        class="bg-muted text-foreground rounded-2xl rounded-tl-sm max-w-[80%] overflow-hidden"
    >
        <div
            class="flex justify-end gap-1 px-3 py-2 bg-muted/40 border-b border-border/50"
        >
            <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
                <Copy class="h-3.5 w-3.5" />
                <span class="text-xs">Kopieren</span>
            </Button>
            <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
                <Edit class="h-3.5 w-3.5" />
                <span class="text-xs">Bearbeiten</span>
            </Button>
            <Button
                variant="ghost"
                size="sm"
                class="h-7 gap-1.5 px-2"
                onclick={() => handleSaveTest(message.answer)}
            >
                <Save class="h-3.5 w-3.5" />
                <span class="text-xs">Speichern</span>
            </Button>
            <Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
                <Play class="h-3.5 w-3.5" />
                <span class="text-xs">Ausführen</span>
            </Button>
        </div>
        <div class="px-4 py-2 break-words whitespace-pre-wrap">
            {message}
        </div>
    </div>
</div>
