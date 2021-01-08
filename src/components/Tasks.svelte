<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { fire, post } from "../misc";
  import { current, loading, tasks } from "../stores";
  import type { Task } from "../stores";
  import type { element } from "svelte/internal";

  let currentTasks: Task[] = [];
  let selected: number;

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
      list: $current.id,
      old: currentTasks[event.oldIndex as number].id,
      new: currentTasks[event.newIndex as number].id,
    });
    if ((await resp.text()) == "1") {
      const task = currentTasks[event.oldIndex as number];
      currentTasks.splice(event.oldIndex as number, 1);
      currentTasks.splice(event.newIndex as number, 0, task);
    } else await fire("Error", "Failed to reorder.", "error");
  };

  const editList = () => {
    console.log("/list/edit");
  };
  const add = () => {
    console.log("/task/add");
  };
  const edit = (event: KeyboardEvent, id: number) => {
    console.log(event);
    if (event.key == "Enter") event.preventDefault();
    console.log(id);
    console.log("/task/edit");
  };

  const handleClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (!target.classList.contains("task")) selected = 0;
    else {
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
    }
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
    cursor: default;
  }

  li > span {
    outline: 0;
  }

  .list-group-item:hover {
    box-shadow: 0 1px 2px 0 rgba(60, 64, 67, 0.302),
      0 1px 3px 1px rgba(60, 64, 67, 0.149);
    outline: 0;
    z-index: 2000;
  }

  .selected {
    cursor: text;
    border-bottom-width: 1px;
    border-color: #1a73e8;
    background-color: #f8f9fa;
  }

  .selected:hover {
    box-shadow: none;
  }
</style>

<svelte:head>
  <title>{$current.list} - My Tasks</title>
</svelte:head>

<svelte:window on:click={handleClick} />

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
      <li class="list-group-item" class:selected={task.id === selected}>
        <span
          class="task"
          contenteditable={task.id === selected}
          on:click={() => (selected = task.id)}
          on:keydown={(e) => edit(e, task.id)}>{task.task}</span>
      </li>
    {/each}
  </ul>
</div>
