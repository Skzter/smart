<script lang="ts">
    import svelteLogo from "./assets/svelte.svg";
    import viteLogo from "/vite.svg";
    import Prompt from "./lib/Prompt.svelte";
    import Box from "./lib/Box.svelte";
    let prompt = $state('');
    let convo = $state([]);
    function onsubmit(){
	convo.push({ q: prompt, a: 42});
	console.log(convo);
	prompt = '';
    }
</script>


<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2">
	{#each convo as c, i}
	    <Box msg={c.q} name={"User"} />
	    <Box msg={c.a} name={"Bot"} />
	{/each}
	<Prompt bind:input={prompt} {onsubmit} />
    </div>
</div>
