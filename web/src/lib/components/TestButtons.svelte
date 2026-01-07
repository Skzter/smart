<script lang="ts">
    import RunWindow from "./RunWindow.svelte";
    import SaveButton from "./SaveButton.svelte";
    import EditButton from "./EditButton.svelte";
    import CopyButton from "./CopyButton.svelte";
    import * as Dialog from "./ui/dialog";
    import RunButton from "./RunButton.svelte";
    import { Runner } from "$lib/runner.svelte";

    let {
        message = $bindable(),
        testRunner,
        iscode,
    }: {
        message: string;
        testRunner: Runner;
        iscode: boolean;
    } = $props();

    let activeTab = $state("run");
</script>

<div class="flex justify-end gap-1 px-2 py-2 border-b">
    <CopyButton bind:code={message} />
    {#if iscode}
        <Dialog.Root>
            {#if testRunner.getCurTest() !== ""}
                <Dialog.Trigger>
                    <RunButton
                        classes="h-7 gap-1.5 px-2 cursor-pointer"
                        variant="outline"
                        size="sm"
                        bind:activeTab
                        {testRunner}
                    />
                </Dialog.Trigger>
            {:else}
                <RunButton
                    classes="h-7 gap-1.5 px-2 cursor-pointer"
                    variant="outline"
                    size="sm"
                    bind:activeTab
                    {testRunner}
                />
            {/if}
            <Dialog.Trigger>
                <EditButton bind:activeTab />
            </Dialog.Trigger>
            <SaveButton
                classes="h-7 gap-1.5 px-2 cursor-pointer"
                variant="outline"
                size="sm"
                bind:code={message}
                {testRunner}
            />
            <RunWindow bind:code={message} bind:activeTab {testRunner} />
        </Dialog.Root>
    {/if}
</div>
