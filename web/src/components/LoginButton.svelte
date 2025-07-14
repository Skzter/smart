<script lang="ts">
    import { auth } from "../lib/authService";
</script>

<button on:click={auth.login}>Log in</button> 