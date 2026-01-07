<script lang="ts">
    import * as Terminal from "$lib/components/ui/terminal";
    import { Runner } from "$lib/runner.svelte";

    let { testRunner }: { testRunner: Runner } = $props();
    let res = $derived(testRunner.result);
</script>

<div class="flex flex-col">
    <div class="px-4 py-2 bg-muted/50 text-sm font-medium">Test Output</div>
    <div class="flex-1 overflow-auto">
        <Terminal.Root
            class="m-0 max-w-none h-full rounded-none border-none"
            delay={0}
            hideDots={true}
            speed={5}
        >
            <Terminal.TypingAnimation
                >&gt; Test wird ausgeführt
            </Terminal.TypingAnimation>
            {#each res as step}
                <Terminal.AnimatedSpan
                    >{step.end ?? step.begin}</Terminal.AnimatedSpan
                >
            {/each}
        </Terminal.Root>
    </div>
</div>
