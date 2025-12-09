<script lang="ts">
    import { LucideMessageSquare } from "@lucide/svelte";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary, ApiMessage } from "$types/api";
    import type { Message } from "$types/message";
    import { getChatById } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { chat, messages, user } from "$lib/shared.svelte";

    let {
        currChat,
    }: {
        currChat: ApiChatSummary;
    } = $props();

    function formatToCET(iso?: string) {
        if (!iso) return "";
        try {
            const d = new Date(iso);
            return new Intl.DateTimeFormat("de-DE", {
                hour: "2-digit",
                minute: "2-digit",
                timeZone: "Europe/Berlin",
            }).format(d);
        } catch {
            return iso.substring ? iso.substring(11, 16) : "";
        }
    }

    async function invokeSwitchChat() {
        try {
            const response = await getChatById();
            changeChat(response.userId, response.id, response.messages);
        } catch (error) {
            let errorMsg = "Unbekannter Fehler";
            if (error instanceof Error) {
                errorMsg = error.message;
            }

            toast.error("Speichern fehlgeschlagen", {
                description: errorMsg,
            });
        }
    }

    function changeChat(userId: string, chatId: string, msg: ApiMessage[]) {
        messages.length = 0;
        messages.push(...convertApiMessagesToMessages(msg));
        user.id = userId;
        chat.id = chatId;
        chat.isLoading = false;
    }

    function convertApiMessagesToMessages(
        apiMessages: ApiMessage[],
    ): Message[] {
        const messages: Message[] = [];

        for (let i = 0; i < apiMessages.length - 1; i += 2) {
            const userMessage = apiMessages[i];
            const assistantMessage = apiMessages[i + 1];

            if (
                userMessage?.role === "user" &&
                assistantMessage?.role === "assistant"
            ) {
                messages.push({
                    question: userMessage.body,
                    answer: assistantMessage.body,
                });
            }
        }

        return messages;
    }
</script>

<Sidebar.MenuButton class="h-full py-1 my-1 flex">
    <LucideMessageSquare class="mr-1" />
    <div
        class="h-full flex flex-col justify-between flex-1 min-w-0"
        role="button"
        tabindex="0"
        onclick={() => invokeSwitchChat()}
        onkeydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                invokeSwitchChat();
            }
        }}
    >
        <p class="text-sm text-gray-900 dark:text-gray-200 truncate">
            {currChat.title === "" ? "Neuer Chat" : currChat.title}
        </p>
        <span class="font-mono text-xs text-gray-500 dark:text-gray-400"
            >{formatToCET(currChat.updatedAt)}</span
        >
    </div>
</Sidebar.MenuButton>
