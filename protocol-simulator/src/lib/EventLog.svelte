<script>
    /** @type {string[][]} */
    export let eventLog = [];
</script>


{#each [...eventLog].reverse() as tick}
    <div class="row">
        {#each tick as line}
            <p>{line}</p>
        {/each}
    </div>
{/each}


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