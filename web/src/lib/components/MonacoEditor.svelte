<script lang="ts">
    import { onMount, onDestroy } from "svelte";

    let {
        value = $bindable(),
        language = "typescript",
        options = {},
    }: {
        value?: string;
        language?: string;
        options?: Record<string, unknown>;
    } = $props();

    let container = $state<HTMLElement | null>(null);
    let editor = $state(null);

    onMount(async () => {
        if (!container) return;

        const monaco = await import("monaco-editor");

        editor = monaco.editor.create(container as HTMLElement, {
            value: value ?? "",
            language,
            automaticLayout: true,
            minimap: { enabled: false },
            theme: "vs-dark",
            // enable word wrap by default; can be overridden via `options` prop
            wordWrap:
                (options && (options as Record<string, unknown>).wordWrap) ??
                "on",
            wrappingStrategy:
                (options &&
                    (options as Record<string, unknown>).wrappingStrategy) ??
                "advanced",
            ...options,
        });

        const model = editor.getModel();
        const disposable = model?.onDidChangeContent(() => {
            const v = editor.getValue();
            // update bindable value so parent bindings update
            value = v;
        });

        onDestroy(() => {
            disposable?.dispose();
            editor?.dispose();
        });
    });

    $effect(() => {
        if (!editor) return;
        if (value !== undefined && editor.getValue() !== value) {
            const pos = editor.getPosition();
            editor.setValue(value);
            if (pos) editor.setPosition(pos);
        }
    });
</script>

<div bind:this={container} class="h-full min-h-[200px] w-full"></div>

<style>
    :global(.monaco-editor) {
        height: 100%;
    }
</style>
