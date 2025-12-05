<script lang="ts">
    import { Play } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button/index.js";
    import * as Dialog from "$lib/components/ui/dialog/index.js";
    import { runContainer } from "$lib/api";
    import SwitchView from "./SwitchView.svelte";
    import CloseButton from "./CloseButton.svelte";
    import ControlButtons from "./ControlButtons.svelte";
    import BrowserView from "./BrowserView.svelte";
    import OutputView from "./OutputView.svelte";
    import TabsView from "./TabsView.svelte";
    import EditView from "./EditView.svelte";
    import ResultView from "./ResultView.svelte";
    import SaveButtons from "./SaveButtons.svelte";
    import { user, chat } from "$lib/shared.svelte";

    let {
        code,
    }: {
        code: string;
    } = $props();

    let testResult = $state("");
    let isLoading = $state(false);

    let view = $state("split");
    let activeTab = $state("run");

    let testId = "test";

    async function handleRunFromButton() {
        isLoading = true;
        try {
            const response = await runContainer({
                userId: user.id,
                testId,
                sessionId: chat.id,
            });
            testResult = response;
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
</script>

<Dialog.Root>
    <Dialog.Trigger>
        <Button
            variant="ghost"
            size="sm"
            class="h-7 gap-1.5 px-2 cursor-pointer"
            onclick={handleRunFromButton}
            disabled={isLoading}
        >
            <Play class="h-3.5 w-3.5" />
            <span class="text-xs">{isLoading ? "Lädt..." : "Ausführen"}</span>
        </Button>
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
                    <SaveButtons
                        onSaveClick={handleSaveClick}
                        onCloseClick={handleCloseClick}
                    />
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
                <EditView {code} {testId} onRunClick={handleRunTest} />
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
