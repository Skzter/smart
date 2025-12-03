<script lang="ts">
    import * as InputGroup from "$lib/components/ui/input-group";
    import { chat } from "$lib/shared.svelte";

    let {
        input = $bindable(),
    }: {
        input: string;
    } = $props();
</script>

<InputGroup.Root class="w-full">
    <InputGroup.Textarea
        bind:value={input}
        class="w-full resize-none min-h-11"
        placeholder="Send a message..."
        rows={1}
        disabled={chat.isLoading}
    />
</InputGroup.Root>
