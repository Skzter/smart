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
