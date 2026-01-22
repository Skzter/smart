<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button, {
        type ButtonSize,
        type ButtonVariant,
    } from "./ui/button/button.svelte";
    import type { Runner } from "$lib/runner.svelte";
    import Spinner from "./ui/spinner/spinner.svelte";

    let {
        code = $bindable(),
        testRunner,
        classes,
        variant,
        size,
    }: {
        code: string;
        testRunner: Runner;
        classes: string;
        variant: ButtonVariant;
        size: ButtonSize;
    } = $props();
</script>

<Button
    {variant}
    {size}
    class={classes}
    onclick={() => testRunner.storeTest(code)}
    disabled={testRunner.getStorageState() === "saving"}
>
    {#if testRunner.getStorageState() === "saving"}
        <Spinner></Spinner>
    {:else}
        <Save class="h-3.5 w-3.5" />
    {/if}
    <p>Speichern</p>
</Button>
