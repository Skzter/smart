<script lang="ts">
    import { Editor, Environment, ModelManager } from "$lib/editor.svelte";
    import { editor } from "monaco-editor";
    import { onMount, onDestroy } from "svelte";

    let {
        value = $bindable(),
        class: className = "",
        options,
    }: {
        value?: string;
        class?: string;
        options?: {
            maxHeight?: number;
            useTextHeight?: boolean;
        };
    } = $props();

    let model = ModelManager.getModel("id");
    if (!model) {
        model = ModelManager.createModel("id", value ?? "", () => {
            if (model) {
                value = model.getMonacoModel().getValue();
            }
        });
    }

    let codeEditor: editor.IStandaloneCodeEditor | undefined = $state();
    let editorContainer: HTMLElement = document.createElement("div");

    onMount(async () => {
        console.log("mount");
        Environment();
        codeEditor = await Editor(editorContainer, {
            theme: "vs-dark",
            wordWrap: "on",
            allowOverflow: true,
            minimap: {
                enabled: false,
            },
            lineHeight: 1.5,
            scrollBeyondLastLine: false,
        });

        codeEditor.setModel(model.getMonacoModel());
    });

    $effect(() => {
        if (model) {
            const monacoModel = model.getMonacoModel();

            if (monacoModel.getValue() !== (value ?? "")) {
                monacoModel.setValue(value ?? "");
            }

            if (codeEditor) {
                let height: number;
                if (options?.useTextHeight) {
                    height = codeEditor.getContentHeight();
                } else {
                    height = editorContainer.clientHeight;
                }
                codeEditor.layout({
                    height: Math.min(height, options?.maxHeight ?? 10000),
                    width: editorContainer.clientWidth,
                });
            }
        }
    });

    onDestroy(() => {
        codeEditor?.dispose();
    });
</script>

<div bind:this={editorContainer} class={className}></div>

<style>
    :global(.monaco-editor) {
        height: 100%;
    }
</style>
