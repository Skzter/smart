<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import { LucideMessageSquare } from "@lucide/svelte";

    export let group: { label: string; summaries: ApiChatSummary[] };

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
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat}
                <Sidebar.MenuItem>
                    <Sidebar.MenuButton class="h-full py-1 my-1 flex">
                        <LucideMessageSquare class="mr-1" />
                        <div
                            class="h-full flex flex-col justify-between flex-1 min-w-0"
                            role="button"
                            tabindex="0"
                            on:click={() => openChat(chat.userId, chat.chatId)}
                            on:keydown={(e) => {
                                if (e.key === "Enter" || e.key === " ") {
                                    e.preventDefault();
                                    openChat(chat.userId, chat.chatId);
                                }
                            }}
                        >
                            <p
                                class="text-sm text-gray-900 dark:text-gray-200 truncate"
                            >
                                {chat.title === "" ? "Neuer Chat" : chat.title}
                            </p>
                            <span class="font-mono text-xs text-gray-500 dark:text-gray-400">{formatToCET(chat.updatedAt)}</span>
                        </div>
                    </Sidebar.MenuButton>
                </Sidebar.MenuItem>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
