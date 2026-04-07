<script>
  import { onMount } from 'svelte';
  import {marked} from "marked";


  let source = '';

  onMount(async () => {
    const res = await fetch('/BASICS.md');
    source = await res.text();
  });
</script>

<div class="flex h-full flex-col">
    <div class="rounded-t-md border border-b-0 border-gray-300 bg-gray-100 px-3 py-2">
        <h1 class="text-lg font-medium">Guide</h1>
    </div>

    <div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-b-md bg-white border border-gray-300 prose">
        {@html marked.parse(source)}
    </div>
</div>