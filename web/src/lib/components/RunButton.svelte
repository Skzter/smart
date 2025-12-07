<script lang="ts">
    import { Play } from "@lucide/svelte";
    import {
        Button,
        type ButtonSize,
        type ButtonVariant,
    } from "$lib/components/ui/button/index.js";
    import { runner } from "$lib/shared.svelte";

    let {
        activeTab = $bindable(),
        classes,
        variant,
        size,
    }: {
        activeTab: string;

        classes: string;
        variant: ButtonVariant;
        size: ButtonSize;
    } = $props();
</script>

<Button
    class={classes}
    {variant}
    {size}
    onclick={() => {
        runner.run();
    }}
>
    <Play class="h-3.5 w-3.5" />
    <p>{runner.isRunning() ? "Lädt..." : "Ausführen"}</p>
</Button>
