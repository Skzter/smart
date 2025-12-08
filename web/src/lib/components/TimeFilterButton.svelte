<script lang="ts">
    import { Calendar } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    let {
        showCalendar = $bindable(),
    }: {
        showCalendar: boolean;
    } = $props();
</script>

<Button
    variant="ghost"
    class="w-full justify-start gap-2 h-10 bg-muted hover:bg-muted/80"
    onclick={() => (showCalendar = !showCalendar)}
>
    <Calendar class="h-4 w-4" />
    <p>Time Filter</p>
</Button>
