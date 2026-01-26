<script lang="ts">
    import type { ApiChatSummary, ApiMessage } from "$types/api";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { LucideMessageSquare, Pencil } from "@lucide/svelte";
    import { getChatById } from "$lib/api";
    import { toast } from "svelte-sonner";
    import { type Message, chat, messages, user } from "$lib/shared.svelte";

    let {
        summary = $bindable(),
        onUpdate,
    }: {
        summary: ApiChatSummary;
        onUpdate?: (updated: ApiChatSummary) => void;
    } = $props();
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

        apiMessages.forEach((element) => {
            if (element.role === "user") {
                messages.push({ t: "user", Message: element.body });
                return;
            }

            if (element.type === "Validation") {
                let msg: { valid: boolean; message: string } = JSON.parse(
                    element.body,
                );
                messages.push({
                    t: "validation",
                    Message: msg.valid ? "Prompt ist Valide" : msg.message,
                });
                return;
            }

            if (element.type === "Generation") {
                messages.push({ t: "generation", Message: element.body });
                return;
            }
        });

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
                        onUpdate?.(summary);
                        edit = false;
                    }}
                    onkeydown={(e) => {
                        if (e.key == "Enter") {
                            summary.title = (
                                e.target as HTMLInputElement
                            ).value;
                            onUpdate?.(summary);
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
