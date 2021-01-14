<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let show = false;
  export let completedTasks: Task[] = [];

  const expand = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (!target.classList.contains("delete")) show = !show;
  };

  const incomplete = async (id: number) => {
    $loading++;
    const resp = await post("/task/incomplete/" + id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count++;
        index = $tasks[$current.list].completed.findIndex(
          (task) => task.id === id
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

  const empty = async () => {
    if (await confirm("These completed tasks")) {
      $loading++;
      const resp = await post("/task/empty/" + $current.id);
      const json = await resp.json();
      $loading--;
      if (json.status) {
        $tasks[$current.list].completed = [];
        dispatch("refresh");
        show = false;
      } else await fire("Error", "Error", "error");
    }
  };
</script>

<div style="height: 100%">
  <div class="completed" on:click={expand}>
    <span>Completed ({completedTasks.length})</span>
    <i class="expand">{show ? "expand_more" : "expand_less"}</i>
    {#if show && completedTasks.length}
      <i class="expand delete" on:click={empty}>delete</i>
    {/if}
  </div>
  {#if show}
    <ul class="list-group list-group-flush" style="height:calc(50% - 85px)">
      {#each completedTasks as task (task.id)}
        <li class="list-group-item">
          <i class="uncheck" on:click={async () => await incomplete(task.id)}>
            done
          </i>
          <span class="task">{task.task}</span>
          <span class="created">
            {new Date(task.created.replace("Z", "")).toLocaleDateString()}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }

  li {
    display: inline-flex;
  }

  .completed {
    padding: 15px 20px;
    margin-top: 16px;
    font-weight: 500;
    color: #5f6368;
    cursor: pointer;
    background-color: rgba(0, 0, 0, 0.125);
  }

  .expand {
    font-family: "Material Icons";
    font-style: normal;
    font-size: 1.5rem;
    float: right;
    line-height: normal;
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
