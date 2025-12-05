<script lang="ts">
    import { getTemplate } from "$lib/api";
    import * as InputGroup from "$lib/components/ui/input-group";
    import { chat } from "$lib/shared.svelte";
    import { onMount } from "svelte";
    import { toast } from "svelte-sonner";

    let {
        input = $bindable(),
        onclick,
    }: {
        input: string;
        onclick: () => void;
    } = $props();

    onMount(async () => {
        try {
            input = await getTemplate();
            input = `
Erzeuge Playwright-Tests via Autoplaywright für meine lokale Seite.
    Base-URL: localhost:8082
    Szenario: Nutzer nutzt die Suche.
    Ablauf: Zuerst gibt der Nutzer beim Reiseziel „Mallorca“ ein.
    Assertions: Im Reiseziel-Feld steht „Mallorca“.
    Testdaten/Setup: Reiseziel Mallorca
`;
        } catch (err: unknown) {
            toast.error((err as Error).message, {
                description: "Das war wohl nichts mit dem Template.",
            });
        }
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
