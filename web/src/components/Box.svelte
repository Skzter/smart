<script lang="ts">
    import { Button } from "flowbite-svelte";
    import { createTest } from "../lib/Api.ts";
    let { msg, name, userId, conversationId, showSave = false} = $props();

    let saveState = $state<'idle' | 'saving' | 'success' | 'error'>('idle');
    let errorMessage = $state("");
    let testId = $state<string | undefined>(undefined);

    async function saveTest(testcode: string) {
        if (!userId || !conversationId) {
            errorMessage = "Failed to save test: userId or conversationId is missing";
            console.error(errorMessage, {userId,conversationId});
            return;
        }

        const sanitizedUserId = userId.includes('|') 
            ? userId.split('|')[1] 
            : userId;

        const request = {
            userId: sanitizedUserId,
            conversationId: conversationId,
            code: testcode,
        };

        try {
            saveState = 'saving';
            const response = await createTest(request);
            testId = response.data.testcaseId
            console.log("Test saved successfully:", response.data);
            saveState = 'success';
        } catch (error: any) {
            console.error("Failed to save test:", error);
            saveState = 'error';
            errorMessage = error?.response?.data?.error || "Failed to save test. Please try again.";

            setTimeout(() => {
                saveState = 'idle';
            }, 2000);
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
        class:bg-gray-200={name === "Bot"}
    >
        <div class="flex items-start justify-between">
            <h1 class="tracking-wide uppercase font-bold text-xl">
                {name}
            </h1>
           <div class="ml-2">
                {#if showSave}
                    <Button
                        color={saveState === 'success' ? 'green' : saveState === 'error' ? 'red' : 'purple'}
                        disabled={saveState === 'saving' || saveState === 'success'}
                        onclick={() => saveTest(msg)}
                    >
                        {#if saveState === 'saving'}
                            Saving...
                        {:else if saveState === 'success'}
                            ✓ Saved
                        {:else if saveState === 'error'}
                            ✗ Error
                        {:else}
                            Save
                        {/if}
                    </Button>
                {/if}
            </div>
        </div>
        <p class="font-sans whitespace-pre-wrap break-words">
            {msg}
        </p>
        {#if testId && saveState === 'success'}
            <p class="text-xs text-gray-600 mt-2 font-mono">
                Test ID: {testId}
            </p>
        {/if}
    </div>
</div>
