<script>
    /** @type {{ tick: number, lines: string[] }[]} */
    export let eventLog = [];
</script>


<div class="overflow-scroll max-h-[calc(100vh-20rem)] shadow-xl p-2 rounded-md">
    {#each [...eventLog].reverse() as tick}
        <div class="row">
            <p>{tick.tick}</p>
            {#each tick.lines as line}
                <p>{line}</p>
            {/each}
        </div>
    {/each}
</div>


<style>

    .row {
        padding: 0.5rem;
    }

    .row:nth-child(odd) {
        background-color: white;
    }

    .row:nth-child(even) {
        background-color: #f3f4f6; /* light grey */
    }

</style>