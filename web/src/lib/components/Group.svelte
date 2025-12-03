<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { getUserChats } from "$lib/Api";
    import type { ApiChatSummary } from "$types/api";
    import { user } from "$lib/shared.svelte";

    let items = $state<ApiChatSummary[]>([]);
    console.log(user.id);

    (async () => {
        if (user.id == undefined) {
            return;
        }
        try {
            items = (await getUserChats(user.id)) as ApiChatSummary[];
        } catch (error) {
            console.log(error);
        }
    })();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel>Application</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each items as item (item.title)}
                <Sidebar.MenuItem>
                    <Sidebar.MenuButton>
                        {#snippet child({ props })}
                            <a
                                onclick={() => {
                                    console.log(
                                        "changing to: ",
                                        item.userId,
                                        item.chatId,
                                    );
                                }}
                                {...props}
                            >
                                <span>{item.title}</span>
                            </a>
                        {/snippet}
                    </Sidebar.MenuButton>
                </Sidebar.MenuItem>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
