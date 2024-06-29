<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import IncompleteTask from "./IncompleteTask.svelte";
  import { tasks } from "../task";

  export let showCompleted = false;
  export let selected = "";

  onMount(() => {
    const sortable = new Sortable(document.querySelector("#tasks")!, {
      animation: 150,
      delay: 200,
      swapThreshold: 0.5,
      onUpdate: async (e) => {
        await tasks.swap(
          $tasks.incomplete[e.oldIndex!],
          $tasks.incomplete[e.newIndex!],
        );
      },
    });
    return () => sortable.destroy();
  });
</script>

<ul
  class="list-group list-group-flush"
  style={showCompleted
    ? "height:calc(50% - 85px)"
    : "height:calc(100% - 170px)"}
  id="tasks"
>
  {#each $tasks.incomplete as task (task.id)}
    <IncompleteTask bind:selected bind:task />
  {/each}
</ul>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }
</style>
