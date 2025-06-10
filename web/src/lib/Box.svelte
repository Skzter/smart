<script lang="ts">
    let {msg, name} = $props();
    let text_align = $state("");
    let bg_col = $state("");
    let padding = $state("");
    let width = $state("");
    let direction = $state("");
    if (name == "User") {
        text_align = "text-end";
        bg_col = "bg-sky-300";
        padding = "p-2";
        width = "w-fit";
        direction = "justify-end";
    } else if (name === "Bot") {
        text_align = "text-start";
        bg_col = "bg-gray-200";
        padding = "p-2";
        width = "w-fit";
        direction = "justify-start";
    }
</script>

<div class = "flex {direction}">
    <div class="{bg_col} {width} {direction} border-2 border-black border-solid rounded-xl">
        <h1 class="{text_align} {padding} text-xl">{name}</h1>
        <p class="{text_align} {padding}">{msg}</p>
    </div>
</div>


