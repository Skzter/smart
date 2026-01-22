<script lang="ts">
    import * as Sidebar from "$lib/components/ui/sidebar";
    import NewChatButton from "./NewChatButton.svelte";
    import TimeFilter from "./TimeFilter.svelte";
    import ChatFilter from "./ChatFilter.svelte";

    import { onMount } from "svelte";
    import { getGroups, createGroup } from "$lib/api";
    import { user } from "$lib/shared.svelte";

    import * as Dialog from "$lib/components/ui/dialog";
    import { Button } from "$lib/components/ui/button";
    import { Input } from "$lib/components/ui/input";

    import type { ApiGroup } from "$types/api";

    let groups: ApiGroup[] = [];
    let open = false;

    let groupName = "";
    let description = "";

    async function loadGroups() {
        try {
            groups = await getGroups();
        } catch (e) {
            console.error("Failed to load groups", e);
            groups = [];
        }
    }

    async function submit() {
        const userId = user.id;
        if (!userId) return;
        if (!groupName.trim()) return;

        await createGroup({
            groupName,
            description,
            userId,
        });

        groupName = "";
        description = "";
        open = false;

        await loadGroups();
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

            {#if groups.length === 0}
                <p class="mt-1 text-xs text-muted-foreground">
                    Keine Gruppen vorhanden
                </p>
            {:else}
                <ul class="mt-1 space-y-1">
                    {#each groups as group}
                        <li class="text-xs truncate">{group.name}</li>
                    {/each}
                </ul>
            {/if}
        </div>
    </div>
</Sidebar.Header>
