<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import type { ExampleData } from "$lib/models/ExampleData";
  import ExampleEditor from "./ExampleEditor.svelte";
  import FlagList from "./FlagList.svelte";
  import Sentence from "./Sentence.svelte";
  import SourceLine from "./SourceLine.svelte";

  function onStartEdit() {
    if (import.meta.env.VITE_ENABLE_EDITOR) {
      editing = true;
    }
  }

  export let value: ExampleData

  let hovered: number[] | null = null;
  let editing: boolean = false;
</script>

{#if editing}
  <ExampleEditor id={value.id} 
    on:saved={() => { editing = false; invalidateAll() }} 
    on:cancel={() => { editing = false }} 
  />
{:else}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="example" on:dblclick={onStartEdit}>
    <FlagList flags={value.flags || []} />
    <SourceLine source={value.source} />
    <Sentence navi
      wordMap={value.wordMap} 
      dictionaryEntries={value.words}
      value={value.text} 
      spans={value.spans} 
      adjacents={value.translatedAdjacent.en || []}
      bind:hovered={hovered}
    />
    {#if !!value.translations.en}
      <Sentence 
        value={value.translations.en} 
        spans={value.translatedSpans.en}
        bind:hovered={hovered} 
      />
    {/if}
  </div>
{/if}

<style lang="sass">
  div.example
    margin-bottom: 2em
</style>