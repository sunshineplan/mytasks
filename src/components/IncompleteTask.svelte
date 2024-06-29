<script lang="ts">
  import { confirm, pasteText } from "../misc";
  import { tasks } from "../task";

  export let selected = "";
  export let task: Task;
  let hover = false;
  let composition = false;

  const del = async () => {
    if (await confirm("This task")) await tasks.delete(task);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (composition) return;
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      const target = event.target as Element;
      target.textContent = target.textContent!.trim();
      await tasks.save(<Task>{ id: task.id, task: target.textContent });
      selected = "";
    }
  };
  const handleClick = async (event: MouseEvent) => {
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
        let task = <Task>{ task: selectedTask.textContent };
        if (selected) task.id = selected;
        await tasks.save(task);
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
  <i class="icon complete" on:click={async () => tasks.complete(task)} />
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
  <span class="created">{new Date(task.created).toLocaleDateString()}</span>
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
