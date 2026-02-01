<script lang="ts">
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import SwitchView from "./SwitchView.svelte";
    import CloseButton from "./CloseButton.svelte";
    import BrowserView from "./BrowserView.svelte";
    import OutputView from "./OutputView.svelte";
    import TabsView from "./TabsView.svelte";
    import EditView from "./EditView.svelte";
    import ResultView from "./ResultView.svelte";
    import { Play } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button";
    import type { Runner } from "$lib/runner.svelte";

    let {
        code = $bindable(),
        activeTab = $bindable(),
        testRunner,
    }: {
        code: string;
        activeTab: string;
        testRunner: Runner;
    } = $props();

    let view = $state("split");

    function handleCloseClick() {
        const closeButton = document.querySelector(
            "[data-dialog-close]",
        ) as HTMLElement;
        closeButton?.click();
    }
</script>

<Dialog.Content
    class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px]
           h-[85vh] flex flex-col p-0 overflow-hidden"
    showCloseButton={false}
>
    <!-- REQUIRED HEADER STRUCTURE -->
    <div class="flex flex-row items-center justify-between border-b px-4 py-4">
        <Dialog.Title class="text-lg font-semibold">
            Button Click Test
        </Dialog.Title>

        <div class="flex items-center gap-2">
            {#if activeTab === "edit"}
                <CloseButton onCloseClick={handleCloseClick} />
            {:else if activeTab === "run"}
                <Button
                    variant="default"
                    size="sm"
                    onclick={() => testRunner.run()}
                    disabled={testRunner.isRunning() || testRunner.getCurTest() === ""}
                >
                    <Play class="h-3.5 w-3.5" />
                    {testRunner.isRunning() ? "Läuft..." : "Erneut ausführen"}
                </Button>
                <SwitchView onCloseClick={handleCloseClick} bind:view />
            {:else if activeTab === "result"}
                <CloseButton onCloseClick={handleCloseClick} />
            {/if}
        </div>
    </div>

    <Dialog.Close hidden data-dialog-close />

    <!-- Tabs -->
    <TabsView bind:activeTab curTest={testRunner.getCurTest()} />

    <!-- EDIT TAB -->
    {#if activeTab === "edit"}
        <div class="flex-1 overflow-visible">
            <EditView bind:activeTab bind:code {testRunner} />
        </div>

        <!-- RUN TAB -->
    {:else if activeTab === "run"}
        <div class="flex flex-col flex-1 overflow-hidden">
            {#if view === "split"}
                <div class="grid grid-cols-2 flex-1 overflow-hidden">
                    <!-- OUTPUT -->
                    <div class="flex flex-col overflow-hidden">
                        <div class="px-4 py-2 bg-muted/50">Test Output</div>
                        <OutputView {testRunner} />
                    </div>

                    <!-- BROWSER -->
                    <div class="flex flex-col overflow-hidden">
                        <BrowserView runner={testRunner} />
                    </div>
                </div>
            {:else if view === "code"}
                <div class="flex flex-col flex-1 overflow-hidden">
                    <div class="px-4 py-2 bg-muted/50">Test Output</div>
                    <OutputView {testRunner} />
                </div>
            {:else if view === "browser"}
                <div class="flex flex-col flex-1 overflow-hidden">
                    <BrowserView runner={testRunner} />
                </div>
            {/if}
        </div>

        <!-- RESULT TAB -->
    {:else if activeTab === "result"}
        <div class="flex-1 overflow-auto">
            <ResultView />
        </div>
    {/if}
</Dialog.Content>
