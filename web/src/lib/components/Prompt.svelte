<script lang="ts">
    import * as InputGroup from "$lib/components/ui/input-group";

    let {
        input = $bindable(),
        isLoading = $bindable(),
    }: {
        input: string;
        isLoading: boolean;
    } = $props();
</script>

<InputGroup.Root class="w-full">
    <InputGroup.Textarea
        bind:value={input}
        class="w-full resize-none min-h-11"
        placeholder="Send a message..."
        rows={1}
        disabled={isLoading}
    />
</InputGroup.Root>
