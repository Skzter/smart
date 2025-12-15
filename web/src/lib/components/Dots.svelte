<script>
    import { Bot } from "@lucide/svelte";
</script>

<div class="flex justify-start gap-2 items-start">
    <div
        class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
    >
        <Bot class="h-7 w-7" />
    </div>
    <div class="bg-muted text-foreground rounded-2xl rounded-bl-sm px-4 py-2">
        <div class="flex gap-1">
            <span class="animate-bounce" style="animation-delay: 0ms;">●</span>
            <span class="animate-bounce" style="animation-delay: 150ms;">●</span
            >
            <span class="animate-bounce" style="animation-delay: 300ms;">●</span
            >
        </div>
    </div>
</div>
