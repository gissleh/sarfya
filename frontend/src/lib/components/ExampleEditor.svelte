<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { BackendClient } from "$lib/backend/client";
  import { blankInput, type Input } from "$lib/models/Input";
  import type { ParsedSentence } from "$lib/models/ParsedSentence";
  import { onMount } from "svelte";
  import WordPicker from "./WordPicker.svelte";
  import FlagToggle from './FlagToggle.svelte';
  import type { ExampleSource } from '$lib/models/ExampleData';

  const dispatch = createEventDispatcher();

  let client: BackendClient | null = null;
  let loadedId: string = "";
  let loadedParse: string = "";
  let parsed: ParsedSentence | null = null;
  
  export let id: string;
  export let another: boolean = false;
  export let defaultSource: Partial<ExampleSource> = {};

  let loading = false;
  let input: Input = blankInput("", defaultSource);
  let error = "";

  function onSave() {
    if (loading || !client) {
      return
    }
    error = "";
    loading = true;

    client.postExample(input)
      .then(() => {
        dispatch("saved")

        if (another) {
          parsed = null
          loadedId = "_"
        } else {
          input = {
            ...blankInput(""),
            source: {...input.source}
          };
        }
      }).catch(err => {
        error = err.error || String(err)
      }).finally(() => {
        loading = false;
      })
  }

  function onDelete() {
    if (loading || !client) {
      return
    }
    loading = true;

    client.deleteExample(id)
      .then(() => {
        dispatch("saved")

        if (another) {
          parsed = null
          loadedId = "_"
        }
      }).catch(err => {
        error = err.error || String(err)
      }).finally(() => {
        loading = false;
      })
  }

  onMount(() => { client = new BackendClient(fetch) })

  $: if (loadedId !== id && client !== null && !loading) {
    if (id !== "") {
      loading = true;
      client.getInput(id).then(newInput => {
        loadedId = newInput.id || "";
        if (another) {
          input = blankInput("", defaultSource);
          input.source = {...newInput.source};
        } else {
          input = {...newInput};
          if (!input.translations.en) {
            input.translations.en = "";
          }
          input.flags = input.flags || [];
        }
      }).finally(() => {
        loading = false;
      })
    } else {
      loadedId = "";
      input = blankInput("", defaultSource);
    }
  }

  $: if (loadedParse !== input.text && client !== null && !loading) {
    const nextText = input.text;
    loading = true;
    
    client.parseSentence(input.text).then(newParsed => {
      parsed = newParsed
      if (input.lookupFilter == null) {
        input.lookupFilter = {};
      }
      for (const id of Object.keys(parsed.entries)) {
        if (!input.lookupFilter[Number(id)] && parsed.entries[Number(id)].length) {
          input.lookupFilter[Number(id)] = ""
        }
      }
      for (const id of Object.keys(input.lookupFilter)) {
        if (parsed.entries[Number(id)]?.length < 2) {
          input.lookupFilter[Number(id)] = ""
        }
      }
    }).catch(err => {
      // do nothing
    }).finally(() => {
      loadedParse = nextText;
      loading = false;
    })
  }
</script>

<div class="editor">
  <label for="text">Na'vi Text</label>
  <textarea name="text" bind:value={input.text}></textarea>
  {#if error}<div class="error">{error}</div>{/if}
  {#each Object.keys(input.translations) as lang}
    <label for="text_{lang}">Translation ({lang})</label>
    <textarea name="text_{lang}" bind:value={input.translations[lang]}></textarea>
  {/each}
  <label>Flags</label>
  <div class="flag-list">
    <FlagToggle bind:value={input.flags} flag="poetry" />
    <FlagToggle bind:value={input.flags} flag="non_canon" />
    <FlagToggle bind:value={input.flags} flag="user_translation" />
    <FlagToggle bind:value={input.flags} flag="reef_dialect" />
    <FlagToggle bind:value={input.flags} flag="proverb" />
    <FlagToggle bind:value={input.flags} flag="slang" />
    <FlagToggle bind:value={input.flags} flag="formal" />
    <FlagToggle bind:value={input.flags} flag="syntax" />
    <FlagToggle bind:value={input.flags} flag="clipped" />
  </div>
  <div class="split">
    <div class="left">
      {#if parsed != null}
        {#each Object.keys(parsed.wordMap).map(Number) as id}
          {#if parsed.entries[id]?.length > 1}
            <label for="source.id">Select Word ({parsed.wordMap[id]})</label>
            <WordPicker options={parsed.entries[id]} bind:value={input.lookupFilter[id]} />
          {/if}
          {#if parsed.entries[id]?.length === 0}
            <div class="missing">{parsed.wordMap[id]} matched no entries!</div>
          {/if}
        {/each}
      {/if}
    </div>
    <div class="right">
      <label for="source.id">Source ID</label>
      <input disabled={loading} type="text" name="source.id" bind:value={input.source.id} />
      <label for="source.date">Source Date</label>
      <input disabled={loading} type="date" name="source.date" bind:value={input.source.date} />
      <label for="source.url">Source URL</label>
      <input disabled={loading} type="text" name="source.url" bind:value={input.source.url} />
      <label for="source.title">Source Title</label>
      <input disabled={loading} type="text" name="source.title" bind:value={input.source.title} />
      <label for="source.author">Source Author</label>
      <input disabled={loading} type="text" name="source.author" bind:value={input.source.author} />
    </div>
  </div>
  <button on:click={onSave}>{id === "" || another ? "Add Example" : "Save Example"}</button>
  {#if id !== ""}
    <button on:click={() => dispatch("cancel")}>Cancel</button>
    <button on:dblclick={onDelete}>Delete</button>
  {/if}
</div>

<style lang="sass">
  div.editor
    min-height: 3em
    margin-bottom: 1em
    padding: 0.5em 1ch
    border-top: 1px solid #456
    border-bottom: 1px solid #456

    div.split
      display: flex
      div.left, div.right
        flex: 100

  label
    display: block
    font-size: 0.75em

  div.missing
    color: #f73
    font-size: 0.75em

  div.error
    color: #f73
    margin-bottom: 1em
    font-size: 0.9em
    text-align: center

  button
    outline: none
    border: none
    background: none
    color: #789
    cursor: pointer

    &:hover
      color: #abc

  input, textarea
    font-family: Arial, Helvetica, sans-serif
    outline: none
    border: none
    background: none
    color: #789
    border: 1px solid
    border-radius: 0.25em
    padding: 0.2em 1ch
    width: 90%
    margin-bottom: 0.5em

    &:focus
      color: #abc

  textarea
    width: 94.75%
    resize: vertical
    min-height: 2em
</style>