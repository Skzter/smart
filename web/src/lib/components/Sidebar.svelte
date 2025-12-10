<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Group from "./Group.svelte";
    import { getChats } from "$lib/api";
    import type { ApiChatSummary } from "$types/api";
    import { ChatDate, user } from "$lib/shared.svelte";
    import Spinner from "./ui/spinner/spinner.svelte";
    import SidebarHeader from "$lib/components/SidebarHeader.svelte";
    import { SvelteMap } from "svelte/reactivity";
    import { toast } from "svelte-sonner";
    import type { DateRange } from "bits-ui";

    let error = $state<string>("");
    let items = $state<ApiChatSummary[] | undefined>(undefined);

    let groupState = $derived(updateGroupsWithDateRange(items, ChatDate.Range));
    $inspect(groupState[0].summaries[0].title);

    (async () => {
        if (user.id == undefined) {
            return;
        }
        try {
            items = (await getUserChats()) as ApiChatSummary[];
        } catch (err) {
            error = "Unbekannter Fehler";
            if (err instanceof Error) {
                error = err.message;
                toast.error(error, {
                    description: "Das war wohl nichts mit der Historie.",
                });
            }
        }
    })();

    function updateGroupsWithDateRange(
        items: ApiChatSummary[] | undefined,
        dateRange: DateRange | undefined,
    ): { label: string; summaries: ApiChatSummary[] }[] {
        if (!items) return [];

        const groups = new SvelteMap<string, ApiChatSummary[]>();

        function add(group: string, item: ApiChatSummary) {
            if (groups.has(group)) {
                groups.get(group)?.push(item);
            } else {
                groups.set(group, [item]);
            }
        }

        items.forEach((item) => {
            const time = new Date(Date.parse(item.updatedAt));

            // Check if item is within date range
            if (isWithinDateRange(time, dateRange)) {
                const category = categorizeByDate(time);
                add(category, item);
            }
        });

        const result: { label: string; summaries: ApiChatSummary[] }[] = [];
        groups.forEach((value, key) => {
            result.push({ label: key, summaries: value });
        });
        $inspect(result);

        return result;
    }

    function isWithinDateRange(
        chatTime: Date,
        dateRange: DateRange | undefined,
    ): boolean {
        if (!dateRange?.start || !dateRange?.end) {
            return true;
        }
        const rangeStart = dateRange.start.toDate("UTC");
        rangeStart.setUTCHours(0, 0, 0, 0);

        const rangeEnd = dateRange.end.toDate("UTC");
        rangeEnd.setUTCHours(23, 59, 59, 999);

        const chatDate = new Date(chatTime);

        return chatDate >= rangeStart && chatDate <= rangeEnd;
    }

    function categorizeByDate(chatTime: Date): string {
        const now = new Date();
        const chatDate = new Date(chatTime);

        if (
            now.getFullYear() === chatDate.getFullYear() &&
            now.getMonth() === chatDate.getMonth()
        ) {
            const daysDiff = now.getDate() - chatDate.getDate();

            if (daysDiff === 0) return "Heute";
            if (daysDiff === 1) return "Gestern";
            if (daysDiff <= 7) return "letzte Woche";
            return "letzten Monat";
        }

        return "früher";
    }
</script>

<Sidebar.Root>
    <SidebarHeader />
    <Sidebar.Content>
        {#if items === undefined}
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
