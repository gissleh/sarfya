<script lang="ts">
  import { goto } from "$app/navigation";
  import Example from "$lib/components/Example.svelte";
  import SearchBox from "$lib/components/SearchBox.svelte";
  import type { ExampleData } from "$lib/models/ExampleData";

  let query = "";

  function reload() {
    if (query) {
      goto(`/search?q=${encodeURIComponent(query)}`)
    }
  }

  const exampleSearches: {q: string, d: string}[] = [
    {q: "lì'u", d: "Search for the word \"lì'u\""},
    {q: "sreferängey", d: "Search for the word \"srefey\""},
    {q: "*:-ri", d: "Search for all uses of the topical suffix"},
    {q: "*:<äng>|<ei>", d: "Search for all uses of the mood infixes"},
    {q: "*:<eyk er>", d: "Search for all verbs using both <eyk> and <er>"},
    {q: "ke +> *:vtr.|vtrm.", d: "Search for negated transitive verbs"},
    {q: "tsun:vim.", d: "Search for the verb \"tsun\" without the noun"},
    {q: "sko:8740", d: "Search for \"sko\" using its ID to disambiguate it from \"tsko\""},
  ]
</script>

<SearchBox on:submit={reload} bind:query={query} />

<div class="help">
  <h2>Example Searches</h2>
  <ul>
    {#each exampleSearches as {q, d} (q)}
      <li><a href="/search?q={encodeURIComponent(q)}"><code>{q}</code></a>: {d}</li>
    {/each}
  </ul>
</div>