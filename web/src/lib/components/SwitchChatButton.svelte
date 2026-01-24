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
            user.id = currChat.userId;
            chat.id = currChat.chatId;
            chat.isLoading = false;
            const response = await getChatById();
            messages.length = 0;
            messages.push(...convertApiMessagesToMessages(response.messages));
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

    function convertApiMessagesToMessages(
        apiMessages: ApiMessage[],
    ): Message[] {
        const messages: Message[] = [];

        for (let i = 0; i < apiMessages.length; i++) {
            const userMessage = apiMessages[i];

            // Skip if not a user message
            if (userMessage?.role !== "user") {
                continue;
            }

            let answerMessage = null;

            // Check next messages for assistant responses
            for (let j = i + 1; j < apiMessages.length; j++) {
                const currentMsg = apiMessages[j];

                // Stop if we hit another user message
                if (currentMsg.role === "user") {
                    break;
                }

                // If it's an assistant message
                if (currentMsg.role === "assistant") {
                    // Try to parse as JSON to check if it's a validation message
                    try {
                        const parsed = JSON.parse(currentMsg.body);
                        if (
                            Object.prototype.hasOwnProperty.call(
                                parsed,
                                "valid",
                            )
                        ) {
                            // If validation failed, use the message as answer
                            if (!parsed.valid && parsed.message) {
                                answerMessage = parsed.message;
                                break;
                            }
                            // If valid is true, continue looking for actual answer
                        } else {
                            // Not a validation message, use as answer
                            answerMessage = currentMsg.body;
                            break;
                        }
                    } catch {
                        // Not JSON, use as regular answer
                        answerMessage = currentMsg.body;
                        break;
                    }
                }
            }

            // Only add if we found an answer
            if (answerMessage) {
                messages.push({
                    question: userMessage.body,
                    answer: answerMessage,
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
