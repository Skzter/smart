<script lang="ts">

import {Toaster} from "$lib/components/ui/sonner/index.ts";
import {CheckCircle2, XCircle} from "@lucide/svelte";
</script>

<Toaster
        position="top-right"
        icons={{
            success: CheckCircle2,
            error: XCircle,
        }}
        toastOptions={{
            style: 'background: rgb(239 68 68); color: white; border-color: rgb(220 38 38);',
            classNames: {
                success: '!bg-emerald-500 !text-white',
                error: '!bg-rose-500 !text-white',
            }
        }}
        duration={3000}
/>