<script lang="ts">
    let {msg, name} = $props();
    let text_align = $state("");
    let bg_col = $state("");
    let padding = $state("");

    if (name == "User") {
        text_align = "text-end";
        bg_col = "bg-sky-300";
        padding = "pr-5";
    } else if (name === "Bot") {
        text_align = "text-start";
        bg_col = "bg-gray-200";
        padding = "pl-5";
    }
</script>

<div class="{bg_col} border-2 border-black border-solid rounded-xl">
    <h1 class="{text_align} {padding} text-xl">{name}</h1>
    <p class="{text_align} {padding}">{msg}</p>
</div>
