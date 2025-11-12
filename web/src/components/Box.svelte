<script lang="ts">
    import { Button } from "flowbite-svelte";
    import { createTest } from "../lib/Api.ts";
    let { msg, name, userId, conversationId, showSave = false} = $props();

    async function saveTest(testcode: string) {
        if (!userId || !conversationId) {
            console.error("Failed to save test: userId or conversationId is missing", {
                userId,
                conversationId
            });
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
            const response = await createTest(request);
            console.log("Test saved successfully:", response.data);
        } catch (error) {
            console.error("Failed to save test:", error);
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
                    <Button color="purple"
                        onclick={() => saveTest(msg)}>
                        Save
                    </Button>
                {/if}
            </div>
        </div>
        <p class="font-sans whitespace-pre-wrap break-words">
            {msg}
        </p>
    </div>
</div>
