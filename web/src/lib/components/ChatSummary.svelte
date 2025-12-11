<script lang="ts">
    import type { ApiChatSummary, ApiMessage } from "$types/api";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { LucideMessageSquare, Pencil } from "@lucide/svelte";
    import type { Message } from "$types/message";
    import { getChatById } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { chat, messages, user } from "$lib/shared.svelte";

    let { summary = $bindable() }: { summary: ApiChatSummary } = $props();
    let edit = $state(false);

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

    function focusAction(el: HTMLInputElement) {
        if (summary.title === "") {
            el.select();
        } else {
            el.focus();
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

    async function invokeSwitchChat() {
        try {
            user.id = summary.userId;
            chat.id = summary.chatId;
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
</script>

<Sidebar.MenuItem>
    <Sidebar.MenuButton
        class="h-12 py-1 mb-1 flex"
        role="button"
        tabindex={0}
        onclick={invokeSwitchChat}
    >
        <LucideMessageSquare class="mr-1" />
        <div class="flex flex-col justify-between min-w-0 mr-2">
            {#if edit}
                <input
                    use:focusAction
                    id={`title${summary.chatId}`}
                    value={summary.title === "" ? "Neuer Chat" : summary.title}
                    onfocusout={(e) => {
                        summary.title = (e.target as HTMLInputElement).value;
                        edit = false;
                    }}
                    onkeydown={(e) => {
                        if (e.key == "Enter") {
                            summary.title = (
                                e.target as HTMLInputElement
                            ).value;
                            edit = false;
                            edit = false;
                        }
                        if (e.key == "Escape") {
                            edit = false;
                        }
                    }}
                />
            {:else}
                <p class="truncate">
                    {summary.title === "" ? "Neuer Chat" : summary.title}
                </p>{/if}
            <span class="font-mono text-xs text-gray-500 dark:text-gray-400"
                >{formatToCET(summary.updatedAt)}</span
            >
        </div>
    </Sidebar.MenuButton>
    <Sidebar.MenuAction
        class="my-2 mr-1 h-8 w-8 top-0! right-0!"
        onclick={() => {
            edit = true;
        }}
    >
        <Pencil class=" text-gray-500 dark:text-gray-400" />
    </Sidebar.MenuAction>
</Sidebar.MenuItem>
