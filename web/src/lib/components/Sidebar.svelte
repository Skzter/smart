<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Group from "./Group.svelte";
    import { getUserChats } from "$lib/api";
    import type { ApiChatSummary } from "$types/api";
    import { ChatDate, ChatFilter, user } from "$lib/shared.svelte";
    import Spinner from "./ui/spinner/spinner.svelte";
    import SidebarHeader from "$lib/components/SidebarHeader.svelte";
    import { SvelteMap } from "svelte/reactivity";
    import { toast } from "svelte-sonner";
    import type { DateRange } from "bits-ui";
    import User from "./User.svelte";

    let error = $state<string>("");
    let items = $state<ApiChatSummary[] | undefined>(undefined);

    let groupState = $derived(
        updateGroupsWithDateRange(
            items,
            ChatDate.Range,
            ChatFilter.sortBy,
            ChatFilter.timeFilter,
        ),
    );

    $effect(() => {
        if (!user.id) return;
        (async () => {
            try {
                items = (await getUserChats()) as ApiChatSummary[];
            } catch (err) {
                if (err instanceof Error) {
                    error = err.message;
                    toast.error(error, {
                        description: "Das war wohl nichts mit der Historie.",
                    });
                } else {
                    error = "Unbekannter Fehler";
                    toast.error(error, {
                        description: "Das war wohl nichts mit der Historie.",
                    });
                }
            }
        })();
    });

    function updateGroupsWithDateRange(
        items: ApiChatSummary[] | undefined,
        dateRange: DateRange | undefined,
        sortBy: "recent" | "created",
        timeFilter: "all" | "today" | "week" | "month",
    ): { label: string; summaries: ApiChatSummary[] }[] {
        if (!items) return [];

        let filteredItems = items.filter((item) => {
            const time = new Date(Date.parse(item.updatedAt));

            if (!applyTimeFilter(time, timeFilter)) {
                return false;
            }

            return isWithinDateRange(time, dateRange);
        });

        filteredItems = sortItems(filteredItems, sortBy);

        const groups = new SvelteMap<string, ApiChatSummary[]>();

        function add(group: string, item: ApiChatSummary) {
            if (groups.has(group)) {
                groups.get(group)?.push(item);
            } else {
                groups.set(group, [item]);
            }
        }

        filteredItems.forEach((item) => {
            const time = new Date(Date.parse(item.updatedAt));
            const category = categorizeByDate(time);
            add(category, item);
        });

        const result: { label: string; summaries: ApiChatSummary[] }[] = [];
        groups.forEach((value, key) => {
            result.push({ label: key, summaries: value });
        });

        return result;
    }

    function sortItems(
        items: ApiChatSummary[],
        sortBy: "recent" | "created",
    ): ApiChatSummary[] {
        return [...items].sort((a, b) => {
            const dateA =
                sortBy === "recent"
                    ? new Date(a.updatedAt)
                    : new Date(a.createdAt);
            const dateB =
                sortBy === "recent"
                    ? new Date(b.updatedAt)
                    : new Date(b.createdAt);
            return dateB.getTime() - dateA.getTime();
        });
    }

    function applyTimeFilter(
        chatTime: Date,
        timeFilter: "all" | "today" | "week" | "month",
    ): boolean {
        if (timeFilter === "all") return true;

        const now = new Date();
        const today = new Date(
            now.getFullYear(),
            now.getMonth(),
            now.getDate(),
        );
        const chatDate = new Date(
            chatTime.getFullYear(),
            chatTime.getMonth(),
            chatTime.getDate(),
        );

        if (timeFilter === "today") {
            return chatDate.getTime() === today.getTime();
        }

        const daysDiff = Math.floor(
            (today.getTime() - chatDate.getTime()) / (1000 * 60 * 60 * 24),
        );

        if (timeFilter === "week") {
            return daysDiff <= 7;
        }

        if (timeFilter === "month") {
            return daysDiff <= 30;
        }

        return true;
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
        {#if error != ""}
            <Sidebar.Group>
                <Sidebar.GroupLabel>{error}</Sidebar.GroupLabel>
            </Sidebar.Group>
        {:else if items === undefined}
            <Sidebar.Group class="mt-2 flex items-center justify-center">
                <Spinner class="size-6"></Spinner>
            </Sidebar.Group>
        {:else}
            {#each groupState, index (index)}
                <Group bind:group={groupState[index]}></Group>
            {/each}
        {/if}
    </Sidebar.Content>
    <Sidebar.Footer>
        <User />
    </Sidebar.Footer>
</Sidebar.Root>
