<script lang="ts">
  import { confirm, pasteText } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let {
    selected = $bindable(),
    task = $bindable(),
  }: {
    selected: string;
    task: Task;
  } = $props();

  let hover = $state(false);
  let composition = $state(false);
  let taskElement: HTMLElement;
  let complete: HTMLElement;
  let edit: HTMLElement;

  let editable = $derived(selected === task.id);

  $effect(() => {
    if (editable) {
      const range = document.createRange();
      range.selectNodeContents(taskElement);
      range.collapse(false);
      const sel = window.getSelection()!;
      sel.removeAllRanges();
      sel.addRange(range);
      taskElement.focus();
    }
  });

  const del = async () => {
    if (await confirm("This task")) await mytasks.deleteTask(task);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (composition) return;
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      task.task = task.task.trim();
      await mytasks.saveTask({ id: task.id, task: task.task } as Task);
      selected = "";
    }
  };
  const handleClick = async (event: MouseEvent) => {
    let target = event.target as HTMLElement;
    if (!editable && !complete.contains(target)) {
      const selectedTask = document.querySelector(".selected>.task");
      if (selectedTask) {
        selectedTask.textContent = selectedTask.textContent?.trim() || "";
        if (selectedTask.textContent)
          await mytasks.saveTask({
            id: selected,
            task: selectedTask.textContent,
          } as Task);
      }
      selected = task.id;
    }
  };
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<li
  class="list-group-item"
  class:selected={editable}
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
  onclick={handleClick}
>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <i
    class="icon complete"
    bind:this={complete}
    onclick={async () => mytasks.completeTask(task)}
  ></i>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <span
    class="task"
    bind:this={taskElement}
    contenteditable={editable}
    oncompositionstart={() => (composition = true)}
    oncompositionend={() => (composition = false)}
    onkeydown={handleKeydown}
    onpaste={pasteText}
  >
    {task.task}
  </span>
  <span class="created">{new Date(task.created).toLocaleDateString()}</span>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <i
    bind:this={edit}
    class:icon={hover}
    class:delete={editable}
    style:display={hover ? "" : "none"}
    onclick={editable ? del : null}
    >{editable ? "delete" : hover ? "edit" : ""}</i
  >
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
