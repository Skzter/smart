<script lang="ts">
    import * as Dialog from "../../src/lib/components/ui/dialog/index.js";
    import RunWindow from "../../src/lib/components/RunWindow.svelte";
    import type { Runner } from "../../src/lib/runner.svelte";

    let {
        code = $bindable(),
        activeTab = $bindable(),
        testRunner,
    }: { code: string; activeTab: string; testRunner: Runner } = $props();
</script>

<Dialog.Root open={true}>
    <RunWindow bind:code bind:activeTab {testRunner} />
</Dialog.Root>
