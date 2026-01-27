<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import ChatSummary from "./ChatSummary.svelte";

    let {
        group,
        updateChatSummary,
    }: {
        group: {
            label: string;
            summaries: ApiChatSummary[];
        };
        updateChatSummary: (chatId: string, updated: ApiChatSummary) => void;
    } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat (chat.chatId)}
                <ChatSummary
                    summary={chat}
                    onUpdate={(updated) =>
                        updateChatSummary(chat.chatId, updated)}
                ></ChatSummary>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
