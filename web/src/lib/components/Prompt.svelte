<script lang="ts">
    import * as InputGroup from "$lib/components/ui/input-group";
    import { chat } from "$lib/shared.svelte";

    let {
        input = $bindable(),
        onclick,
    }: {
        input: string;
        onclick: () => void;
    } = $props();

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter" && input.trim() && !e.shiftKey) {
            onclick();
            input = "";
            e.preventDefault();
        }
    }
</script>

<InputGroup.Root class="w-full">
    <InputGroup.Textarea
        bind:value={input}
        class="w-full resize-none min-h-11"
        placeholder="Send a message..."
        rows={1}
        onkeydown={handleKeyPress}
        disabled={chat.isLoading}
    />
</InputGroup.Root>
