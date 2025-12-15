<script lang="ts">
    import { Monitor } from "@lucide/svelte";
</script>

<div class="flex flex-col">
    <div
        class="px-4 py-2 border-b bg-muted/50 text-sm font-medium flex items-center gap-2"
    >
        Vorschau
    </div>
    <div
        class="flex-1 overflow-auto bg-gray-50 flex items-center justify-center"
    >
        <div class="text-center text-muted-foreground">
            <div class="text-10xl mb-4"><Monitor /></div>
            <p class="text-sm">Browser Vorschau</p>
        </div>
    </div>
</div>
