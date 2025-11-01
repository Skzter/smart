<script lang="ts">
    import { Button } from "flowbite-svelte";
    import { getTemplate } from "../lib/Api.ts";
    import Popup from "./Popup.svelte";
    let { input = $bindable(""), onclick } = $props();
    
    let showPopup = $state(false);
    let popupMessage = $state("");
    let popupTitle = $state("Error");

    function handleKeyPress(e: KeyboardEvent) {
        if (e.key === "Enter" && input.trim() && !e.shiftKey) {
            onclick();
            input = "";
            e.preventDefault();
        }
    }
    
    async function loadTemplate() {
        try {
            const res = await getTemplate("/template");

            if (!res.ok) throw { status: res.status, message: await res.text() };
            const data = await res.json();
            input = data.template;
        } catch (err) {
            const status = err.status;
            if (status === 404) popupMessage = "Template nicht gefunden (404)";
            else if (status >= 500) popupMessage = "Serverfehler (500)";
            else popupMessage = "Interner Server Error";
            popupTitle = "Template Error";
            showPopup = true;
        }
    }
</script>

<div class="flex flex-row w-screen items-center bg-white border-t gap-2 p-4">
    <textarea
        onkeydown={handleKeyPress}
        bind:value={input}
        placeholder="Prompt"
        required
        class="w-9/10 resize-none overflow-y-auto"
        rows={3}
    ></textarea>
    <Button
        color="purple"
        class="w-1/10 h-1/3"
        {onclick}
        disabled={!input.trim()}>Send</Button
    >
    <Button
        color="purple"
        class="w-1/10 h-1/3"
        onclick={loadTemplate}
    >Template</Button>
</div>

<Popup bind:isOpen={showPopup} message={popupMessage} title={popupTitle} />
