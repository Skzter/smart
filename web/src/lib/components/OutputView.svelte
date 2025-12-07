<script lang="ts">
    import * as Terminal from "$lib/components/ui/terminal";
    import { Runner } from "$lib/runner.svelte";

    let { testRunner = $bindable() }: { testRunner: Runner } = $props();
</script>

<div class="flex flex-col">
    <div class="px-4 py-2 bg-muted/50 text-sm font-medium">
        Test Output {testRunner.result}
    </div>
    <div class="flex-1 overflow-auto">
        <Terminal.Root
            class="m-0 max-w-none h-full rounded-none border-none"
            delay={0}
            hideDots={true}
        >
            <Terminal.TypingAnimation
                >&gt; Test wird ausgeführt...</Terminal.TypingAnimation
            >
        </Terminal.Root>
    </div>
</div>
