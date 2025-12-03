<script lang="ts">
    import { Save } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { saveTestLocal } from "$lib/api";
    import { AxiosError } from "axios";
    import type { SaveState } from "$types/save";
    import { chat, user } from "$lib/shared.svelte";
    import { toast } from "svelte-sonner";

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
                "Missing IDs - ChatID: " + chat.id + " UserID: " + user.id,
            );
            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer- oder Konversations-ID fehlt.",
            });
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

            toast.success("Test erfolgreich gespeichert!");
        } catch (error) {
            console.error("Failed to save test:", error);
            saveState = "error";

            let errorMsg = "Unbekannter Fehler";
            if (error instanceof AxiosError) {
                errorMsg = error.response?.data?.message || error.message;
            } else if (error instanceof Error) {
                errorMsg = error.message;
            }
            errorMessage = msg;
        }

            toast.error("Speichern fehlgeschlagen", {
                description: errorMsg,
            });
        } finally {
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
        disabled={saveState === "saving"}
>
    <Save class="h-3.5 w-3.5" />
    <span class="text-xs">
        {saveState === "saving" ? "Speichert..." : "Speichern"}
    </span>
</Button>
