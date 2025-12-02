<script lang="ts">
    import Footer from "./Footer.svelte";
    import Chat from "./Chat.svelte";

    let isLoading = $state(false);
</script>

<Chat bind:isLoading />
<Footer bind:isLoading />
