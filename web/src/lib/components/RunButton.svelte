<script lang="ts">
    import { Play } from "@lucide/svelte";
    import {
        Button,
        type ButtonSize,
        type ButtonVariant,
    } from "$lib/components/ui/button/index.js";
    import type { Runner } from "$lib/runner.svelte";

    let {
        activeTab = $bindable(),
        testRunner,
        classes,
        variant,
        size,
    }: {
        activeTab: string;
        testRunner: Runner;
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
        activeTab = "run";
        testRunner.run();
    }}
    disabled={testRunner.isRunning() || testRunner.getCurTest() === ""}
>
    <Play class="h-3.5 w-3.5" />
    <p>{testRunner.isRunning() ? "Lädt..." : "Ausführen"}</p>
</Button>
