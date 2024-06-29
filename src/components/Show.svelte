<script lang="ts">
  import Cookies from "js-cookie";
  import { onMount, createEventDispatcher } from "svelte";
  import Incomplete from "./Incomplete.svelte";
  import Completed from "./Completed.svelte";
  import { fire, confirm, poll, pasteText } from "../misc";
  import { list, lists, tasks, init } from "../task";
  import { loading } from "../stores";

  const dispatch = createEventDispatcher();

  let selected: string;
  let editable = false;
  let showCompleted = false;
  let composition = false;

  $: $list, tasks.get(), (editable = false);

  const subscribe = async (signal: AbortSignal) => {
    const resp = await poll(signal);
    if (resp.ok) {
      const last = await resp.text();
      if (last && Cookies.get("last") != last) {
        loading.start();
        await init();
        $list = $list;
        loading.end();
      }
      await subscribe(signal);
    } else if (resp.status == 401) {
      dispatch("reload");
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

  const editList = async (list: string) => {
    list = list.trim();
    if ($list.list != list) return (await lists.edit(list)) == 0;
    return true;
  };
  const add = async (task: string) => {
    task = task.trim();
    if (task) if ((await tasks.save(<Task>{ task })) != 0) return;
    const selected = document.querySelector(".selected");
    if (selected) selected.remove();
  };
  const edit = async (id: string, task: string) => {
    task = task.trim();
    const index = $tasks.incomplete.findIndex((task) => task.id === id);
    if ($tasks.incomplete[index].task != task)
      await tasks.save(<Task>{ id, task });
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
        target.textContent = $list.list;
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (target.textContent) target.textContent = "";
      else {
        target.textContent = $list.list;
        editable = false;
      }
    }
  };
  const listClick = async () => {
    if (editable) {
      if ($lists.length == 1)
        await fire("Error", "You must have at least one list!", "error");
      else if (await confirm("This list")) await lists.delete();
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
        target.textContent = $list.list;
        editable = false;
      }
    }
  };
</script>

<svelte:head>
  <title>{$list.list} - My Tasks</title>
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
        {$list.list}
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
  <Incomplete bind:showCompleted bind:selected />
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
</style>
