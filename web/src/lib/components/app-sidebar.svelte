<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Group from "./Group.svelte";
    import { getUserChats } from "$lib/api";
    import type { ApiChatSummary } from "$types/api";
    import { user } from "$lib/shared.svelte";

    let groupState = $state<{ label: String; summaries: ApiChatSummary[] }[]>(
        [],
    );
    let groups = new Map<string, ApiChatSummary[]>();
    function add(group: string, item: ApiChatSummary) {
        if (groups.has(group)) {
            groups.get(group)?.push(item);
        } else {
            groups.set(group, [item]);
        }
    }

    let items: ApiChatSummary[] | undefined;
    (async () => {
        if (user.id == undefined) {
            return;
        }
        try {
            items = (await getUserChats(user.id)) as ApiChatSummary[];
        } catch (error) {
            console.log(error);
        }

        console.log(items);
        items?.forEach((item) => {
            let now = new Date();
            let time = new Date(Date.parse(item.updatedAt));
            if (
                now.getFullYear === time.getFullYear &&
                now.getMonth === time.getMonth
            ) {
                let days = now.getDate() - time.getDate();
                switch (days) {
                    case 0:
                        add("Heute", item);
                        return;
                    case 1:
                        add("Gestern", item);
                        return;
                }
                if (days <= 7) {
                    add("letzte Woche", item);
                    return;
                }
                add("letzten Monat", item);
                return;
            }
            add("früher", item);
        });

        let g: { label: String; summaries: ApiChatSummary[] }[] = [];
        groups.forEach((value, key) => {
            g.push({ label: key, summaries: value });
        });
        groupState = g;
    })();
</script>

<Sidebar.Root>
    <Sidebar.Content>
        {#each groupState as g, index}
            <Group bind:group={groupState[index]}></Group>
        {/each}
    </Sidebar.Content>
</Sidebar.Root>
