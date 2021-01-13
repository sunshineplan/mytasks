<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let showCompleted = false;
  export let selected = 0;
  export let incompleteTasks: Task[] = [];
  export let completedTasks: Task[] = [];

  const complete = async (id: number) => {
    $loading++;
    const resp = await post("/task/complete/" + id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count--;
        index = incompleteTasks.findIndex((task) => task.id === id);
        $tasks[$current.list].completed = [
          {
            id: json.id,
            task: incompleteTasks[index].task,
            created: new Date().toLocaleString(),
          },
          ...completedTasks,
        ];
        incompleteTasks.splice(index, 1);
        dispatch("refresh");
        return;
      }
    }
    await fire("Error", "Error", "error");
  };

  const del = async (id: number) => {
    if (await confirm("This task")) {
      $loading++;
      const resp = await post("/task/delete/" + id);
      $loading--;
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        let index = incompleteTasks.findIndex((task) => task.id === id);
        incompleteTasks.splice(index, 1);
        index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count--;
        dispatch("refresh");
      }
    }
  };

  const handleKeydown = (event: KeyboardEvent, id: number) => {
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      dispatch("edit", { id, task: (event.target as HTMLElement).innerText });
      selected = 0;
    }
  };
  const handleClick = (event: MouseEvent, id: number) => {
    if (selected !== id) {
      const target = event.target as HTMLElement;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
      const selectedTarget = document.querySelector(".selected>.task");
      if (selectedTarget)
        dispatch("edit", {
          id: selected,
          task: (selectedTarget as HTMLElement).innerText,
        });
      selected = id;
    }
  };
</script>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }

  li {
    display: inline-flex;
  }

  .created {
    padding: 0.75rem 0;
    color: #5f6368;
    width: 80px;
    text-align: right;
  }

  .list-group-item:hover {
    box-shadow: 0 1px 2px 0 rgba(60, 64, 67, 0.302),
      0 1px 3px 1px rgba(60, 64, 67, 0.149);
    outline: 0;
    z-index: 2000;
  }
</style>

<ul
  class="list-group list-group-flush"
  style={showCompleted ? 'height:calc(50% - 85px)' : 'height:calc(100% - 170px)'}
  id="tasks">
  {#each incompleteTasks as task (task.id)}
    <li class="list-group-item" class:selected={task.id === selected}>
      <i class="check" on:click={async () => await complete(task.id)} />
      <span
        class="task"
        contenteditable={task.id === selected}
        on:keydown={(e) => handleKeydown(e, task.id)}
        on:click={(e) => handleClick(e, task.id)}>{task.task}</span>
      <span
        class="created">{new Date(task.created.replace('Z', '')).toLocaleDateString()}</span>
      <i
        style="visibility:{task.id === selected ? 'visible' : 'hidden'}"
        class="operation delete"
        on:click={async () => await del(task.id)}>delete</i>
    </li>
  {/each}
</ul>
