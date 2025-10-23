<script lang="ts">
    let { msg, name } = $props();
</script>

<div
    class="flex m-4"
    class:justify-end={name === "User"}
    class:justify-start={name === "Bot"}
>
    <div
        class="font-mono w-full p-2.5 border-2 border-black border-solid rounded-xl"
        class:justify-end={name === "User"}
        class:justify-start={name === "Bot"}
        class:bg-sky-300={name === "User"}
        class:bg-gray-200={name === "Bot"}
    >
        <h1
            class="tracking-wide uppercase font-bold text-xl"
            class:text-end={name === "User"}
            class:text-start={name === "Bot"}
        >
            {name}
        </h1>
        <p
            class="font-sans wrap-break-word whitespace-pre-wrap"
            class:text-end={name === "User"}
            class:text-start={name === "Bot"}
        >
            {msg}
        </p>
    </div>
</div>
