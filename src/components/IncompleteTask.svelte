<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let selected = 0;
  export let task: Task;

  const complete = async (id: number) => {
    $loading++;
    const resp = await post("/task/complete/" + id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count--;
        index = $tasks[$current.list].incomplete.findIndex(
          (task) => task.id === id
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
  };

  const del = async (id: number) => {
    if (await confirm("This task")) {
      $loading++;
      const resp = await post("/task/delete/" + id);
      $loading--;
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        let index = $tasks[$current.list].incomplete.findIndex(
          (task) => task.id === id
        );
        $tasks[$current.list].incomplete.splice(index, 1);
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

<li class="list-group-item" class:selected={task.id === selected}>
  <i class="icon check" on:click={async () => await complete(task.id)} />
  <span
    class="task"
    contenteditable={task.id === selected}
    on:keydown={(e) => handleKeydown(e, task.id)}
    on:click={(e) => handleClick(e, task.id)}>
    {task.task}
  </span>
  <span class="created">
    {new Date(task.created.replace("Z", "")).toLocaleDateString()}
  </span>
  <i
    style="visibility:{task.id === selected ? 'visible' : 'hidden'}"
    class="icon delete"
    on:click={async () => await del(task.id)}>delete</i
  >
</li>

<style>
  li {
    display: inline-flex;
  }

  .check:before {
    content: "radio_button_unchecked";
  }

  .check:hover:before {
    content: "done";
    color: #468dff;
  }

  .check:hover {
    background-color: #e6ecf0;
    border-radius: 50%;
  }

  .created {
    padding: 0.75rem 0;
    color: #5f6368;
    width: 80px;
    text-align: right;
  }
</style>
