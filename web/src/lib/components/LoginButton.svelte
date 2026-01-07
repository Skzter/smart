<script lang="ts">
    import { auth } from "$lib/authService";
    import Button from "./ui/button/button.svelte";
</script>

<Button size="lg" onclick={auth.login}>Log in</Button>
