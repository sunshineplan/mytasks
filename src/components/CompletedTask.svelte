<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, post, confirm } from "../misc";
  import { current, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let task: Task;
  let hover = false;

  const revert = async () => {
    const resp = await post("/completed/revert/" + task.id);
    if (resp.ok) {
      const json = await resp.json();
      if (json.status && json.id) {
        let index = $lists.findIndex((list) => list.list === $current.list);
        $lists[index].incomplete++;
        $lists[index].completed--;
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
      await fire("Error", "Error", "error");
      dispatch("reload");
    } else await fire("Error", await resp.text(), "error");
  };

  const del = async () => {
    if (await confirm("This completed task")) {
      const resp = await post("/completed/delete/" + task.id);
      if (resp.ok) {
        const json = await resp.json();
        if (json.status) {
          const index = $tasks[$current.list].completed.findIndex(
            (i) => task.id === i.id
          );
          $tasks[$current.list].completed.splice(index, 1);
          dispatch("refresh");
          return;
        }
        await fire("Error", "Error", "error");
        dispatch("reload");
      } else await fire("Error", await resp.text(), "error");
    }
  };
</script>

<li
  class="list-group-item"
  on:mouseenter={() => (hover = true)}
  on:mouseleave={() => (hover = false)}
>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <i class="icon revert" on:click={revert}>done</i>
  <span class="task">{task.task}</span>
  <span class="created">
    {new Date(task.created.replace("Z", "")).toLocaleDateString()}
  </span>
  {#if hover}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
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
    text-decoration: line-through;
  }
</style>
