<script lang="ts">
    import RunWindow from "./RunWindow.svelte";
    import SaveButton from "./SaveButton.svelte";
    import EditButton from "./EditButton.svelte";
    import CopyButton from "./CopyButton.svelte";
    import * as Dialog from "./ui/dialog";
    import RunButton from "./RunButton.svelte";

    let {
        message = $bindable(),
        iscode,
    }: {
        message: string;
        iscode: boolean;
    } = $props();

    let activeTab = $state("run");
</script>

<div class="flex justify-end gap-1 px-2 py-2 border-b">
    <CopyButton bind:code={message} />
    {#if iscode}
        <Dialog.Root>
            <Dialog.Trigger>
                <RunButton
                    classes="h-7 gap-1.5 px-2 cursor-pointer"
                    variant="outline"
                    size="sm"
                    bind:activeTab
                />
            </Dialog.Trigger>
            <Dialog.Trigger>
                <EditButton />
            </Dialog.Trigger>
            <SaveButton
                classes="h-7 gap-1.5 px-2 cursor-pointer"
                variant="outline"
                size="sm"
                bind:code={message}
            />
            <RunWindow bind:code={message} bind:activeTab />
        </Dialog.Root>
    {/if}
</div>
