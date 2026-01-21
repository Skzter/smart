<script lang="ts">
    import { getTemplate } from "$lib/api";
    import * as InputGroup from "$lib/components/ui/input-group";
    import { chat, user } from "$lib/shared.svelte";
    import { toast } from "svelte-sonner";

    let {
        input = $bindable(),
        onclick,
    }: {
        input: string;
        onclick: () => void;
    } = $props();

    $effect(() => {
        if (!user.id) return;
        (async () => {
            try {
                input = await getTemplate();
            } catch (err: unknown) {
                toast.error((err as Error).message, {
                    description: "Das war wohl nichts mit dem Template.",
                });
            }
        })();
    });

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
