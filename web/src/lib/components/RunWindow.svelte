<script lang="ts">
    import { Play } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { runContainer } from "$lib/api";
    import SwitchView from "./SwitchView.svelte";
    import BrowserView from "./BrowserView.svelte";
    import OutputView from "./OutputView.svelte";
    import TabsView from "./TabsView.svelte";
    import EditView from "./EditView.svelte";
    import ResultView from "./ResultView.svelte";
    import SaveButtons from "./SaveButtons.svelte";

    let {
        code,
    }: {
        code: string;
    } = $props();

    // Mock-Werte für API-Call (diese sollten von außen kommen)
    const userId = "687270280dca20b77cfdcf73";
    const testId = "b6f75688-c6ba-43c9-9d3b-f80132755def";
    const sessionId = "94b7e18c-e2e4-4d17-ac21-40c5c8b86162";
    let testResult = $state("");
    let isLoading = $state(false);

    let isFullscreenBrowser = $state(false);
    let isFullscreenCode = $state(false);
    let activeTab = $state("run");

    async function handleRunFromButton() {
        isLoading = true;
        try {
            const response = await runContainer({
                userId,
                testId,
                sessionId,
            });
            testResult = response.result;
            activeTab = "result";
        } catch (error) {
            console.error("Error running test:", error);
        } finally {
            isLoading = false;
        }
    }

    function handleTabChange(event: CustomEvent) {
        activeTab = event.detail;
    }

    function handleRunTest(result: string) {
        testResult = result;
        activeTab = "result";
    }

    function handleSaveClick() {
        // Speichern-Logik hier
        console.log("Speichern geklickt");
    }

    function handleCloseClick() {
        const closeButton = document.querySelector(
            "[data-dialog-close]",
        ) as HTMLElement;
        closeButton?.click();
    }

    function handleSplitClick() {
        isFullscreenBrowser = false;
        isFullscreenCode = false;
    }

    function handleMonitorClick() {
        if (isFullscreenBrowser) {
            isFullscreenBrowser = false;
        } else {
            isFullscreenBrowser = true;
            isFullscreenCode = false;
        }
    }

    function handleCodeClick() {
        if (isFullscreenCode) {
            isFullscreenCode = false;
        } else {
            isFullscreenCode = true;
            isFullscreenBrowser = false;
        }
    }
</script>

<Dialog.Root>
    <Dialog.Trigger>
        <Button
            variant="ghost"
            size="sm"
            class="h-7 gap-1.5 px-2"
            onclick={handleRunFromButton}
            disabled={isLoading}
        >
            <Play class="h-3.5 w-3.5" />
            <span class="text-xs">{isLoading ? "Lädt..." : "Ausführen"}</span>
        </Button>
    </Dialog.Trigger>
    <Dialog.Content
        class="sm:max-w-[90vw] md:max-w-[80vw] lg:max-w-[1170px] h-[85vh] flex flex-col p-0"
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
                    <SaveButtons
                        onSaveClick={handleSaveClick}
                        onCloseClick={handleCloseClick}
                    />
                {:else if activeTab === "run"}
                    <SwitchView
                        onSplitClick={handleSplitClick}
                        onMonitorClick={handleMonitorClick}
                        onCodeClick={handleCodeClick}
                        onCloseClick={handleCloseClick}
                        activeView={isFullscreenBrowser
                            ? "fullscreen"
                            : isFullscreenCode
                              ? "code"
                              : "split"}
                    />
                {/if}
            </div>
        </div>
        <Dialog.Close hidden data-dialog-close />

        <TabsView bind:activeTab on:tabChange={handleTabChange} />

        {#if activeTab === "edit"}
            <div class="flex-1 overflow-hidden">
                <EditView
                    {code}
                    {userId}
                    {testId}
                    {sessionId}
                    onRunClick={handleRunTest}
                />
            </div>
        {:else if activeTab === "run"}
            <div
                class="flex-1 {isFullscreenBrowser || isFullscreenCode
                    ? 'grid grid-cols-1'
                    : 'grid grid-cols-2'} gap-0 overflow-hidden"
            >
                {#if !isFullscreenBrowser && !isFullscreenCode}
                    <OutputView result={testResult} />
                {:else if isFullscreenCode}
                    <OutputView result={testResult} />
                {/if}
                {#if !isFullscreenCode}
                    <BrowserView />
                {/if}
            </div>
        {:else if activeTab === "result"}
            <div class="flex-1 overflow-hidden">
                <ResultView />
            </div>
        {/if}
    </Dialog.Content>
</Dialog.Root>
