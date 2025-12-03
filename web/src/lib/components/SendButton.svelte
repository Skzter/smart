<script lang="ts">
    import { Send } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { chat, messages, user } from "$lib/shared.svelte";
    import type { ApiChatRequest } from "$types/api";
    import { getChatResponse } from "$lib/api";

    let {
        input = $bindable(),
    }: {
        input: string;
    } = $props();

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter" && input.trim() && !e.shiftKey) {
            onclick();
            input = "";
            e.preventDefault();
        }
    }

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

        const userQuestion = input;
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

<Button
    variant="default"
    size="icon"
    class="w-10 h-10"
    disabled={chat.isLoading}
    onkeydown={handleKeyPress}
    {onclick}
>
    <Send />
</Button>
