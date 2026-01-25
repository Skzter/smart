<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import ChatSummary from "./ChatSummary.svelte";


    let {
        group,
        updateChatTitleState,
    }: {
        group: {
            label: string;
            summaries: ApiChatSummary[];
        };
        updateChatTitleState: (chatId: string, title: string) => void;
    } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as summary (summary.chatId)}
                <ChatSummary {summary} {updateChatTitleState}></ChatSummary>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
