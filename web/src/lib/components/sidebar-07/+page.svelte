<script lang="ts">
	import AppSidebar from "$lib/components/app-sidebar.svelte";
	import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
    import Send from "@lucide/svelte/icons/send";
    import Plus from "@lucide/svelte/icons/plus";
  	import { Button } from "$lib/components/ui/button/index.js";
  	import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
  	import * as InputGroup from "$lib/components/ui/input-group/index.js";
  	import * as Tooltip from "$lib/components/ui/tooltip/index.js";
    let voiceEnabled = false;
    let textarea: HTMLTextAreaElement;
</script>
<Sidebar.Provider>
    <AppSidebar />
    <Sidebar.Inset>
        <header class="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
            <div class="flex items-center gap-2 px-4">
                <Sidebar.Trigger class="-ms-1" />
                <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
                <Breadcrumb.Root />
            </div>
        </header>

        <div class="flex flex-1 flex-col gap-4 p-4 pt-0 h-full">
            <div class="flex-1 flex items-start justify-center min-h-0">
                <div class="w-full max-w-6xl bg-muted/50 h-full rounded-xl md:min-h-min overflow-auto p-6 min-h-0">
                    <!-- CHAT MESSAGES -->
                </div>
            </div>

            <div class="p-4 pt-0 sticky bottom-0 bg-background z-10">
                <div class="flex items-center justify-center w-full">
                    <div class="w-full max-w-6xl">
                        <ButtonGroup.Root class="[--radius:9999rem] w-full flex items-center justify-center gap-3">
                            <ButtonGroup.Root class="shrink-0">
                                <Button variant="outline" size="icon" class="w-10 h-10">
                                    <Plus />
                                </Button>
                            </ButtonGroup.Root>
                            <ButtonGroup.Root class="flex-1">
                                <div class="flex w-full items-center gap-2">
                                    <InputGroup.Root class="w-full">
                                        <InputGroup.Textarea
                                            bind:this={textarea}
                                            class="w-full resize-none min-h-11 max-h-36 overflow-y-auto"
                                            placeholder="Send a message..."
                                            rows={1}
                                            on:input={() => {
                                                if (textarea) {
                                                    textarea.style.height = 'auto';
                                                    textarea.style.height = Math.min(textarea.scrollHeight, 24 * 6) + 'px';
                                                }
                                            }}
                                        />
                                        <p class="text-muted-foreground text-sm">
                                    <script lang="ts">
                                        let textarea: HTMLTextAreaElement;
                                    </script>
                                    </InputGroup.Root>
                                    <Button variant="default" size="icon" class="w-10 h-10">
                                        <Send />
                                    </Button>
                                </div>
                            </ButtonGroup.Root>
                        </ButtonGroup.Root>
                    </div>
                </div>
            </div>
        </div>
    </Sidebar.Inset>
</Sidebar.Provider>


