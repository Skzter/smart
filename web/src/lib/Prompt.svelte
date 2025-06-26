<script lang="ts">
    import { Button } from "flowbite-svelte";
    let { input = $bindable(""), onclick } = $props();

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter" && input.trim() && !e.shiftKey) {
            onclick();
            input = "";
            e.preventDefault();
        }
    }
</script>

<div
    class="fixed bottom-0 left-0 w-full flex flex-row w-screen items-center bg-white p-4 border-t gap-2"
    on:keydown={handleKeyPress}
>
    <textarea
        bind:value={input}
        placeholder="Prompt"
        required
        class="w-9/10 resize-none overflow-y-auto"
        rows={3}
    />
    <Button
        color="purple"
        class="w-1/10 h-1/3"
        {onclick}
        disabled={!input.trim()}>Send</Button
    >
</div>
