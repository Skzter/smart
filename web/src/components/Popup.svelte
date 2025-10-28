<script lang="ts">
    import { fade, slide } from "svelte/transition";
    
    let { message = "", title = "Alert", type = "error", isOpen = $bindable(false) } = $props();
    
    function closePopup() {
        isOpen = false;
    }
    
    function handleEscapeKey(event: KeyboardEvent) {
        if (event.key === "Escape") {
            closePopup();
        }
    }
    
    $effect(() => {
        if (isOpen) {
            window.addEventListener("keydown", handleEscapeKey);
            return () => {
                window.removeEventListener("keydown", handleEscapeKey);
            };
        }
    });
</script>

{#if isOpen}
    <div class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50" transition:fade={{ duration: 200 }}>
        <div 
            class="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 border-2 border-gray-300"
            role="dialog"
            aria-modal="true"
            aria-labelledby="popup-title"
            transition:slide={{ axis: "y", duration: 200 }}
        >
            <div class="p-6">
                <div class="flex items-center justify-between mb-4">
                    <h3 
                        id="popup-title"
                        class="text-2xl font-bold uppercase tracking-wide"
                        class:text-red-600={type === "error"}
                        class:text-green-600={type === "success"}
                        class:text-blue-600={type === "info"}
                    >
                        {title}
                    </h3>
                    <button 
                        onclick={closePopup}
                        class="text-gray-400 hover:text-gray-600 transition-colors"
                        aria-label="Close"
                    >
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <p class="text-gray-700 mb-6 whitespace-pre-wrap">{message}</p>
                <div class="flex justify-end">
                    <button
                        onclick={closePopup}
                        class="px-6 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-semibold uppercase"
                    >
                        OK
                    </button>
                </div>
            </div>
        </div>
    </div>
{/if}

