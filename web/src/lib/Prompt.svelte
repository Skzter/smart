<script lang="ts">
    import {Button, Input} from "flowbite-svelte";
    let { input = $bindable(''), onsubmit } = $props();
</script>

<div class="flex">
    <Input type="text" bind:value={input} placeholder="Prompt" required />
    <Button color="purple" onclick={onsubmit}>Send</Button>
</div>
