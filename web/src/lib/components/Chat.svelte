<script lang="ts">
    import UserMessage from "./UserMessage.svelte";
    import BotMessage from "./BotMessage.svelte";
    import Dots from "./Dots.svelte";

    import { chat, messages, GroupsState } from "$lib/shared.svelte";
    import { assignChatToGroups, removeChatFromGroup } from "$lib/api";
    import { toast } from "svelte-sonner";

    const groupNameById = $derived(
        new Map<string, string>(GroupsState.items.map((g) => [g.id, g.name])),
    );

    let selectedGroupIds = $derived.by(() => [...(chat.groups ?? [])]);

    let container: HTMLElement | undefined = $state();

    $effect(() => {
        if (container && (chat.isLoading || messages.length > 0)) {
            setTimeout(() => {
                container?.scrollTo({
                    top: container.scrollHeight,
                    behavior: "smooth",
                });
            }, 50);
        }
    });

    async function saveGroups() {
        if (!chat.id) return;

        try {
            await assignChatToGroups(chat.id, selectedGroupIds);
            chat.groups = [...selectedGroupIds];
            toast.success("Gruppen gespeichert");
        } catch {
            toast.error("Gruppen konnten nicht gespeichert werden");
        }
    }

    async function removeGroup(groupId: string) {
        if (!chat.id) return;

        try {
            await removeChatFromGroup(chat.id, groupId);
            chat.groups = chat.groups.filter((g) => g !== groupId);
            selectedGroupIds = selectedGroupIds.filter((g) => g !== groupId);
        } catch {
            toast.error("Gruppe konnte nicht entfernt werden");
        }
    }
</script>

<div class="flex flex-1 flex-col gap-4 p-4 pt-0 h-full">
    <div class="flex-1 flex items-start justify-center min-h-0">
        <div
    bind:this={container}
    class="w-full max-w-6xl bg-muted/50 h-full rounded-xl md:min-h-min overflow-auto p-6 min-h-0 flex flex-col"
>
            {#if chat.groups?.length}
                <div class="mb-2 flex flex-wrap gap-2">
                    {#each chat.groups as groupId}
                        <button
                            type="button"
                            class="flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground hover:bg-muted/80"
                            aria-label="Gruppe entfernen"
                            onclick={() => removeGroup(groupId)}
                        >
                            {groupNameById.get(groupId) ?? groupId}
                            <span class="text-xs opacity-70">✕</span>
                        </button>
                    {/each}
                </div>
            {/if}

            {#if selectedGroupIds.length !== (chat.groups?.length ?? 0)}
                <button
                    class="mb-4 w-fit rounded-md bg-muted px-3 py-1 text-xs text-muted-foreground hover:bg-muted/80"
                    onclick={saveGroups}
                >
                    Gruppen speichern
                </button>
            {/if}


            {#if messages.length === 0}
                <div
                    class="flex items-center justify-center flex-1 text-muted-foreground"
                >
                    <p>Start a conversation...</p>
                </div>
            {:else}
                <div class="flex-1"></div>
                <div class="flex flex-col gap-4">
                    {#each messages as message}
                        {#if message.t === "user"}
                            <UserMessage message={message.Message} />
                        {:else}
                            <BotMessage msg={message} />
                        {/if}
                    {/each}

                    {#if chat.isLoading}
                        <Dots />
                    {/if}
                </div>
            {/if}
        </div>
    </div>
</div>
