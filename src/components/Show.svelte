<script lang="ts">
  import { onMount, createEventDispatcher } from "svelte";
  import Incomplete from "./Incomplete.svelte";
  import Completed from "./Completed.svelte";
  import { fire, confirm, post, pasteText } from "../misc";
  import { current, lists, tasks, poll, init } from "../task";
  import { loading } from "../stores";

  const dispatch = createEventDispatcher();

  let currentIncomplete: Task[] = [];
  let currentCompleted: Task[] = [];
  let selected: string;
  let editable = false;
  let showCompleted = false;
  let composition = false;

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
      if (!$current.list) if ($lists.length) $current = $lists[0];
      const resp = await post("/get", { list: $current.list });
      if (resp.ok) $tasks[$current.list] = await resp.json();
      else {
        await fire("Error", await resp.text(), "error");
        return;
      }
    }
    refresh();
  };

  $: $current, getTasks(), (editable = false);

  const editList = async (list: string) => {
    list = list.trim();
    if ($current.list != list) {
      const resp = await post("/list/edit", { old: $current.list, new: list });
      let json: any = {};
      if (resp.ok) {
        json = await resp.json();
        if (json.status) {
          const index = $lists.findIndex((list) => list.list === $current.list);
          $lists[index].list = list;
          delete Object.assign($tasks, { [list]: currentIncomplete })[
            $current.list
          ];
          $current = $lists[index];
          return true;
        }
      } else json.message = await resp.text();
      await fire("Error", json.message ? json.message : "Error", "error");
      dispatch("reload");
      return false;
    }
    return true;
  };
  const add = async (task: string) => {
    task = task.trim();
    if (task) {
      const resp = await post("/task/add", { task, list: $current.list });
      let json: any = {};
      if (resp.ok) {
        json = await resp.json();
        if (json.status && json.id) {
          const index = $lists.findIndex((list) => list.list === $current.list);
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
        await fire("Error", "Error", "error");
      } else await fire("Error", await resp.text(), "error");
    } else {
      const selected = document.querySelector(".selected");
      if (selected) selected.remove();
    }
  };
  const edit = async (id: string, task: string) => {
    task = task.trim();
    const index = currentIncomplete.findIndex((task) => task.id === id);
    if (currentIncomplete[index].task != task) {
      currentIncomplete[index].task = task;
      const resp = await post("/task/edit/" + id, {
        task,
        list: $current.list,
      });
      if (!resp.ok) await fire("Error", await resp.text(), "error");
    }
  };

  const addTask = async () => {
    const task = document.querySelector(".selected>.task");
    if (!selected && task) {
      task.textContent = task.textContent!.trim();
      await add(task.textContent);
    }
    selected = "";
    const ul = document.querySelector("#tasks")!;
    const li = document.createElement("li");
    li.classList.add("list-group-item", "selected");
    const span = document.createElement("span");
    span.classList.add("task");
    span.style.paddingLeft = "48px";
    li.appendChild(span);
    li.addEventListener("paste", pasteText);
    let composition = false;
    li.addEventListener("compositionstart", () => {
      composition = true;
    });
    li.addEventListener("compositionend", () => {
      composition = false;
    });
    li.addEventListener("keydown", async (event) => {
      if (composition) return;
      if (event.key == "Enter" || event.key == "Escape") {
        event.preventDefault();
        const target = event.target as Element;
        target.textContent = target.textContent!.trim();
        await add(target.textContent);
      }
    });
    ul.insertBefore(li, ul.childNodes[0]);
    span.setAttribute("contenteditable", "true");
    span.focus();
    const range = document.createRange();
    range.selectNodeContents(span);
    range.collapse(false);
    const sel = window.getSelection()!;
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const listKeydown = async (event: KeyboardEvent) => {
    if (composition) return;
    const target = event.target as Element;
    target.textContent = target.textContent!.trim();
    if (event.key == "Enter") {
      event.preventDefault();
      if (target.textContent) editable = !(await editList(target.textContent));
      else {
        target.textContent = $current.list;
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (target.textContent) target.textContent = "";
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
        const resp = await post("/list/delete", { list: $current.list });
        if (resp.ok) {
          const json = await resp.json();
          if (json.status) {
            const index = $lists.findIndex(
              (list) => list.list === $current.list,
            );
            $lists.splice(index, 1);
            delete $tasks[$current.list];
            $current = $lists[0];
          } else {
            await fire("Error", "Error", "error");
            dispatch("reload");
          }
        } else await fire("Error", await resp.text(), "error");
      }
    } else {
      editable = true;
      const target = document.querySelector<HTMLElement>("#list")!;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection()!;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };

  const handleWindowClick = async (event: MouseEvent) => {
    if ($loading) return;
    const target = event.target as Element;
    if (
      target.parentNode &&
      !(target.parentNode as Element).classList.contains("selected") &&
      target.textContent !== "Add Task"
    ) {
      const id = selected;
      const task = document.querySelector(".selected>.task");
      if (task) {
        task.textContent = task.textContent!.trim();
        if (id) await edit(id, task.textContent);
        else await add(task.textContent);
      }
      selected = "";
    }
    if (
      target.id !== "list" &&
      !target.classList.contains("edit") &&
      !target.classList.contains("swal2-confirm") &&
      editable
    ) {
      const list = document.querySelector("#list")!;
      list.textContent = list.textContent!.trim();
      if (list.textContent) editable = !(await editList(list.textContent));
      else {
        target.textContent = $current.list;
        editable = false;
      }
    }
  };

  const subscribe = async (signal: AbortSignal) => {
    const resp = await poll(signal);
    if (resp.status == 200) await subscribe(signal);
    else if (resp.status == 401) {
      dispatch("reload");
    } else if (resp.status == 409) {
      loading.start();
      await init();
      await getTasks(true);
      loading.end();
      await subscribe(signal);
    } else {
      await new Promise((sleep) => setTimeout(sleep, 30000));
      await subscribe(signal);
    }
  };
  onMount(() => {
    const controller = new AbortController();
    subscribe(controller.signal);
    return () => controller.abort();
  });
</script>

<svelte:head>
  <title>{$current.list} - My Tasks</title>
</svelte:head>

<svelte:window on:click={handleWindowClick} />

<div style="height: 100%">
  <header style="padding-left: 20px">
    <div style="height: 50px">
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <span
        class="h3"
        id="list"
        class:editable
        contenteditable={editable}
        on:compositionstart={() => {
          composition = true;
        }}
        on:compositionend={() => {
          composition = false;
        }}
        on:keydown={listKeydown}
        on:paste={pasteText}
      >
        {$current.list}
      </span>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
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
    on:add={async (e) => await add(e.detail.task)}
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
