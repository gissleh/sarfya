<script lang="ts">
  let inputElem: HTMLInputElement

  export let query = "";
</script>

<form on:submit|preventDefault><input bind:this={inputElem} bind:value={query} /></form>

<style lang="sass">
  input
    display: block
    width: 600ch
    max-width: 92%
    background: none
    padding: 0.5em
    font-size: 1.5em
    color: #789
    outline: none
    border: 1px solid
    border-radius: 1em
    margin: 1em 0

    &:focus
      color: #abc
</style>