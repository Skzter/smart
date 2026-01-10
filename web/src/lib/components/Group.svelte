<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import ChatSummary from "./ChatSummary.svelte";


    let {
        group = $bindable(),
        updateChatTitleStance,
    }: {
        group: {
            label: string;
            summaries: ApiChatSummary[];
        };
        updateChatTitleStance: (chatId: string, title: string) => void;
    } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat, key (chat.chatId)}
                <ChatSummary bind:summary={group.summaries[key]} {updateChatTitleStance}></ChatSummary>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
