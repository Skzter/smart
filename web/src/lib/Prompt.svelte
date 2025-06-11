<script lang="ts">
    import {Button, Input} from "flowbite-svelte";
    let {input = $bindable(""), onsubmit} = $props();
</script>

<div class="fixed bottom-0 left-0 w-full bg-white p-4 border-t flex">
    <Input
        type="text"
        bind:value={input}
        placeholder="Prompt"
        required
        class="flex-1 mr-2"
    />
    <Button color="purple" onclick={onsubmit}>Send</Button>
</div>
