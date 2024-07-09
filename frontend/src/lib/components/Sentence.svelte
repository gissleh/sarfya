<script lang="ts">
  import type { DictionaryEntry } from "$lib/models/DictionaryEntry";
  import type { SentencePart } from "$lib/models/Sentence";

  function onHover(index: number, from?: number) {
    if (from != null && hovered !== value[from].ids) {
      return;
    }

    if (index >= 0) {
      hovered = value[index].ids || null;
    } else {
      hovered = null;
    }
  }

  function generateLink(part: SentencePart) {
    if (!part.ids || part.ids?.length === 0) {
      return void(0)
    }

    const entry = dictionaryEntries[part.ids[0]]?.[0];
    if (entry == null || !entry.id) {
      return void(0)
    }

    const word = entry.word.replace("+", "");

    return `/search?q=${encodeURIComponent(word)}:${encodeURIComponent(entry.id)}`
  }

  export let value: SentencePart[];
  export let spans: number[][] = [];
  export let adjacents: number[][] = [];
  export let navi: boolean = false;
  export let dictionaryEntries: {[id: number]: DictionaryEntry[]} = {};
  export let hovered: number[] | null = null;
  export let wordMap: {[id: number]: string} = {};

  function generateText(part: SentencePart) {
    let text = part.text;
    if (part.lp) {
      text = "(" + text;
    }
    if (part.rp) {
      text += ")";
    }

    return text;
  }

  $: adjacnetFlat = adjacents.flat();
  $: spansFlat = spans.flat();
</script>

<div class="sentence" class:navi>
  {#each value as part, i (i)}
    <!-- TODO: Make this less messy -->
    <!-- svelte-ignore a11y_mouse_events_have_key_events -->
    <a class="part" class:newline={part.newline}
      href={generateLink(part)}
      class:adjacent={adjacnetFlat.includes(i)}
      class:hover={!!hovered?.find(id => part.ids?.includes(id))}
      class:selected={spansFlat.includes(i)}
      on:touchmove={() => onHover(i)}
      on:mouseover={() => onHover(i)}
      on:mouseleave={() => onHover(-1, i)}
    >
      {generateText(part)}
    </a>
  {/each}
</div>

<style lang="sass">
  div.sentence
    color: #89a
    font-style: italic
    line-height: 1.25em
    width: 100%

    &.navi
      font-size: 1.25em
      font-weight: 600
      color: #abc
      font-style: normal
      margin-bottom: 0.25em

    a.part
      color: inherit
      text-decoration: none
      border-bottom: 1px solid rgba(0,0,0,0)

      &.hover
        background-color: #222233

      &.newline::before
        content: "\A"
        white-space: pre

      &.adjacent
        border-color: #fc1

      &.selected
        color: #fc1
        border-color: rgba(0,0,0,0)
</style>