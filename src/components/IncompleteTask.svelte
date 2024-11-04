<script lang="ts">
  import { confirm, pasteText } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let {
    selected = $bindable(),
    task = $bindable(),
  }: {
    selected?: string;
    task: Task;
  } = $props();

  let hover = $state(false);
  let composition = $state(false);

  const del = async () => {
    if (await confirm("This task")) await mytasks.deleteTask(task);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (composition) return;
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      const target = event.target as Element;
      target.textContent = target.textContent!.trim();
      await mytasks.saveTask({ id: task.id, task: target.textContent } as Task);
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
        let task = { task: selectedTask.textContent } as Task;
        if (selected) task.id = selected;
        await mytasks.saveTask(task);
      }
      selected = task.id;
    }
  };
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<li
  class="list-group-item"
  class:selected={task.id === selected}
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
  onclick={handleClick}
>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <i class="icon complete" onclick={async () => mytasks.completeTask(task)}></i>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <span
    class="task"
    contenteditable={task.id === selected}
    oncompositionstart={() => {
      composition = true;
    }}
    oncompositionend={() => {
      composition = false;
    }}
    onkeydown={handleKeydown}
    onpaste={pasteText}
  >
    {task.task}
  </span>
  <span class="created">{new Date(task.created).toLocaleDateString()}</span>
  {#if task.id === selected}
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <i class="icon delete" onclick={del}>delete</i>
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
