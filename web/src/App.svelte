<script lang="ts">
    import svelteLogo from "./assets/svelte.svg";
    import viteLogo from "/vite.svg";
    import Prompt from "./lib/Prompt.svelte";
    import Box from "./lib/Box.svelte";
    import {Spinner} from "flowbite-svelte";
    import {getResponse} from "./lib/Api.svelte.ts";
    let prompt = $state("");
    let convo = $state([]);

    // Params for Dummy API
    const params = {
        _locale: "de_DE",
        _quantity: 1,
    };

    async function onsubmit() {
        const UserQuestion = prompt;
        prompt = "";
        const answerPromise = getResponse(prompt, params);
        convo.push({question: UserQuestion, answer: answerPromise});
    }
    $inspect(convo);
    $inspect(prompt);
</script>

<div class="flex w-screen justify-center">
    <div class="flex flex-col w-8/10 gap-2 overflow-auto px-4 pt-4 pb-28">
        {#each convo as c}
            <Box msg={c.question} name={"User"} />
            {#await c.answer}
                <Spinner color="blue" />
            {:then answer}
                <Box msg={answer} name={"Bot"} />
            {:catch error}
                {console.log(error)}
                <Box msg={error} name={"Bot"} />
            {/await}
        {/each}
        <Prompt bind:input={prompt} {onsubmit} />
    </div>
</div>
