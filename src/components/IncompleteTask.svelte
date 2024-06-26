<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, confirm, post, pasteText } from "../misc";
  import { current, lists, tasks } from "../task";

  const dispatch = createEventDispatcher();

  export let selected = "";
  export let task: Task;
  let hover = false;
  let composition = false;

  const complete = async () => {
    const resp = await post("/task/complete/" + task.id);
    if (resp.ok) {
      const json = await resp.json();
      if (json.status && json.id) {
        let index = $lists.findIndex((list) => list.list === $current.list);
        $lists[index].incomplete--;
        $lists[index].completed++;
        index = $tasks[$current.list].incomplete.findIndex(
          (i) => task.id === i.id,
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
      await fire("Error", "Error", "error");
      dispatch("reload");
    } else await fire("Error", await resp.text(), "error");
  };

  const del = async () => {
    if (await confirm("This task")) {
      const resp = await post("/task/delete/" + task.id);
      if (resp.ok) {
        const json = await resp.json();
        if (json.status) {
          let index = $tasks[$current.list].incomplete.findIndex(
            (i) => task.id === i.id,
          );
          $tasks[$current.list].incomplete.splice(index, 1);
          index = $lists.findIndex((list) => list.list === $current.list);
          $lists[index].incomplete--;
          dispatch("refresh");
        } else {
          await fire("Error", "Error", "error");
          dispatch("reload");
        }
      } else await fire("Error", await resp.text(), "error");
    }
  };

  const handleKeydown = (event: KeyboardEvent) => {
    if (composition) return;
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      const target = event.target as Element;
      target.textContent = target.textContent!.trim();
      dispatch("edit", {
        id: task.id,
        task: target.textContent,
      });
      selected = "";
    }
  };
  const handleClick = (event: MouseEvent) => {
    let target = event.target as HTMLElement;
    if (
      selected !== task.id &&
      !target.classList.contains("complete") &&
      !target.classList.contains("delete")
    ) {
      target = target.parentNode!.querySelector(".task")!;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection()!;
      sel.removeAllRanges();
      sel.addRange(range);
      const selectedTask = document.querySelector(".selected>.task");
      if (selectedTask) {
        selectedTask.textContent = selectedTask.textContent!.trim();
        if (selected)
          dispatch("edit", {
            id: selected,
            task: selectedTask.textContent,
          });
        else
          dispatch("add", {
            task: selectedTask.textContent,
          });
      }
      selected = task.id;
    }
  };
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<li
  class="list-group-item"
  class:selected={task.id === selected}
  on:mouseenter={() => (hover = true)}
  on:mouseleave={() => (hover = false)}
  on:click={handleClick}
>
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <i class="icon complete" on:click={complete} />
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <span
    class="task"
    contenteditable={task.id === selected}
    on:compositionstart={() => {
      composition = true;
    }}
    on:compositionend={() => {
      composition = false;
    }}
    on:keydown={handleKeydown}
    on:paste={pasteText}
  >
    {task.task}
  </span>
  <span class="created">
    {new Date(task.created.replace("Z", "")).toLocaleDateString()}
  </span>
  {#if task.id === selected}
    <!-- svelte-ignore a11y-no-static-element-interactions -->
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
    cursor: default;
  }
</style>
