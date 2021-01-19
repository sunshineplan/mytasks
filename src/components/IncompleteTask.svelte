<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let selected = 0;
  export let task: Task;
  let hover = false;

  const complete = async () => {
    $loading++;
    const resp = await post("/task/complete/" + task.id);
    $loading--;
    if (resp.ok) {
      const json = await resp.json();
      if (json.status && json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].incomplete--;
        $lists[index].completed++;
        index = $tasks[$current.list].incomplete.findIndex(
          (i) => task.id === i.id
        );
        $tasks[$current.list].completed = [
          {
            id: json.id,
            task: $tasks[$current.list].incomplete[index].task,
            created: new Date().toLocaleString(),
          },
          ...$tasks[$current.list].completed,
        ];
        $tasks[$current.list].incomplete.splice(index, 1);
        dispatch("refresh");
        return;
      }
    }
    await fire("Error", "Error", "error");
    dispatch("reload");
  };

  const del = async () => {
    if (await confirm("This task")) {
      $loading++;
      const resp = await post("/task/delete/" + task.id);
      $loading--;
      if (resp.ok) {
        let index = $tasks[$current.list].incomplete.findIndex(
          (i) => task.id === i.id
        );
        $tasks[$current.list].incomplete.splice(index, 1);
        index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].incomplete--;
        dispatch("refresh");
      } else {
        await fire("Error", await resp.text(), "error");
        dispatch("reload");
      }
    }
  };

  const handleKeydown = (event: KeyboardEvent) => {
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      const target = event.target as Element;
      target.textContent = (target.textContent as string).trim();
      dispatch("edit", {
        id: task.id,
        task: target.textContent,
      });
      selected = 0;
    }
  };
  const handleClick = (event: MouseEvent) => {
    let target = event.target as HTMLElement;
    if (
      selected !== task.id &&
      !target.classList.contains("complete") &&
      !target.classList.contains("delete")
    ) {
      target = (target.parentNode as Element).querySelector(
        ".task"
      ) as HTMLElement;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
      const selectedTask = document.querySelector(".selected>.task");
      if (selectedTask) {
        selectedTask.textContent = (selectedTask.textContent as string).trim();
        dispatch("edit", {
          id: selected,
          task: selectedTask.textContent,
        });
      }
      selected = task.id;
    }
  };
</script>

<li
  class="list-group-item"
  class:selected={task.id === selected}
  on:mouseenter={() => (hover = true)}
  on:mouseleave={() => (hover = false)}
  on:click={handleClick}
>
  <i class="icon complete" on:click={complete} />
  <span
    class="task"
    contenteditable={task.id === selected}
    on:keydown={handleKeydown}>
    {task.task}
  </span>
  <span class="created">
    {new Date(task.created.replace("Z", "")).toLocaleDateString()}
  </span>
  {#if task.id === selected}
    <i class="icon delete" on:click={del}>delete</i>
  {:else if hover}
    <i class="icon">edit</i>
  {/if}
</li>

<style>
  li {
    display: inline-flex;
  }

  .complete:before {
    content: "radio_button_unchecked";
  }

  .complete:hover:before {
    content: "done";
    color: #468dff;
  }

  .complete:hover {
    background-color: #e6ecf0;
    border-radius: 50%;
  }

  .created {
    padding: 0.75rem 0;
    color: #5f6368;
    width: 80px;
    text-align: right;
    cursor: default;
  }
</style>
