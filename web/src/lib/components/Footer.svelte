<script lang="ts">
    import { getChatResponse } from "$lib/Api";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import { chat, messages, user } from "$lib/shared.svelte";
    import type { ApiChatRequest } from "$types/api";
    import Prompt from "./Prompt.svelte";
    import SendButton from "./SendButton.svelte";

    let input = $state("");

    let paramsChatRequest: ApiChatRequest = $derived({
        message: {
            body: "",
            role: "user",
        },
        userId: user.id,
        conversationId: chat.id,
    });

    async function onclick() {
        if (!user.id) {
            console.error("User is not authenticated.");
            return;
        }

        const userQuestion = input.trim();
        input = "";
        messages.push({
            question: userQuestion,
            answer: "",
        });
        chat.isLoading = true;
        paramsChatRequest.message.body = userQuestion;
        paramsChatRequest.userId = user.id;

        try {
            const answer = await getChatResponse(paramsChatRequest);
            messages[messages.length - 1].answer = answer.message.body;
            chat.id = answer.conversationId;
        } catch (err: unknown) {
            messages[messages.length - 1].answer = (err as Error).message;
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
