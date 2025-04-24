<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { confirm, fire, loading, pasteText } from "../misc.svelte";
  import { mytasks } from "../task.svelte";
  import Completed from "./Completed.svelte";
  import IncompleteTask from "./IncompleteTask.svelte";

  let selected = $state("");
  let editable = $state(false);
  let showCompleted = $state(false);
  let composition = $state(false);
  let showNewTask = $state(false);
  let newTask = $state("");
  let listElement: HTMLElement;
  let listEditIcon: HTMLElement;
  let addTaskButton: HTMLElement;
  let tasks: HTMLElement;
  let newTaskElement: HTMLElement;

  $effect(() => {
    if (editable) {
      listElement.focus();
      return;
    } else if (showNewTask) {
      newTaskElement.focus();
      return;
    }
    mytasks.getTasks();
    editable = false;
  });

  onMount(() => {
    mytasks.subscribe();
    return () => mytasks.abort();
  });

  onMount(() => {
    const sortable = new Sortable(tasks, {
      animation: 150,
      delay: 200,
      swapThreshold: 0.5,
      onUpdate: async (e) => {
        await mytasks.swapTask(
          mytasks.incomplete[e.oldIndex!],
          mytasks.incomplete[e.newIndex!],
        );
      },
    });
    return () => sortable.destroy();
  });

  const editList = async () => {
    listElement.textContent = listElement.textContent?.trim() || "";
    if (listElement.textContent) {
      if (listElement.textContent != mytasks.list.list)
        editable = (await mytasks.editList(listElement.textContent)) != 0;
      else editable = false;
    } else {
      listElement.textContent = mytasks.list.list;
      editable = false;
    }
  };
  const add = async () => {
    newTask = newTask.trim();
    if (newTask)
      if ((await mytasks.saveTask({ task: newTask } as Task)) != 0) return;
    newTask = "";
    showNewTask = false;
  };
  const edit = async (id: string, task: string) => {
    task = task.trim();
    const index = mytasks.incomplete.findIndex((task) => task.id === id);
    if (mytasks.incomplete[index].task != task)
      await mytasks.saveTask({ id, task } as Task);
    selected = "";
  };

  const addTask = async () => {
    if (showNewTask) await add();
    newTask = "";
    showNewTask = true;
    const range = document.createRange();
    range.selectNodeContents(newTaskElement);
    range.collapse(false);
    const sel = window.getSelection()!;
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const listKeydown = async (event: KeyboardEvent) => {
    if (composition) return;
    if (event.key == "Enter") {
      event.preventDefault();
      await editList();
    } else if (event.key == "Escape") {
      listElement.textContent = listElement.textContent?.trim() || "";
      if (listElement.textContent) listElement.textContent = "";
      else {
        listElement.textContent = mytasks.list.list;
        editable = false;
      }
    }
  };
  const listClick = async () => {
    if (editable) {
      if (mytasks.lists.length == 1)
        await fire("Error", "You must have at least one list!", "error");
      else if (await confirm("This list")) await mytasks.deleteList();
    } else {
      editable = true;
      const range = document.createRange();
      range.selectNodeContents(listElement);
      range.collapse(false);
      const sel = window.getSelection()!;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };

  const handleWindowClick = async (event: MouseEvent) => {
    if (loading.show) return;
    const target = event.target as Element;
    if (
      editable &&
      !listElement.contains(target) &&
      !listEditIcon.contains(target) &&
      !target.classList.contains("swal2-confirm")
    )
      await editList();
    if (
      showNewTask &&
      !newTaskElement.contains(target) &&
      !addTaskButton.contains(target)
    )
      await add();
    if (selected && !tasks.contains(target)) {
      const task = document.querySelector(".selected>.task");
      if (task) {
        task.textContent = task.textContent?.trim() || "";
        await edit(selected, task.textContent);
      }
    }
  };
</script>

<svelte:head>
  <title>{mytasks.list.list} - My Tasks</title>
</svelte:head>

<svelte:window onclick={handleWindowClick} />

<div style="height: 100%">
  <header style="padding-left: 20px">
    <div style="height: 50px">
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <span
        class="h3"
        bind:this={listElement}
        id="list"
        class:editable
        contenteditable={editable}
        oncompositionstart={() => (composition = true)}
        oncompositionend={() => (composition = false)}
        onkeydown={listKeydown}
        onpaste={pasteText}
      >
        {mytasks.list.list}
      </span>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <i class="icon edit" bind:this={listEditIcon} onclick={listClick}>
        {editable ? "delete" : "edit"}
      </i>
    </div>
    <button class="btn btn-primary" bind:this={addTaskButton} onclick={addTask}>
      Add Task
    </button>
  </header>
  <ul
    class="list-group list-group-flush"
    bind:this={tasks}
    style={showCompleted
      ? "height: calc(50% - 85px)"
      : "height: calc(100% - 170px)"}
  >
    <li
      class="list-group-item"
      class:new-task={showNewTask}
      style:display={showNewTask ? "" : "none"}
    >
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <span
        class="task"
        style="padding-left: 48px"
        bind:this={newTaskElement}
        bind:textContent={newTask}
        contenteditable
        oncompositionstart={() => (composition = true)}
        oncompositionend={() => (composition = false)}
        onpaste={pasteText}
        onkeydown={async (event) => {
          if (composition) return;
          if (event.key == "Enter" || event.key == "Escape") {
            event.preventDefault();
            await add();
          }
        }}
      ></span>
    </li>
    {#each mytasks.incomplete as task, i (task.id)}
      <IncompleteTask bind:selected bind:task={mytasks.incomplete[i]} />
    {/each}
  </ul>
  <Completed bind:show={showCompleted} />
</div>

<style>
  .h3 {
    cursor: default;
  }

  .edit {
    font-size: 1.25rem;
    color: #007bff;
    padding: 0;
  }

  .edit:hover {
    color: #0056b3;
  }

  .editable {
    text-decoration: underline;
  }

  #list {
    outline: 0;
    display: inline-block;
    min-width: 10px;
    padding-right: 1rem;
  }

  ul {
    cursor: default;
    overflow-y: auto;
  }
</style>
