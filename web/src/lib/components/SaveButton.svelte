<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button, {
        type ButtonSize,
        type ButtonVariant,
    } from "./ui/button/button.svelte";
    import type { SaveState } from "$types/save";
    import type { Runner } from "$lib/runner.svelte";

    let {
        code = $bindable(),
        testRunner = $bindable(),
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

    let saveState = $state<SaveState>("idle");

    $effect(() => {
        console.log(
            "SaveState: " + saveState + " TestID: " + testRunner.getCurTest(),
        );
    });
</script>

<Button
    {variant}
    {size}
    class={classes}
    onclick={() => testRunner.storeTest(code)}
    disabled={saveState === "saving"}
>
    <Save class="h-3.5 w-3.5" />
    <p>Speichern</p>
</Button>
