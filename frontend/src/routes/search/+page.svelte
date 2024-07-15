<script lang="ts">
  import { goto, invalidateAll } from "$app/navigation";
  import Example from "$lib/components/Example.svelte";
  import ExampleEditor from "$lib/components/ExampleEditor.svelte";
  import ExampleListHeader from "$lib/components/ExampleListHeader.svelte";
  import SearchBox from "$lib/components/SearchBox.svelte";
  import type { ExampleSet, ExampleSource } from "$lib/models/ExampleData";

  function reload() {
    if (query) {
      goto(`/search?q=${encodeURIComponent(query)}`)
    } else {
      goto("/")
    }
  }

  export let data: {exampleSets: ExampleSet[], query: string, error: string}

  $: query = data.query;
  $: querySource = data.query.split(/,+!/g).find(s => s.trim().startsWith("src:"))
  $: querySourceData = querySource ? ({
    id: querySource.split("src:").pop()!,
    date: new Date(
      querySource.split("-")
        .filter(v => !Number.isNaN(parseInt(v)))
        .slice(0,3)
        .join("-") + "T00:00:00Z"
    ).toISOString().slice(0, 10) as `${number}-${number}-${number}`,
  }) : ({});
  $: console.log(querySourceData);
</script>

<SearchBox on:submit={reload} bind:query={query} />
<div class="error">{data.error}</div>

{#each data.exampleSets as exampleSet, i}
  <ExampleListHeader hideHeader={data.exampleSets.length === 1} entries={exampleSet.entries||[]} />
  {#each exampleSet.examples as example (example.id + "/" + i)}
    <Example value={example} />
  {/each}
{/each}

{#if import.meta.env.VITE_ENABLE_EDITOR === "true"}
  {#if !!querySource && data.exampleSets[0]?.examples.length > 0}
    <ExampleEditor on:saved={() => invalidateAll()} id={data.exampleSets[0].examples[0].id} another />
  {:else}
    <ExampleEditor on:saved={() => invalidateAll()} id="" defaultSource={querySourceData} />
  {/if}
{/if}

<style lang="sass">
  div.error
    color: #f73
    text-align: center
    margin-top: 1em
</style>