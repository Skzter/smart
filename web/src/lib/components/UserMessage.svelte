<script lang="ts">
    import { User } from "@lucide/svelte";

    let {
        message,
    }: {
        message: string;
    } = $props();
</script>

<!-- User message -->
<div class="flex justify-end gap-2 items-start">
    <div
        class="bg-primary text-primary-foreground rounded-2xl rounded-tr-sm px-4 py-2 max-w-[80%] wrap-break-word whitespace-pre-wrap"
    >
        {message}
    </div>
    <div
        class="h-8 w-8 shrink-0 rounded-full bg-primary flex items-center justify-center"
    >
        <User class="h-7 w-7 text-primary-foreground" />
    </div>
</div>
