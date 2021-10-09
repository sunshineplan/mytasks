<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount, createEventDispatcher } from "svelte";
  import IncompleteTask from "./IncompleteTask.svelte";
  import { fire, post } from "../misc";
  import { current } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let showCompleted = false;
  export let selected = "";
  export let incompleteTasks: Task[] = [];

  onMount(() => {
    const sortable = new Sortable(
      document.querySelector("#tasks") as HTMLElement,
      {
        animation: 150,
        delay: 200,
        swapThreshold: 0.5,
        onUpdate,
      }
    );
    return () => sortable.destroy();
  });

  const onUpdate = async (event: Sortable.SortableEvent) => {
    const resp = await post("/task/reorder", {
      list: $current.list,
      orig: incompleteTasks[event.oldIndex as number].id,
      dest: incompleteTasks[event.newIndex as number].id,
    });
    if (resp.ok) {
      if ((await resp.text()) == "1") {
        const task = incompleteTasks[event.oldIndex as number];
        incompleteTasks.splice(event.oldIndex as number, 1);
        incompleteTasks.splice(event.newIndex as number, 0, task);
      } else {
        await fire("Error", "Failed to reorder task", "error");
        dispatch("reload");
      }
    } else await fire("Error", await resp.text(), "error");
  };
</script>

<ul
  class="list-group list-group-flush"
  style={showCompleted
    ? "height:calc(50% - 85px)"
    : "height:calc(100% - 170px)"}
  id="tasks"
>
  {#each incompleteTasks as task (task.id)}
    <IncompleteTask bind:selected bind:task on:refresh on:edit on:reload />
  {/each}
</ul>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }
</style>
