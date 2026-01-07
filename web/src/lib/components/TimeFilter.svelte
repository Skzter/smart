<script lang="ts">
    import { slide } from "svelte/transition";
    import { RangeCalendar } from "$lib/components/ui/range-calendar";
    import { CalendarDate } from "@internationalized/date";
    import TimeFilterButton from "./TimeFilterButton.svelte";
    import { ChatDate } from "$lib/shared.svelte";

    let showCalendar = $state(false);
    const today = new Date();
    const placeholder = new CalendarDate(
        today.getFullYear(),
        today.getMonth() + 1,
        today.getDate(),
    );
</script>

<div class="w-full">
    <TimeFilterButton bind:showCalendar />
    {#if showCalendar}
        <div
            transition:slide={{ duration: 300 }}
            class="w-full bg-background border rounded-lg shadow-sm mt-2 overflow-hidden flex items-center justify-center"
        >
            <div class="scale-[0.85] -my-7">
                <RangeCalendar bind:value={ChatDate.Range} {placeholder} />
            </div>
        </div>
    {/if}
</div>
