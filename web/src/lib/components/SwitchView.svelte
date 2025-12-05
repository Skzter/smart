<script lang="ts">
    import { Button } from "$lib/components/ui/button/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { CodeXml, Columns2, Monitor } from "@lucide/svelte";
    import { Separator } from "$lib/components/ui/separator/index.js";
    import ControlButtons from "./ControlButtons.svelte";

    let {
        view = $bindable(),
        onCloseClick,
    }: {
        view: string;
        onCloseClick: () => void;
    } = $props();
</script>

<div class="flex items-center gap-2">
    <ButtonGroup.Root>
        <Button
            variant={view === "code" ? "default" : "secondary"}
            size="sm"
            onclick={() => {
                view = "code";
            }}
        >
            <CodeXml class="h-4 w-4" />
        </Button>
        <ButtonGroup.Separator />
        <Button
            variant={view === "split" ? "default" : "secondary"}
            size="sm"
            onclick={() => {
                view = "split";
            }}
        >
            <Columns2 class="h-4 w-4" />
        </Button>
        <ButtonGroup.Separator />
        <Button
            variant={view === "browser" ? "default" : "secondary"}
            size="sm"
            onclick={() => {
                view = "browser";
            }}
        >
            <Monitor class="h-4 w-4" />
        </Button>
    </ButtonGroup.Root>

    <Separator orientation="vertical" class="h-6" />

    <ControlButtons {onCloseClick} />
</div>
