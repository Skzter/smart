<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    function handleSaveTest(msg: string) {
        console.log(msg);
        console.log(userId, conversationId);
    }

    let {
        code,
        userId,
        conversationId,
    }: {
        code: string;
        userId: string;
        conversationId: string;
    } = $props();
</script>

<Button
    variant="ghost"
    size="sm"
    class="h-7 gap-1.5 px-2"
    onclick={() => handleSaveTest(code)}
>
    <Save class="h-3.5 w-3.5" />
    <span class="text-xs">Speichern</span>
</Button>
