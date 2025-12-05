<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Group from "./Group.svelte";
    import { getUserChats } from "$lib/api";
    import type { ApiChatSummary } from "$types/api";
    import { user } from "$lib/shared.svelte";
    import Spinner from "./ui/spinner/spinner.svelte";
    import SidebarHeader from "$lib/components/SidebarHeader.svelte";
    import { SvelteMap } from "svelte/reactivity";
    import { toast } from "svelte-sonner";

    let error = $state<string>("");
    let groupState = $state<
        { label: string; summaries: ApiChatSummary[] }[] | undefined
    >(undefined);

    let items: ApiChatSummary[] | undefined;
    (async () => {
        let groups = new SvelteMap<string, ApiChatSummary[]>();
        function add(group: string, item: ApiChatSummary) {
            if (groups.has(group)) {
                groups.get(group)?.push(item);
            } else {
                groups.set(group, [item]);
            }
        }

        groupState = undefined;
        if (user.id == undefined) {
            return;
        }
        try {
            items = (await getUserChats(user.id)) as ApiChatSummary[];
        } catch (err) {
            error = "Unbekannter Fehler";
            if (err instanceof Error) {
                error = err.message;
                toast.error(error, {
                    description: "Das war wohl nichts mit der Historie.",
                });
            }
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

        let g: { label: string; summaries: ApiChatSummary[] }[] = [];
        groups.forEach((value, key) => {
            g.push({ label: key, summaries: value });
        });
        groupState = g;
    })();
</script>

<Sidebar.Root>
    <SidebarHeader />
    <Sidebar.Content>
        {#if groupState == undefined}
            <Sidebar.Group class="mt-2 flex items-center justify-center">
                <Spinner class="size-6"></Spinner>
            </Sidebar.Group>
        {:else if error != ""}
            <Sidebar.Group>
                <Sidebar.GroupLabel>{error}</Sidebar.GroupLabel>
            </Sidebar.Group>
        {:else}
            {#each groupState, index (index)}
                <Group bind:group={groupState[index]}></Group>
            {/each}
        {/if}
    </Sidebar.Content>
</Sidebar.Root>
