<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { saveTestLocal } from "$lib/Api";
    import { AxiosError } from "axios";
    import type { SaveState } from "$types/save";
    import { chat, user } from "$lib/shared.svelte";

    let {
        code,
    }: {
        code: string;
    } = $props();

    let errorMessage = $state("");
    let saveState = $state<SaveState>("idle");
    let testId = "";

    $effect(() => {
        console.log("SaveState: " + saveState + " TestID: " + testId);
    });

    async function saveTest(testcode: string) {
        if (!user.id || !chat.id) {
            errorMessage =
                "Failed to save test: userId or conversationId is missing";
            console.error(
                errorMessage + " ChatID: " + chat.id + "UserID: " + user.id,
            );
            return;
        }

        const sanitizedUserId = user.id.includes("|")
            ? user.id.split("|")[1]
            : user.id;

        saveState = "saving";
        try {
            let test = await saveTestLocal({
                userId: sanitizedUserId,
                conversationId: chat.id,
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
