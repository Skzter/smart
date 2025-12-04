<script lang="ts">
  export let activeTab = 'run';

  function handleTabClick(tab: string) {
    activeTab = tab;
  }
</script>

<div class="flex border-b">
  <button 
    class="px-4 py-3 text-sm font-medium border-b-2 {activeTab === 'edit' ? 'border-black' : 'border-transparent'} hover:border-gray-300"
    onclick={() => handleTabClick('edit')}
  >
    Edit
  </button>
  <button 
    class="px-4 py-3 text-sm font-medium border-b-2 {activeTab === 'run' ? 'border-black' : 'border-transparent'} hover:border-gray-300"
    onclick={() => handleTabClick('run')}
  >
    Run
  </button>
  <button 
    class="px-4 py-3 text-sm font-medium border-b-2 {activeTab === 'result' ? 'border-black' : 'border-transparent'} hover:border-gray-300"
    onclick={() => handleTabClick('result')}
  >
    Result
  </button>
</div>
