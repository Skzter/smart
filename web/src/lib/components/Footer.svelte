<script lang="ts">
    import { generatePrompt, validatePrompt } from "$lib/api";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import { chat, messages, user } from "$lib/shared.svelte";
    import type { ApiChatRequest } from "$types/api";
    import Prompt from "./Prompt.svelte";
    import SendButton from "./SendButton.svelte";

    let input = $state("");

    async function onclick() {
        if (!user.id) {
            console.error("User is not authenticated.");
            return;
        }

        const userQuestion = input.trim();
        input = "";
        messages.push({ t: "user", Message: userQuestion });
        chat.isLoading = true;
        let paramsChatRequest: ApiChatRequest = {
            prompt: userQuestion,
            userId: user.id,
            chatId: chat.id,
        };

        let valid: boolean;

        try {
            const validationAnswer = await validatePrompt(paramsChatRequest);
            valid = validationAnswer.message.body == "";
            messages.push({
                t: "validation",
                Message: valid
                    ? "Prompt ist Valide"
                    : validationAnswer.message.body,
            });
            chat.id = validationAnswer.chatId;
            paramsChatRequest.chatId = chat.id;

            if (valid) {
                const generationAnswer =
                    await generatePrompt(paramsChatRequest);
                messages.push({
                    t: "generation",
                    Message: generationAnswer.message.body,
                });
            }
        } catch (err: unknown) {
            messages.push({ t: "error", Message: (err as Error).message });
        } finally {
            chat.isLoading = false;
        }
    }
</script>

<div class="p-4 pt-0 sticky bottom-0 bg-background z-10">
    <div class="flex items-center justify-center w-full">
        <div class="w-full max-w-6xl">
            <ButtonGroup.Root
                class="[--radius:1rem] w-full flex items-center justify-center gap-3 "
            >
                <ButtonGroup.Root class="flex-1">
                    <div class="flex w-full items-center gap-2">
                        <Prompt {onclick} bind:input />
                        <SendButton {onclick} bind:input />
                    </div>
                </ButtonGroup.Root>
            </ButtonGroup.Root>
        </div>
    </div>
</div>
