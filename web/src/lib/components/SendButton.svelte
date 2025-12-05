<script lang="ts">
    import { Send } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { chat, messages, user } from "$lib/shared.svelte";
    import type { ApiChatRequest } from "$types/api";
    import { getChatResponse } from "$lib/api";

    let {
        input = $bindable(),
        onclick,
    }: {
        input: string;
        onclick: () => void;
    } = $props();
</script>

<Button
    variant="default"
    size="icon"
    class="w-10 h-10 cursor-pointer"
    disabled={chat.isLoading || !input.trim()}
    {onclick}
>
    <Send />
</Button>
