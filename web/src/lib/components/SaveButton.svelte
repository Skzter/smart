<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { saveTestLocal } from "$lib/Api";
    import { AxiosError } from "axios";

    let {
        code,
        userId,
        conversationId,
    }: {
        code: string;
        userId: string;
        conversationId: string;
    } = $props();

    let errorMessage = $state("");
    let saveState = $state("");
    let testId = "";

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

        saveState = "saving";
        try {
            let test = await saveTestLocal({
                userId: sanitizedUserId,
                conversationId: conversationId,
                code: testcode,
            });
            testId = test.testcaseId;
            console.log("Test saved successfully:", test);
            saveState = "success";
        } catch (error) {
            console.error("Failed to save test:", error);
            saveState = "error";
            let msg = "Unknown error";
            if (error instanceof AxiosError) {
                msg = error.message;
            }
            errorMessage = msg;
        }

        setTimeout(() => {
            saveState = "idle";
        }, 2000);
    }
</script>

<Button
    variant="ghost"
    size="sm"
    class="h-7 gap-1.5 px-2"
    onclick={() => saveTest(code)}
>
    <Save class="h-3.5 w-3.5" />
    <span class="text-xs">Speichern</span>
</Button>
