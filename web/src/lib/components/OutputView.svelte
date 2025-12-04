<script lang="ts">
    import * as Terminal from '$lib/components/ui/terminal';
</script>

<div class="flex flex-col border-r">
  <div class="px-4 py-2 border-b bg-muted/50 text-sm font-medium">
    Test Output
  </div>
  <div class="flex-1 overflow-auto">
    <Terminal.Root class="m-0 max-w-none h-full" delay={0} hideDots={true}>
      <Terminal.TypingAnimation>&gt; Test wird ausgeführt...</Terminal.TypingAnimation>
    </Terminal.Root>
  </div>
</div>