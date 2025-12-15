<script lang="ts" module>
    import type { WithChildren } from "bits-ui";
    import type { HTMLAttributes } from "svelte/elements";

    export type WindowPropsWithoutHTML = WithChildren & {
        contentClass?: string;
        hideDots?: boolean;
    };

    export type WindowProps = HTMLAttributes<HTMLDivElement> &
        WindowPropsWithoutHTML;
</script>

<script lang="ts">
    import { cn } from "$lib/utils.js";

    let {
        children,
        class: className,
        contentClass,
        hideDots = false,
    }: WindowProps = $props();
</script>

<div
    class={cn(
        "border-border bg-background aspect-video w-full rounded-lg border",
        className,
    )}
>
    {#if !hideDots}
        <div class="border-b border-inherit p-4">
            <div class="flex items-center gap-2">
                <div class="size-2 rounded-full bg-[#ef4444]"></div>
                <div class="size-2 rounded-full bg-[#eab308]"></div>
                <div class="size-2 rounded-full bg-[#22c55e]"></div>
            </div>
        </div>
    {/if}
    <div class={cn("p-4", contentClass)}>
        {@render children?.()}
    </div>
</div>
