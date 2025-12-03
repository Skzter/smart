<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import { getUserChats } from "$lib/api";
    import type { ApiChatSummary } from "$types/api";
    import { user } from "$lib/shared.svelte";
    import { LucideMessageSquare } from "@lucide/svelte";

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
            {#each items as item}
                <Sidebar.MenuItem>
                    <Sidebar.MenuButton
                        class="h-full py-1"
                        onclick={() => {
                            console.log(
                                "changing to: ",
                                item.userId,
                                item.chatId,
                            );
                        }}
                    >
                        <LucideMessageSquare class="mr-1" />
                        <div>
                            <p class="text-sm text-gray-900 dark:text-gray-200">
                                {item.title == "" ? "Neuer Chat" : item.title}
                            </p>
                            <span
                                class="font-mono text-xs text-gray-500 dark:text-gray-400"
                                >{item.updatedAt.substring(11, 16)}</span
                            >
                        </div>
                        <span> </span>
                    </Sidebar.MenuButton>
                </Sidebar.MenuItem>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
