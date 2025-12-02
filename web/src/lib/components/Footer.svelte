<script lang="ts">
    import Button from "./ui/button/button.svelte";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import * as InputGroup from "$lib/components/ui/input-group";
    import { Send } from "@lucide/svelte";

    let {
        input = $bindable(),
        isLoading = $bindable(),
    }: {
        input: string;
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
                        <InputGroup.Root class="w-full">
                            <InputGroup.Textarea
                                bind:value={input}
                                class="w-full resize-none min-h-11"
                                placeholder="Send a message..."
                                rows={1}
                                disabled={isLoading}
                            />
                        </InputGroup.Root>
                        <Button
                            variant="default"
                            size="icon"
                            class="w-10 h-10"
                            disabled={isLoading}
                        >
                            <Send />
                        </Button>
                    </div>
                </ButtonGroup.Root>
            </ButtonGroup.Root>
        </div>
    </div>
</div>
