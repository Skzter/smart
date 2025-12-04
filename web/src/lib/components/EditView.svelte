<script lang="ts">
    import Code from "./Code.svelte";

    let {
        code,
    }: {
        code: string;
    } = $props();
</script>

<div class="w-full h-full p-6 overflow-auto">
     <Code code={code} />
</div>
