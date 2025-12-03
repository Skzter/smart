<script lang="ts">
    import { Send } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";

    let {
        userId,
        input = $bindable(),
        isLoading = $bindable(),
    }: {
        userId: string;
        input: string;
        isLoading: boolean;
    } = $props();

    // onclick mit user id
</script>

<Button variant="default" size="icon" class="w-10 h-10" disabled={isLoading}>
    <Send />
</Button>
