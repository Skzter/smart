<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar";
    import NewChatButton from "./NewChatButton.svelte";
    import TimeFilter from "./TimeFilter.svelte";
    import ChatFilter from "./ChatFilter.svelte";

    import { onMount } from "svelte";
    import { getGroups, createGroup } from "$lib/api";
    import { user, GroupFilter, GroupsState } from "$lib/shared.svelte";

    import * as Dialog from "$lib/components/ui/dialog";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";

    import { toast } from "svelte-sonner";

    let open = false;
    let groupName = "";
    let description = "";

    async function loadGroups() {
        GroupsState.isLoading = true;
        GroupsState.error = "";
        try {
            GroupsState.items = await getGroups();
        } catch (e) {
            console.error("Failed to load groups", e);
            GroupsState.items = [];
            GroupsState.error = "Failed to load groups";
            toast.error("Gruppen konnten nicht geladen werden", {
                description: "Bitte versuch es später nochmal.",
            });
        } finally {
            GroupsState.isLoading = false;
        }
    }

    function toggleGroup(groupId: string) {
        const selected = GroupFilter.selectedIds;
        if (selected.includes(groupId)) {
            GroupFilter.selectedIds = selected.filter((id) => id !== groupId);
        } else {
            GroupFilter.selectedIds = [...selected, groupId];
        }
    }

    function resetGroupFilter() {
        GroupFilter.selectedIds = [];
    }

    async function submit() {
        const userId = user.id;
        if (!userId) return;

        const name = groupName.trim();
        if (!name) return;

        try {
            await createGroup({
                groupName: name,
                description: description.trim(),
                userId,
            });

            toast.success("Gruppe erstellt");

            groupName = "";
            description = "";
            open = false;

            await loadGroups();
        } catch (e) {
            console.error("Failed to create group", e);
            toast.error("Gruppe konnte nicht erstellt werden");
        }
    }

    onMount(loadGroups);
</script>

<Sidebar.Header class="border-b">
    <div class="flex flex-col gap-2 p-2">
        <NewChatButton />
        <TimeFilter />
        <ChatFilter />

        <!-- Groups -->
        <div class="mt-2 border-t pt-2">
            <div class="flex items-center justify-between text-sm font-medium">
                <span>Groups</span>

                <div class="flex items-center gap-2">
                    <Button
                        size="sm"
                        variant="ghost"
                        onclick={resetGroupFilter}
                        disabled={GroupFilter.selectedIds.length === 0}
                        aria-label="Gruppenfilter zurücksetzen"
                    >
                        Reset
                    </Button>

                    <Dialog.Root bind:open>
                        <Dialog.Trigger>
                            <Button
                                size="icon"
                                variant="ghost"
                                aria-label="Neue Gruppe"
                            >
                                +
                            </Button>
                        </Dialog.Trigger>

                        <Dialog.Content class="sm:max-w-md">
                            <Dialog.Header>
                                <Dialog.Title>Neue Gruppe</Dialog.Title>
                            </Dialog.Header>

                            <div class="flex flex-col gap-2 py-2">
                                <Input
                                    placeholder="Gruppenname"
                                    bind:value={groupName}
                                />
                                <Input
                                    placeholder="Beschreibung (optional)"
                                    bind:value={description}
                                />
                            </div>

                            <Dialog.Footer>
                                <Button
                                    disabled={!groupName.trim() || !user.id}
                                    onclick={submit}
                                >
                                    Erstellen
                                </Button>
                            </Dialog.Footer>
                        </Dialog.Content>
                    </Dialog.Root>
                </div>
            </div>

            {#if GroupsState.isLoading}
                <p class="mt-1 text-xs text-muted-foreground">
                    Loading groups…
                </p>
            {:else if GroupsState.error}
                <p class="mt-1 text-xs text-destructive">
                    {GroupsState.error}
                </p>
            {:else if GroupsState.items.length === 0}
                <p class="mt-1 text-xs text-muted-foreground">
                    Keine Gruppen vorhanden
                </p>
            {:else}
                <ul class="mt-2 space-y-1">
                    {#each GroupsState.items as group}
                        <li>
                            <button
                                type="button"
                                class="flex w-full items-center justify-between rounded-md px-2 py-1 text-left text-xs hover:bg-muted"
                                aria-pressed={GroupFilter.selectedIds.includes(
                                    group.id,
                                )}
                                onclick={() => toggleGroup(group.id)}
                            >
                                <span class="truncate">{group.name}</span>
                                {#if GroupFilter.selectedIds.includes(group.id)}
                                    <span class="ml-2 text-muted-foreground">
                                        ✓
                                    </span>
                                {/if}
                            </button>
                        </li>
                    {/each}
                </ul>
            {/if}
        </div>
    </div>
</Sidebar.Header>
