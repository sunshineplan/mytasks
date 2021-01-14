<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let task: Task;
  let hover = false;

  const revert = async () => {
    $loading++;
    const resp = await post("/completed/revert/" + task.id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count++;
        index = $tasks[$current.list].completed.findIndex(
          (i) => task.id === i.id
        );
        $tasks[$current.list].incomplete = [
          {
            id: json.id,
            task: $tasks[$current.list].completed[index].task,
            created: new Date().toLocaleString(),
          },
          ...$tasks[$current.list].incomplete,
        ];
        $tasks[$current.list].completed.splice(index, 1);
        dispatch("refresh");
        return;
      }
    }
    await fire("Error", "Error", "error");
  };

  const del = async () => {
    $loading++;
    const resp = await post("/completed/delete/" + task.id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      const index = $tasks[$current.list].completed.findIndex(
        (i) => task.id === i.id
      );
      $tasks[$current.list].completed.splice(index, 1);
      dispatch("refresh");
      return;
    }
    await fire("Error", "Error", "error");
  };
</script>

<li
  class="list-group-item"
  on:mouseenter={() => (hover = true)}
  on:mouseleave={() => (hover = false)}
>
  <i class="icon revert" on:click={revert}>done</i>
  <span class="task">{task.task}</span>
  <span class="created">
    {new Date(task.created.replace("Z", "")).toLocaleDateString()}
  </span>
  {#if hover}
    <i class="icon delete" on:click={del}>delete</i>
  {/if}
</li>

<style>
  li {
    display: inline-flex;
  }

  .revert {
    content: "done";
    color: #468dff;
  }

  .revert:hover {
    background-color: #e6ecf0;
    border-radius: 50%;
  }

  .task {
    padding: 0.75rem 0;
    width: calc(100% - 176px);
    text-decoration: line-through;
  }

  .created {
    padding: 0.75rem 0;
    color: #5f6368;
    width: 80px;
    text-align: right;
  }
</style>
