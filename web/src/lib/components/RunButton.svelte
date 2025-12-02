<script lang="ts">
    import { Play } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
</script>

<Button variant="ghost" size="sm" class="h-7 gap-1.5 px-2">
    <Play class="h-3.5 w-3.5" />
    <span class="text-xs">Ausführen</span>
</Button>
