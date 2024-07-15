<script lang="ts">
  import type { DictionaryEntry } from "$lib/models/DictionaryEntry";
  import DictWord from "./DictWord.svelte";

  export let entries: DictionaryEntry[]
  export let hideHeader: boolean = false;
</script>

<div class="header" class:empty={hideHeader || entries.length === 0}>
  {#each entries as entry}
    <DictWord value={entry} noModifiers />
  {/each}
</div>

<style lang="sass">
  div.header
    margin-top: 1.5em
    margin-bottom: 1em
    padding-bottom: 0.125em
    font-size: 2em
    border-bottom: 1px solid #789

    &.empty
      display: none
</style>