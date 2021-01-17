<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Incomplete from "./Incomplete.svelte";
  import Completed from "./Completed.svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, lists, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  let currentIncomplete: Task[] = [];
  let currentCompleted: Task[] = [];
  let selected: number;
  let editable = false;
  let showCompleted = false;

  const refresh = () => {
    $lists = $lists;
    currentIncomplete = $tasks[$current.list].incomplete;
    currentCompleted = $tasks[$current.list].completed;
  };

  const reload = async () => {
    await getTasks(true);
    dispatch("reload");
  };

  const getTasks = async (force?: boolean) => {
    if (!force) showCompleted = false;
    if (!$tasks.hasOwnProperty($current.list) || force) {
      $loading++;
      const resp = await post("/get", { list: $current.id });
      $tasks[$current.list] = await resp.json();
      $loading--;
    }
    refresh();
  };

  $: $current, getTasks(), (editable = false);

  const editList = async (list: string) => {
    if ($current.list != list) {
      $loading++;
      const resp = await post("/list/edit/" + $current.id, {
        list: list.trim(),
      });
      $loading--;
      let json: any = {};
      if (resp.ok) {
        json = await resp.json();
        if (json.status) {
          const index = $lists.findIndex((list) => list.id === $current.id);
          $lists[index].list = list;
          delete Object.assign($tasks, { [list]: currentIncomplete })[
            $current.list
          ];
          return true;
        }
      }
      await fire("Error", json.message ? json.message : "Error", "error");
      dispatch("reload");
      return false;
    }
  };
  const add = async (task: string) => {
    if (task.trim()) {
      $loading++;
      const resp = await post("/task/add", {
        task: task.trim(),
        list: $current.id,
      });
      $loading--;
      let json: any = {};
      if (resp.ok) {
        json = await resp.json();
        if (json.status && json.id) {
          const index = $lists.findIndex((list) => list.id === $current.id);
          $lists[index].incomplete++;
          $tasks[$current.list].incomplete = [
            { id: json.id, task, created: new Date().toLocaleString() },
            ...currentIncomplete,
          ];
          currentIncomplete = $tasks[$current.list].incomplete;
          const selected = document.querySelector(".selected");
          if (selected) selected.remove();
          return;
        }
      }
      await fire("Error", "Error", "error");
    } else {
      const selected = document.querySelector(".selected");
      if (selected) selected.remove();
    }
  };
  const edit = async (id: number, task: string) => {
    const index = currentIncomplete.findIndex((task) => task.id === id);
    if (currentIncomplete[index].task != task) {
      currentIncomplete[index].task = task;
      $loading++;
      const resp = await post("/task/edit/" + id, {
        task: task.trim(),
        list: $current.id,
      });
      $loading--;
      if (!resp.ok) await fire("Error", "Error", "error");
    }
  };

  const addTask = async () => {
    const selectedTarget = document.querySelector(".selected>.task");
    if (!selected && selectedTarget)
      await add((selectedTarget as HTMLElement).innerText);
    selected = 0;
    const ul = document.querySelector("#tasks") as Element;
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

  const listKeydown = async (event: KeyboardEvent) => {
    const target = event.target as Element;
    const list = (target.textContent as string).trim();
    if (event.key == "Enter") {
      event.preventDefault();
      if (list) editable = !(await editList(list));
      else {
        target.textContent = $current.list;
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (list) target.textContent = "";
      else {
        target.textContent = $current.list;
        editable = false;
      }
    }
  };
  const listClick = async () => {
    if (editable) {
      if ($lists.length == 1)
        await fire("Error", "You must have at least one list!", "error");
      else if (await confirm("This list")) {
        $loading++;
        const resp = await post("/list/delete/" + $current.id);
        $loading--;
        if (resp.ok) {
          const index = $lists.findIndex((list) => list.id === $current.id);
          $lists.splice(index, 1);
          delete $tasks[$current.list];
          $current = $lists[0];
        } else {
          await fire("Error", await resp.text(), "error");
          dispatch("reload");
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

  const handleWindowClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      target.parentNode &&
      !(target.parentNode as Element).classList.contains("selected") &&
      target.textContent !== "Add Task"
    ) {
      const id = selected;
      const selectedTarget = document.querySelector(".selected>.task");
      if (selectedTarget)
        if (id) await edit(id, (selectedTarget as HTMLElement).innerText);
        else await add((selectedTarget as HTMLElement).innerText);
      selected = 0;
    }
    if (
      target.id !== "list" &&
      !target.classList.contains("edit") &&
      !target.classList.contains("swal2-confirm") &&
      editable
    ) {
      const list = (document.querySelector("#list") as Element).textContent;
      if (list && list.trim()) editable = !(await editList(list));
      else {
        target.textContent = $current.list;
        editable = false;
      }
    }
  };
</script>

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
        on:keydown={listKeydown}>
        {$current.list}
      </span>
      <span on:click={listClick}>
        {#if !editable}
          <i class="icon edit">edit</i>
        {:else}<i class="icon edit">delete</i>{/if}
      </span>
    </div>
    <button class="btn btn-primary" on:click={addTask}>Add Task</button>
  </header>
  <Incomplete
    bind:showCompleted
    bind:selected
    bind:incompleteTasks={currentIncomplete}
    on:edit={async (e) => await edit(e.detail.id, e.detail.task)}
    on:refresh={refresh}
    on:reload={reload}
  />
  <Completed
    bind:show={showCompleted}
    bind:completedTasks={currentCompleted}
    on:refresh={refresh}
    on:reload={reload}
  />
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
</style>
