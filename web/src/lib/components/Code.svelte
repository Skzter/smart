<script lang="ts">
    import { tokenize } from "$lib/syntaxHighlighting";

    let {
        code = $bindable(),
    }: {
        code: string;
    } = $props();

    let tokens = $derived(tokenize(code));
</script>

<!-- pre preserves whitespace in them so this has to be like this -->
<pre
    class="font-sans text-base leading-relaxed whitespace-pre-wrap wrap-break-words m-1.5">{#each tokens as token}<span
            class="token-{token.type}">{token.value}</span
        >{/each}</pre>
