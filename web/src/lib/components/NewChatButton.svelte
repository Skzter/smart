<script lang="ts">
    import { Plus } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { chat, messages } from "$lib/shared.svelte";

    function NewChat() {
        chat.id = "";
        chat.isLoading = false;
        chat.groups = [];
        // empty the array
        messages.length = 0;
    }
</script>

<Button
    variant="ghost"
    class="w-full justify-start gap-2 h-10 bg-muted hover:bg-muted/80"
    onclick={NewChat}
>
    <Plus class="h-4 w-4" />
    <p>New Chat</p>
</Button>
