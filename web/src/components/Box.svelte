<script lang="ts">
    import { Button } from "flowbite-svelte";
    let { msg, name, showSave = false} = $props();
    
    // TODO: userid und conversationId noch dazu bekommen
    function saveTest(testcode: string) {
        // TODD: durch echte funktionalität mit api call ersetzen
        console.log("Saving testcode:", testcode);
    }
</script>

<div
    class="flex m-4"
    class:justify-end={name === "User"}
    class:justify-start={name === "Bot"}
>
    <div
        class="font-mono p-2.5 border-2 border-black border-solid rounded-xl"
        class:w-[75%]={name === "User"}
        class:w-fit={name === "Bot"}
        class:bg-sky-300={name === "User"}
        class:bg-gray-200={name === "Bot"}
    >
        <div class="flex items-start justify-between">
            <h1 class="tracking-wide uppercase font-bold text-xl">
                {name}
            </h1>
           <div class="ml-2">
                {#if showSave}
                    <Button color="purple"
                        onclick={() => saveTest(msg)}>
                        Speichern
                    </Button>
                {/if}
            </div>
        </div>
        <p class="font-sans whitespace-pre-wrap break-words">
            {msg}
        </p>
    </div>
</div>
