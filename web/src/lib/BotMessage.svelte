<script lang="ts">
    let props = $props();
</script>

<div class=" p-3 mr-10 w-1/2">
    <p class="p-1">Bot answering:</p>
    <div class="p-3 border-2 border-double rounded-2xl bg-gray-200">
        <p class="font-mono">{props.message}</p>
    </div>
</div>
