<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { BootstrapButtons, post } from "../misc";
  import { current, loading, tasks, Task } from "../stores";

  let currentTasks: Task[] = [];

  const getTasks = async () => {
    if (!$tasks.hasOwnProperty($current.list)) {
      $loading++;
      const resp = await post("/task/get", { list: $current.id });
      $tasks[$current.list] = await resp.json();
      $loading--;
    }
    currentTasks = $tasks[$current.list];
  };

  $: $current && getTasks();

  onMount(() => {
    const sortable = new Sortable(
      document.querySelector("#mytasks") as HTMLElement,
      {
        animation: 150,
        delay: 100,
        swapThreshold: 0.5,
        onUpdate,
      }
    );
    return () => sortable.destroy();
  });

  const onUpdate = async (event: Sortable.SortableEvent) => {
    const resp = await post("/reorder", {
      old: currentTasks[event.oldIndex as number].id,
      new: currentTasks[event.newIndex as number].id,
    });
    if ((await resp.text()) == "1") {
      const task = currentTasks[event.oldIndex as number];
      currentTasks.splice(event.oldIndex as number, 1);
      currentTasks.splice(event.newIndex as number, 0, task);
    } else await BootstrapButtons.fire("Error", "Failed to reorder.", "error");
  };

  const editList = () => {
    console.log("/list/edit");
  };
  const add = () => {
    console.log("/task/add");
  };
</script>

<style>
  .icon {
    color: #007bff !important;
    cursor: pointer;
  }

  .icon:hover {
    color: #0056b3 !important;
  }

  .h3 {
    cursor: default;
  }

  .edit {
    font-size: 18px;
  }

  ul {
    height: calc(100% - 100px);
    padding: 0 10px;
    cursor: default;
  }
</style>

<svelte:head>
  <title>{$current.list} - My Tasks</title>
</svelte:head>

<div style="height: 100%">
  <header style="padding-left: 20px">
    <div style="height: 50px">
      <span class="h3">{$current.list}</span>
      <span class="btn icon" on:click={editList}>
        <i class="material-icons edit">edit</i>
      </span>
    </div>
    <button class="btn btn-primary" on:click={add}>Add Task</button>
  </header>
  <ul class="list-group list-group-flush" id="mytasks">
    {#each currentTasks as task (task.id)}
      <li class="list-group-item">{task.task}</li>
    {/each}
  </ul>
</div>
