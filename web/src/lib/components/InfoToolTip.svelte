<script lang="ts">
    let {text}: {text: string} = $props();
</script>

<span>
   <i class="fa fa-info-circle fa-1x" aria-label={text} title={text}></i>
</span>
