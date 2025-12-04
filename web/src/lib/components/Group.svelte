<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import type { ApiChatSummary } from "$types/api";
    import { LucideMessageSquare } from "@lucide/svelte";

    let {
        group = $bindable(),
    }: { group: { label: String; summaries: ApiChatSummary[] } } = $props();
</script>

<Sidebar.Group>
    <Sidebar.GroupLabel class="uppercase">{group.label}</Sidebar.GroupLabel>
    <Sidebar.GroupContent>
        <Sidebar.Menu>
            {#each group.summaries as chat}
                <Sidebar.MenuItem>
                    <Sidebar.MenuButton
                        class="h-full py-1 my-1"
                        onclick={() => {
                            console.log(
                                "changing to: ",
                                chat.userId,
                                chat.chatId,
                            );
                        }}
                    >
                        <LucideMessageSquare class="mr-1" />
                        <div class="h-full flex flex-col justify-between">
                            <p class="text-sm text-gray-900 dark:text-gray-200">
                                {chat.title == "" ? "Neuer Chat" : chat.title}
                            </p>
                            <span
                                class="font-mono text-xs text-gray-500 dark:text-gray-400"
                                >{chat.updatedAt.substring(11, 16)}</span
                            >
                        </div>
                    </Sidebar.MenuButton>
                </Sidebar.MenuItem>
            {/each}
        </Sidebar.Menu>
    </Sidebar.GroupContent>
</Sidebar.Group>
