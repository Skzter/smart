<script lang="ts">
    import { Toaster } from "svelte-sonner";
    import { CheckCircle2, XCircle } from "@lucide/svelte";
</script>

<Toaster
    position="top-right"
    icons={{
        success: CheckCircle2,
        error: XCircle,
    }}
    toastOptions={{
        classes: {
            toast: "!border-none",
            success: "!bg-emerald-500 !text-white",
            error: "!bg-rose-700 !text-white",
        },
        descriptionClass: "!text-white",
    }}
    duration={3000}
/>
