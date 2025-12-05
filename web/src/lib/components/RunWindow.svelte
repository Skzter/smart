<script lang="ts">
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import SwitchView from "./SwitchView.svelte";
    import CloseButton from "./CloseButton.svelte";
    import BrowserView from "./BrowserView.svelte";
    import OutputView from "./OutputView.svelte";
    import TabsView from "./TabsView.svelte";
    import EditView from "./EditView.svelte";
    import ResultView from "./ResultView.svelte";
    import RunButton from "./RunButton.svelte";

    let {
        code = $bindable(),
    }: {
        code: string;
    } = $props();

    let testResult = $state("");
    let isLoading = $state(false);

    let view = $state("split");
    let activeTab = $state("run");

    function handleTabChange(event: CustomEvent) {
        activeTab = event.detail;
    }

    function handleCloseClick() {
        const closeButton = document.querySelector(
            "[data-dialog-close]",
        ) as HTMLElement;
        closeButton?.click();
    }
</script>

<Dialog.Root>
    <Dialog.Trigger>
        <RunButton
            classes="h-7 gap-1.5 px-2 cursor-pointer"
            variant="outline"
            size="sm"
            bind:isLoading
            bind:testResult
            bind:activeTab
        />
    </Dialog.Trigger>
    <Dialog.Content
        class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px] h-[85vh] flex flex-col p-0 overflow-hidden"
        showCloseButton={false}
    >
        <div
            class="flex flex-row items-center justify-between border-b px-4 py-4"
        >
            <Dialog.Title class="text-lg font-semibold"
                >Button Click Test</Dialog.Title
            >
            <div class="flex items-center gap-2">
                {#if activeTab === "edit"}
                    <CloseButton onCloseClick={handleCloseClick} />
                {:else if activeTab === "run"}
                    <SwitchView onCloseClick={handleCloseClick} bind:view />
                {:else if activeTab === "result"}
                    <CloseButton onCloseClick={handleCloseClick} />
                {/if}
            </div>
        </div>
        <Dialog.Close hidden data-dialog-close />

        <TabsView bind:activeTab on:tabChange={handleTabChange} />

        {#if activeTab === "edit"}
            <div class="flex-1 overflow-hidden">
                <EditView
                    bind:activeTab
                    bind:code
                    bind:isLoading
                    bind:testResult
                />
            </div>
        {:else if activeTab === "run"}
            <div
                class="flex-1 {view === 'split'
                    ? 'grid grid-cols-2'
                    : 'grid grid-cols-1'} gap-0 overflow-hidden"
            >
                {#if view == "split" || view == "code"}
                    <OutputView result={testResult} />
                {/if}
                {#if view == "split" || view == "browser"}
                    <BrowserView />
                {/if}
            </div>
        {:else if activeTab === "result"}
            <div class="flex-1 overflow-auto">
                <ResultView />
            </div>
        {/if}
    </Dialog.Content>
</Dialog.Root>
