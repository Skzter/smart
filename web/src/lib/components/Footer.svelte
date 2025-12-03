<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import Prompt from "./Prompt.svelte";
    import SendButton from "./SendButton.svelte";

    let input = $state("");

    let {
        userId,
        isLoading = $bindable(),
    }: {
        userId: string;
        isLoading: boolean;
    } = $props();
</script>

<!-- Hier nochmal zwei Components für Input und die Button fürs abschicken -->
<div class="p-4 pt-0 sticky bottom-0 bg-background z-10">
    <div class="flex items-center justify-center w-full">
        <div class="w-full max-w-6xl">
            <ButtonGroup.Root
                class="[--radius:1rem] w-full flex items-center justify-center gap-3"
            >
                <ButtonGroup.Root class="flex-1">
                    <div class="flex w-full items-center gap-2">
                        <Prompt bind:isLoading bind:input />
                        <SendButton {userId} bind:input bind:isLoading />
                    </div>
                </ButtonGroup.Root>
            </ButtonGroup.Root>
        </div>
    </div>
</div>
