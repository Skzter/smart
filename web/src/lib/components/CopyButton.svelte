<script lang="ts">
    import { Copy, Check } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";

    let { code }: { code: string } = $props();

    let copied = $state(false);

    async function handleCopy() {
        try {
            await navigator.clipboard.writeText(code);
            copied = true;
            setTimeout(() => {
                copied = false;
            }, 1000);
        } catch (err) {
            console.error("Failed to copy:", err);
        }
    }
</script>

<Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2 cursor-pointer" onclick={handleCopy}>
    {#if copied}
        <Check class="h-3.5 w-3.5" />
    {:else}
        <Copy class="h-3.5 w-3.5" />
    {/if}
    <span class="text-xs">Kopieren</span>
</Button>
