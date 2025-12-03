<script lang="ts">
    import { Send } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { chat } from "$lib/shared.svelte";

    let {
        input = $bindable(),
    }: {
        input: string;
    } = $props();

    // onclick mit user id
</script>

<Button
    variant="default"
    size="icon"
    class="w-10 h-10"
    disabled={chat.isLoading}
>
    <Send />
</Button>
