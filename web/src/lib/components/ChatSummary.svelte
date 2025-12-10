<script lang="ts">
    import type { ApiChatSummary } from "$types/api";
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { LucideMessageSquare, Pencil } from "@lucide/svelte";

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

    function openChat(userId?: string, chatId?: string) {
        console.log("changing to: ", userId, chatId);
    }

    const classes = "h-full flex flex-col justify-between flex-1 min-w-0";

    function focusAction(el: HTMLInputElement) {
        summary.title === "" ? el.select() : el.focus();
    }
</script>

<Sidebar.MenuItem>
    <Sidebar.MenuButton
        class="h-12 py-1 mb-1 flex"
        role="button"
        tabindex={0}
        onclick={() => openChat(summary.userId, summary.chatId)}
    >
        <LucideMessageSquare class="mr-1" />
        <div class="flex flex-col justify-between flex-1 min-w-0">
            {#if edit}
                <input
                    use:focusAction
                    id={`title${summary.chatId}`}
                    class={classes}
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
                <p class={classes}>
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
