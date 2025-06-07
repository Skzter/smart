<script lang="ts">
    import svelteLogo from "./assets/svelte.svg";
    import viteLogo from "/vite.svg";
    import Prompt from "./lib/Prompt.svelte";
    import Box from "./lib/Box.svelte";
    import { getResponse } from "./lib/Api.svelte.ts";
    let prompt = $state('');
    let convo = $state([]);


    function onsubmit(){
	getResponse(prompt, convo);
	prompt = '';
    }
    $inspect(convo);
    $inspect(prompt);
</script>


<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2">
	{#each convo as c, i}
	    <Box msg={c.question} name={"User"} />
	    <Box msg={c.answer} name={"Bot"} />
	{/each}
	<Prompt bind:input={prompt} {onsubmit} />
    </div>
</div>
