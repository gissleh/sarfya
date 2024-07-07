<script lang="ts">
  export let value: string[]
  export let flag: string

  function onClick() {
    if (value.includes(flag)) {
      value = [...value].filter(v => v !== flag)
    } else {
      value = [...value, flag];
    }
  }
</script>

<button on:click={onClick} class="flag" class:checked={value.includes(flag)}>
  {flag}
</button>

<style lang="sass">
  button.flag
    padding: 0.125em 1ch
    border: none
    border-radius: 0.5em
    background-color: #035
    color: #1cf
    font-weight: 800
    margin-right: 1ch
    margin-bottom: 0.1em
    opacity: 0.5

    &.checked
      opacity: 1
</style>