<script lang="ts">
    import { Calendar } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import { slide } from "svelte/transition";
    import { RangeCalendar } from "$lib/components/ui/range-calendar";
    import type { DateRange } from "bits-ui";
    import { CalendarDate } from "@internationalized/date";

    let showCalendar = $state(false);
    let value = $state<DateRange | undefined>(undefined);

    const today = new Date();
    const placeholder = new CalendarDate(
        today.getFullYear(),
        today.getMonth() + 1,
        today.getDate(),
    );

    function toggleCalendar() {
        showCalendar = !showCalendar;
    }

    $effect(() => {
        if (value) {
            console.log("Selected range:", value);
        }
    });
</script>

<div class="w-full">
    <Button
        variant="ghost"
        class="w-full justify-start gap-2 h-10 bg-muted hover:bg-muted/80"
        onclick={toggleCalendar}
    >
        <Calendar class="h-4 w-4" />
        <span>Time Filter</span>
    </Button>

    {#if showCalendar}
        <div
            transition:slide={{ duration: 300 }}
            class="w-full bg-background border rounded-lg shadow-sm mt-2 overflow-hidden flex items-center justify-center"
        >
            <div class="scale-[0.85] -my-7">
                <RangeCalendar bind:value {placeholder} />
            </div>
        </div>
    {/if}
</div>
