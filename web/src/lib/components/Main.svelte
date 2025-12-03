<script lang="ts">
    import Footer from "./Footer.svelte";
    import Chat from "./Chat.svelte";

    let {
        userId = $bindable(),
    }: {
        userId: string;
    } = $props();
    let isLoading = $state(false);
</script>

<Chat {userId} bind:isLoading />
<Footer {userId} bind:isLoading />
