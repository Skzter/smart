<script lang="ts">
    import { Play } from "@lucide/svelte";
    import { Button } from "$lib/components/ui/button/index.js";
    import { chat, user } from "$lib/shared.svelte";
    import { runContainer } from "$lib/api";
    import { toast } from "svelte-sonner";

    let {
        isLoading = $bindable(),
        activeTab = $bindable(),
        testResult = $bindable(),
    }: {
        isLoading: boolean;
        activeTab: string;
        testResult: string;
    } = $props();

    async function runTest() {
        if (!user.id || !chat.id || !chat.currTestId) {
            console.error(
                "Missing IDs - ChatID: " +
                    chat.id +
                    " UserID: " +
                    user.id +
                    " TestID: " +
                    chat.currTestId,
            );
            toast.error("Speichern fehlgeschlagen", {
                description: "Benutzer-, -Konversations oder Test-ID fehlt.",
            });
            return;
        }

        const sanitizedUserId = user.id.includes("|")
            ? user.id.split("|")[1]
            : user.id;

        isLoading = true;
        try {
            const response = await runContainer({
                userId: sanitizedUserId,
                testId: chat.currTestId,
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
</script>

<Button
    variant="outline"
    size="sm"
    class="h-7 gap-1.5 px-2 cursor-pointer"
    onclick={runTest}
    disabled={isLoading}
>
    <Play class="h-3.5 w-3.5" />
    <span class="text-xs">{isLoading ? "Lädt..." : "Ausführen"}</span>
</Button>
