<script lang="ts">
    let { msg, name } = $props();
    let text_align = "";
    let bg_col = "";
    let direction = "";

    if (name == "User") {
        text_align = "text-end";
        bg_col = "bg-sky-300";
        direction = "justify-end";
    } else if (name === "Bot") {
        text_align = "text-start";
        bg_col = "bg-gray-200";
        direction = "justify-start";
    }
</script>

<div class="flex {direction} m-4">
    <div
        class=" font-mono {bg_col} w-fit {direction} p-2.5 border-2 border-black border-solid rounded-xl"
    >
        <h1 class="{text_align} tracking-wide uppercase font-bold text-xl">
            {name}
        </h1>
        <p class="{text_align} font-sans whitespace-pre-wrap">{msg}</p>
    </div>
</div>
