<script lang="ts">
    import { Plus } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";

    function handleNewChat() {
        console.log("New chat started");
    }
</script>

<Button variant="ghost" class="w-full justify-start gap-2 h-10 bg-muted hover:bg-muted/80" onclick={handleNewChat}>
    <Plus class="h-4 w-4" />
    <span>New Chat</span>
</Button>