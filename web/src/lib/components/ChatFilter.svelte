<script lang="ts">
    import { Funnel } from "@lucide/svelte";

    import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    
    let sortBy = $state("recent");
    let timeFilter = $state("all");
</script>

<DropdownMenu.Root>
    <DropdownMenu.Trigger>
        {#snippet child({ props })}
            <Button
                {...props}
                variant="ghost"
                class="w-full justify-start gap-2 h-10 bg-muted hover:bg-muted/80"
            >
                <Funnel class="h-4 w-4" />
                <p>Chat Filter</p>
            </Button>
        {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="w-56">
        <DropdownMenu.Group>
            <DropdownMenu.Label>Sortieren nach</DropdownMenu.Label>
            <DropdownMenu.Separator />
            <DropdownMenu.RadioGroup bind:value={sortBy}>
                <DropdownMenu.RadioItem value="recent">
                    Zuletzt genutzt
                </DropdownMenu.RadioItem>
                <DropdownMenu.RadioItem value="created">
                    Erstellt am
                </DropdownMenu.RadioItem>
            </DropdownMenu.RadioGroup>
        </DropdownMenu.Group>
        <DropdownMenu.Group>
            <DropdownMenu.Label>Zeitraum</DropdownMenu.Label>
            <DropdownMenu.Separator />
            <DropdownMenu.RadioGroup bind:value={timeFilter}>
                <DropdownMenu.RadioItem value="all">
                    Alle Chats
                </DropdownMenu.RadioItem>
                <DropdownMenu.RadioItem value="week">
                    Letzte Woche
                </DropdownMenu.RadioItem>
                <DropdownMenu.RadioItem value="month">
                    Letzter Monat
                </DropdownMenu.RadioItem>
            </DropdownMenu.RadioGroup>
        </DropdownMenu.Group>
    </DropdownMenu.Content>
</DropdownMenu.Root>



