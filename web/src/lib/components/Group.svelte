<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import SwitchChatButton from "./SwitchChatButton.svelte";

    let {
        group,
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
            {#each group.summaries as chat}
                <Sidebar.MenuItem>
                    <SwitchChatButton currChat={chat} />
                </Sidebar.MenuItem>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
