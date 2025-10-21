<script lang="ts">
    let { msg, name } = $props();
</script>

<div
    class="flex m-4"
    class:justify-end={name === "User"}
    class:justify-start={name === "Bot"}
>
    <div
        class="font-mono p-2.5 border-2 border-black border-solid rounded-xl break-words"
        class:w-[75%]={name === "User"}
        class:w-fit={name === "Bot"}
        class:bg-sky-300={name === "User"}
        class:bg-gray-200={name === "Bot"}
    >

        <h1 class="tracking-wide uppercase font-bold text-xl">
            {name}
        </h1>
        <p class="font-sans whitespace-pre-wrap break-words">
            {msg}
        </p>
    </div>
</div>
