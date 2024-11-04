<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { mytasks } from "../task.svelte";
  import IncompleteTask from "./IncompleteTask.svelte";

  let {
    showCompleted = $bindable(),
    selected = $bindable(),
  }: {
    showCompleted?: boolean;
    selected?: string;
  } = $props();

  onMount(() => {
    const sortable = new Sortable(document.querySelector("#tasks")!, {
      animation: 150,
      delay: 200,
      swapThreshold: 0.5,
      onUpdate: async (e) => {
        await mytasks.swapTask(
          mytasks.tasks.incomplete[e.oldIndex!],
          mytasks.tasks.incomplete[e.newIndex!],
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
  {#each mytasks.tasks.incomplete as task, i (task.id)}
    <IncompleteTask bind:selected bind:task={mytasks.tasks.incomplete[i]} />
  {/each}
</ul>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }
</style>
