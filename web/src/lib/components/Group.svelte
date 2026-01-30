<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import ChatSummary from "./ChatSummary.svelte";

    let {
        group,
        updateChatTitleState,
        updateChatSummary,
    }: {
        group: {
            label: string;
            summaries: ApiChatSummary[];
        };
        updateChatTitleState: (chatId: string, title: string) => void;
        updateChatSummary: (chatId: string, updated: ApiChatSummary) => void;
    } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat (chat.chatId)}
                <ChatSummary
                    {updateChatTitleState}
                    summary={chat}
                    onUpdate={(updated) =>
                        updateChatSummary(chat.chatId, updated)}
                ></ChatSummary>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
