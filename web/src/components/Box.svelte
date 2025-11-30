<script lang="ts">
    import { Button, Spinner, Modal, type ModalProps } from "flowbite-svelte";
    import { saveTestLocal, runContainer } from "../lib/Api.ts";
    import { AxiosError } from "axios";
    import Code from "./Code.svelte";

    let { msg, name, userId, conversationId, isCode = false } = $props();
    let saveState = $state<"idle" | "saving" | "success" | "error">("idle");
    let errorMessage = $state("");
    let testId = $state<string | undefined>(undefined);

    async function saveTest(testcode: string) {
        if (!userId || !conversationId) {
            errorMessage =
                "Failed to save test: userId or conversationId is missing";
            console.error(errorMessage, { userId, conversationId });
            return;
        }

        const sanitizedUserId = userId.includes("|")
            ? userId.split("|")[1]
            : userId;

        const request = {
            userId: sanitizedUserId,
            conversationId: conversationId,
            code: testcode,
        };

        try {
            saveState = "saving";
            const response = await saveTestLocal(request);
            testId = response.data.testcaseId;
            console.log("Test saved successfully:", response.data);
            saveState = "success";
        } catch (error: unknown) {
            console.error("Failed to save test:", error);
            saveState = "error";

            if (error instanceof AxiosError) {
                errorMessage =
                    error.response?.data?.error ||
                    "Failed to save test. Please try again.";
            } else {
                errorMessage = "Failed to save test. Please try again.";
            }

            setTimeout(() => {
                saveState = "idle";
            }, 2000);
        }
    }

    let logShowModal = $state(false);
    let logModalMSG = $state("");
    let logRunning = $state(false);
    let size: ModalProps["size"] = $state("xl");
    type ModalPlacementType =
        | "top-left"
        | "top-center"
        | "top-right"
        | "center-left"
        | "center"
        | "center-right"
        | "bottom-left"
        | "bottom-center"
        | "bottom-right";
    let placement: ModalPlacementType = $state("top-center");
    async function RunTest() {
        if (!userId || !conversationId || !testId) {
            errorMessage =
                "Failed to run test: userId, conversationId or testid is missing";
            console.error(errorMessage, { userId, conversationId });
            return;
        }
        const sanitizedUserId = userId.includes("|")
            ? userId.split("|")[1]
            : userId;
        const request = {
            userId: sanitizedUserId,
            testId: testId,
            sessionId: conversationId,
        };

        console.log(request);
        try {
            logModalMSG = "";
            logShowModal = true; // Show the modal instead of popup
            logRunning = true;
            const response = await runContainer(request);
            console.log(response);
            logRunning = false;
            logModalMSG = response.data.result;
        } catch (err) {
            console.log(err);
            logRunning = false;
            logModalMSG = err.response.data.message;
        }
    }
</script>

<div
    class="flex m-4"
    class:justify-end={name === "User"}
    class:justify-start={name === "Bot"}
>
    <div
        class="font-mono p-2.5 border-2 border-black border-solid rounded-xl"
        class:w-[75%]={name === "User"}
        class:w-fit={name === "Bot"}
        class:bg-sky-300={name === "User"}
        class:bg-gray-200={name === "Bot" && !isCode}
        class:bg-(--code-background)={name === "Bot" && isCode}
    >
        <div class="flex items-start justify-between">
            <h1
                class="tracking-wide uppercase font-bold text-xl"
                class:text-(--heading)={name === "Bot" && isCode}
            >
                {name}
            </h1>
            <div class="flex">
                <div class="ml-2">
                    {#if isCode}
                        <Button
                            color={saveState === "success"
                                ? "green"
                                : saveState === "error"
                                  ? "red"
                                  : "purple"}
                            disabled={saveState === "saving" ||
                                saveState === "success"}
                            onclick={() => saveTest(msg)}
                        >
                            {#if saveState === "saving"}
                                Saving...
                            {:else if saveState === "success"}
                                ✓ Saved
                            {:else if saveState === "error"}
                                ✗ Error
                            {:else}
                                Save
                            {/if}
                        </Button>
                    {/if}
                </div>
                <div class="ml-2">
                    {#if isCode && saveState === "success"}
                        <Button color="purple" onclick={RunTest}
                            >Run Test</Button
                        >
                    {/if}
                </div>
            </div>
        </div>
        {#if isCode}
            <Code message={msg} />
        {:else}
            <p class="font-sans whitespace-pre-wrap break-words">
                {msg}
            </p>
        {/if}

        {#if testId && saveState === "success"}
            <p class="text-xs text-gray-600 mt-2 font-mono">
                Test ID: {testId}
            </p>
        {/if}
    </div>
</div>

<Modal title="Test Result" form bind:open={logShowModal} {size} {placement}>
    <div class="space-y-4">
        {#if logRunning}
            <Spinner type="dots" color="purple" />
        {/if}
        <pre
            class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
      {logModalMSG}
    </pre>
    </div>
</Modal>
