<script lang="ts">
  import type { ExampleSource } from "$lib/models/ExampleData";

  export let source: ExampleSource;

  $: domain = (new URL(source.url)).hostname
</script>

<div class="source">
  <span class="name"><a href={source.url}>{source.title}</a></span>
  <span class="link">({domain}, {source.date})</span>
  <span class="link">— <a href="/search?q=src:{encodeURIComponent(source.id)}">View all from source</span>
</div>

<style lang="sass">
  div.source
    font-size: 0.75em
    margin-bottom: 0.25em

    a
      color: #789
      text-decoration: none

      &:hover
        color: #abc
</style>