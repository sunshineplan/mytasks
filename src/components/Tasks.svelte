<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  let currentTasks: Task[] = [];
  let currentCompleteds: Task[] = [];
  let selected: number;
  let editable = false;
  let showCompleted = false;

  const getTasks = async () => {
    showCompleted = false;
    if (!$tasks.hasOwnProperty($current.list)) {
      $loading++;
      const resp = await post("/task/get", { list: $current.id });
      $tasks[$current.list] = await resp.json();
      $loading--;
    }
    currentTasks = $tasks[$current.list].tasks;
    currentCompleteds = $tasks[$current.list].completeds;
  };

  $: $current && getTasks();

  onMount(() => {
    const sortable = new Sortable(
      document.querySelector("#tasks") as HTMLElement,
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

  const editList = async (list: string) => {
    const index = $lists.findIndex((list) => list.id === $current.id);
    if ($current.list != list) {
      $lists[index].list = list;
      delete Object.assign($tasks, { [list]: currentTasks })[$current.list];
      $loading++;
      const resp = await post("/list/edit/" + $current.id, {
        list: list.trim(),
      });
      const json = await resp.json();
      $loading--;
      if (!json.status)
        fire("Error", json.message ? json.message : "Error", "error");
    }
  };
  const add = async (task: string) => {
    if (task.trim()) {
      $loading++;
      const resp = await post("/task/add", {
        task: task.trim(),
        list: $current.id,
      });
      const json = await resp.json();
      $loading--;
      if (json.status) {
        if (json.id) {
          const index = $lists.findIndex((list) => list.id === $current.id);
          $lists[index].count++;
          $tasks[$current.list].tasks = [
            { id: json.id, task, created: new Date().toLocaleString() },
            ...currentTasks,
          ];
          currentTasks = $tasks[$current.list].tasks;
          const selected = document.querySelector(".selected");
          if (selected) selected.remove();
        }
      } else {
        await fire("Error", json.message ? json.message : "Error", "error");
      }
    } else {
      const selected = document.querySelector(".selected");
      if (selected) selected.remove();
    }
  };
  const edit = async (id: number, task: string) => {
    const index = currentTasks.findIndex((task) => task.id === id);
    if (currentTasks[index].task != task) {
      currentTasks[index].task = task;
      $loading++;
      const resp = await post("/task/edit/" + id, {
        task: task.trim(),
        list: $current.id,
      });
      const json = await resp.json();
      $loading--;
      if (!json.status)
        fire("Error", json.message ? json.message : "Error", "error");
    }
  };

  const addTask = async () => {
    const selectedTarget = document.querySelector(".selected>.task");
    if (!selected && selectedTarget)
      await add((selectedTarget as HTMLElement).innerText);
    selected = 0;
    const ul = document.querySelector("#tasks") as HTMLLIElement;
    const li = document.createElement("li");
    li.classList.add("list-group-item", "selected");
    const span = document.createElement("span");
    span.classList.add("task");
    span.style.paddingLeft = "48px";
    li.appendChild(span);
    li.addEventListener("keydown", async (event) => {
      if (event.key == "Enter" || event.key == "Escape") {
        event.preventDefault();
        await add((event.target as HTMLElement).innerText);
      }
    });
    ul.insertBefore(li, ul.childNodes[0]);
    span.setAttribute("contenteditable", "true");
    span.focus();
    const range = document.createRange();
    range.selectNodeContents(span);
    range.collapse(false);
    const sel = window.getSelection() as Selection;
    sel.removeAllRanges();
    sel.addRange(range);
  };
  const completeTask = async (id: number) => {
    $loading++;
    const resp = await post("/task/complete/" + id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count--;
        $lists = $lists;
        index = currentTasks.findIndex((task) => task.id === id);
        $tasks[$current.list].completeds = [
          {
            id: json.id,
            task: currentTasks[index].task,
            created: new Date().toLocaleString(),
          },
          ...currentCompleteds,
        ];
        currentCompleteds = $tasks[$current.list].completeds;
        currentTasks.splice(index, 1);
        currentTasks = currentTasks;
        return;
      }
    }
    await fire("Error", "Error", "error");
  };
  const incompleteTask = async (id: number) => {
    $loading++;
    const resp = await post("/task/incomplete/" + id);
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        let index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count++;
        $lists = $lists;
        index = currentCompleteds.findIndex((task) => task.id === id);
        $tasks[$current.list].tasks = [
          {
            id: json.id,
            task: currentCompleteds[index].task,
            created: new Date().toLocaleString(),
          },
          ...currentTasks,
        ];
        currentTasks = $tasks[$current.list].tasks;
        currentCompleteds.splice(index, 1);
        currentCompleteds = currentCompleteds;
        return;
      }
    }
    await fire("Error", "Error", "error");
  };
  const emptyCompleted = async () => {
    if (await confirm("These completed tasks")) {
      $loading++;
      const resp = await post("/task/empty/" + $current.id);
      const json = await resp.json();
      $loading--;
      if (json.status) currentCompleteds = [];
      else await fire("Error", "Error", "error");
    }
  };
  const delTask = async (id: number) => {
    if (await confirm("This task")) {
      $loading++;
      const resp = await post("/task/delete/" + id);
      $loading--;
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        let index = currentTasks.findIndex((task) => task.id === id);
        currentTasks.splice(index, 1);
        currentTasks = currentTasks;
        index = $lists.findIndex((list) => list.id === $current.id);
        $lists[index].count--;
      }
    }
  };

  const listKeydown = async (event: KeyboardEvent) => {
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      await editList((event.target as HTMLElement).innerText);
      editable = false;
    }
  };
  const listClick = async () => {
    if (editable) {
      if ($lists.length == 1)
        await fire("Error", "You must have at least one list!", "error");
      else {
        if (await confirm("This list")) {
          $loading++;
          const resp = await post("/list/delete/" + $current.id);
          $loading--;
          if (!resp.ok) await fire("Error", await resp.text(), "error");
          else {
            const index = $lists.findIndex((list) => list.id === $current.id);
            $lists.splice(index, 1);
            delete $tasks[$current.list];
            $current = $lists[0];
          }
        }
      }
    } else {
      editable = true;
      const target = document.querySelector("#list") as HTMLElement;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };
  const taskKeydown = async (event: KeyboardEvent, id: number) => {
    if (event.key == "Enter" || event.key == "Escape") {
      event.preventDefault();
      await edit(id, (event.target as HTMLElement).innerText);
      selected = 0;
    }
  };
  const taskClick = async (event: MouseEvent, id: number) => {
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
        await edit(selected, (selectedTarget as HTMLElement).innerText);
      selected = id;
    }
  };
  const expand = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (!target.classList.contains("delete")) showCompleted = !showCompleted;
  };
  const handleWindowClick = async (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (!target.classList.contains("task") && target.innerText !== "Add Task") {
      const id = selected;
      const selectedTarget = document.querySelector(".selected>.task");
      if (selectedTarget)
        if (id) await edit(id, (selectedTarget as HTMLElement).innerText);
        else await add((selectedTarget as HTMLElement).innerText);
      selected = 0;
    }
    if (target.id != "list" && !target.classList.contains("edit") && editable) {
      await editList(
        (document.querySelector("#list") as HTMLElement).innerText
      );
      editable = false;
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

  .completeds {
    padding: 15px 20px;
    margin-top: 16px;
    font-weight: 500;
    color: #5f6368;
    cursor: pointer;
    background-color: rgba(0, 0, 0, 0.125);
  }

  .expand {
    font-family: "Material Icons";
    font-style: normal;
    font-size: 1.5rem;
    float: right;
    line-height: normal;
  }

  .completed {
    text-decoration: line-through;
  }

  .editable {
    text-decoration: underline;
  }
</style>

<svelte:head>
  <title>{$current.list} - My Tasks</title>
</svelte:head>

<svelte:window on:click={handleWindowClick} />

<div style="height: 100%">
  <header style="padding-left: 20px">
    <div style="height: 50px">
      <span
        class="h3"
        id="list"
        class:editable
        contenteditable={editable}
        on:keydown={listKeydown}>{$current.list}</span>
      <span class="btn icon" on:click={listClick}>
        {#if !editable}
          <i class="material-icons edit">edit</i>
        {:else}<i class="material-icons edit">delete</i>{/if}
      </span>
    </div>
    <button class="btn btn-primary" on:click={addTask}>Add Task</button>
  </header>
  <ul
    class="list-group list-group-flush"
    style={showCompleted ? 'height:calc(50% - 85px)' : 'height:calc(100% - 170px)'}
    id="tasks">
    {#each currentTasks as task (task.id)}
      <li class="list-group-item" class:selected={task.id === selected}>
        <i class="check" on:click={async () => await completeTask(task.id)} />
        <span
          class="task"
          contenteditable={task.id === selected}
          on:keydown={async (e) => await taskKeydown(e, task.id)}
          on:click={async (e) => await taskClick(e, task.id)}>{task.task}</span>
        <span
          class="created">{new Date(task.created).toLocaleDateString()}</span>
        <i
          style="visibility:{task.id === selected ? 'visible' : 'hidden'}"
          class="operation delete"
          on:click={async () => await delTask(task.id)}>delete</i>
      </li>
    {/each}
  </ul>
  <div style="height: 100%">
    <div class="completeds" on:click={expand}>
      <span>Completed ({currentCompleteds.length})</span>
      <i class="expand">{showCompleted ? 'expand_more' : 'expand_less'}</i>
      {#if showCompleted && currentCompleteds.length}
        <i class="expand delete" on:click={emptyCompleted}>delete</i>
      {/if}
    </div>
    {#if showCompleted}
      <ul
        class="list-group list-group-flush"
        id="completeds"
        style="height:calc(50% - 85px)">
        {#each currentCompleteds as task (task.id)}
          <li class="list-group-item">
            <i
              class="uncheck"
              on:click={async () => await incompleteTask(task.id)}>done</i>
            <span class="completed">{task.task}</span>
            <span
              class="created">{new Date(task.created).toLocaleDateString()}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>
