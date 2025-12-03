<script lang="ts">
    import { Send } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { chat } from "$lib/shared.svelte";

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
    class="w-10 h-10"
    disabled={chat.isLoading}
    {onclick}
>
    <Send />
</Button>
