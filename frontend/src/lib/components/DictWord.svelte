<script lang="ts">
  import type { DictionaryEntry } from "$lib/models/DictionaryEntry";

  export let value: DictionaryEntry
  export let noModifiers: boolean = false;
</script>

<div class="dict-entry">
  {#if !noModifiers}
    {#if value.prefixes}
      <div class="prefix">{value.prefixes.map(p => `${p}-`).join("")}</div>
    {/if}
    {#if value.lenitions}
      <div class="prefix">{value.lenitions.join(" ")}</div>
    {/if}
  {/if}
  <div class="word">{value.word}</div>
  {#if !noModifiers}
    {#if value.infixes}
      <div class="infix">{"<"}{value.infixes.map(p => `${p}`).join(" ")}{">"}</div>
    {/if}
    {#if value.suffixes}
      <div class="suffix">{value.suffixes.map(p => `-${p}`).join("")}</div>
    {/if}
  {/if}
  <div class="pos">({value.pos})</div>
</div>

<style lang="sass">
  div.dict-entry
    display: flex

    > div.prefix
      padding-right: 0.5ch

    > div.infix, div.pos
      padding-left: 0.5ch

    > div
      color: #789

    > div.word
      color: #abc
</style>