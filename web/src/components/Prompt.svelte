<script lang="ts">
    import {
        Spinner,
        Button,
        Modal,
        type ModalProps,
        type ModalPlacementType,
    } from "flowbite-svelte";
    import { getTemplate, runContainer } from "../lib/Api.ts";
    import Popup from "./Popup.svelte";
    let { input = $bindable(""), onclick } = $props();

    let showPopup = $state(false);
    let popupMessage = $state("");
    let popupTitle = $state("Error");

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter" && input.trim() && !e.shiftKey) {
            onclick();
            input = "";
            e.preventDefault();
        }
    }

    async function loadTemplate() {
        try {
            let response = await getTemplate("/template");
            input = response.data.template;
        } catch {
            popupMessage = "NO TEMPLATE FOUND!";
            popupTitle = "Template Error";
            showPopup = true;
        }
    }

    let flowModal = $state(false);
    let flowModalMSG = $state("");
    let flowRunning = $state(false);
    let size: ModalProps["size"] = $state("xl"); // Set default value
    let placement: ModalPlacementType = $state("top-center");
    async function runTest() {
        try {
            flowModal = true; // Show the modal instead of popup
            flowRunning = true;
            let response = await runContainer("/run");
            flowRunning = false;
            flowModalMSG = response.data;
        } catch (err) {
            flowRunning = false;
            flowModalMSG = err.response.data;
        }
    }
</script>

<div class="flex flex-row w-screen items-center bg-white border-t gap-2 p-4">
    <textarea
        onkeydown={handleKeyPress}
        bind:value={input}
        placeholder="Prompt"
        required
        class="w-9/10 resize-none overflow-y-auto"
        rows={3}
    ></textarea>
    <Button
        color="purple"
        class="w-1/10 h-1/3"
        {onclick}
        disabled={!input.trim()}>Send</Button
    >
    <Button color="purple" class="w-1/10 h-1/3" onclick={loadTemplate}
        >Template</Button
    >
    <Button color="purple" class="w-1/10 h-1/3" onclick={runTest}>Run</Button>
</div>

<Popup bind:isOpen={showPopup} message={popupMessage} title={popupTitle} />

<Modal title="Test Result" form bind:open={flowModal} {size} {placement}>
    <div class="space-y-4">
        {#if flowRunning}
            <Spinner type="dots" color="purple" />
        {/if}
        <pre
            class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
      {flowModalMSG}
    </pre>
    </div>
</Modal>
