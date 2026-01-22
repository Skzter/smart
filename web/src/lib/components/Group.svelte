<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import ChatSummary from "./ChatSummary.svelte";

    let {
        group = $bindable(),
    }: {
        group: {
            label: string;
            summaries: ApiChatSummary[];
        };
    } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat, key (chat.chatId)}
                <ChatSummary bind:summary={group.summaries[key]}></ChatSummary>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
