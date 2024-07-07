<script lang="ts">
  import type { DictionaryEntryWithFilter } from "$lib/models/ParsedSentence";
  import DictWord from "./DictWord.svelte";

  export let options: DictionaryEntryWithFilter[];
  export let value: string;

  function onToggle(opt: DictionaryEntryWithFilter) {
    if (!selected.includes(opt.filter)) {
      value = value + ";" + opt.filter;
    } else {
      value = selected.filter(f => f !== opt.filter).join(";");
    }
  }

  $: selected = value.split(";").map(f => f.trim());
  $: {
    const limitedFilter = value
      .split(";")
      .map(f => f.trim())
      .filter(f => options.find(o => o.filter === f))
      .join(";");

    if (limitedFilter != value) {
      value = limitedFilter;
    }
  }
</script>

<div class="selector">
  {#each options as option, i}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <button class="option" 
      class:selected={selected.includes(option.filter)}
      on:click|preventDefault={() => onToggle(option)}
    ><DictWord value={option} /></button>
  {/each}
</div>

<style lang="sass">
  div.selector
    display: flex
    flex-wrap: wrap
    margin: 0.25em 0 0.75em 0
    font-size: 0.70em

    button.option
      background-color: #2d2d3f
      border-radius: 0.5em
      padding: 0.125em 0.75ch
      margin-right: 0.25em
      border: 1px solid #2d2d3f

      &.selected
        border-color: #abc
</style>