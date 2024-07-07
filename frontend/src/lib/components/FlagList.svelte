<script lang="ts">
  export let flags: string[];
</script>

<div class="flags">
  {#each flags.sort() as flag}
    <div class="flag">{flag.split("_").map(t => t.charAt(0).toUpperCase() + t.slice(1)).join(" ")}</div>
  {/each}
</div>

<style lang="sass">
  div.flags
    display: flex
    font-size: 0.75em
  
    div.flag
      padding: 0.125em 1ch
      border-radius: 0.5em
      background-color: #035
      color: #1cf
      font-weight: 800
      margin-right: 1ch
      margin-bottom: 0.1em
</style>