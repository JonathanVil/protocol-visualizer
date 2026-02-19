<script>
    /** @typedef {import('$lib/types.js').Message} Message */

    /** @type {Message} */
    export let message;
</script>


<div class="pointer-events-auto relative whitespace-nowrap rounded-lg border border-white/10 bg-slate-900/90 px-2 py-1.5 text-[12px] leading-[1.2] text-white shadow-[0_8px_20px_rgba(0,0,0,0.35)]">
    <div>
        <div class="font-bold">{message.id}</div>
        <div class="text-xs text-gray-400">{message.source} → {message.destination}</div>
    </div>
</div>