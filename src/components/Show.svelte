<script lang="ts">
  import { onMount } from "svelte";
  import { confirm, fire, loading, pasteText } from "../misc.svelte";
  import { mytasks } from "../task.svelte";
  import Completed from "./Completed.svelte";
  import Incomplete from "./Incomplete.svelte";

  let selected = $state("");
  let editable = $state(false);
  let showCompleted = $state(false);
  let composition = $state(false);

  $effect(() => {
    mytasks.getTasks();
    editable = false;
  });

  onMount(() => {
    mytasks.subscribe(true);
    return () => mytasks.controller.abort();
  });

  const editList = async (list: string) => {
    list = list.trim();
    if (mytasks.list.list != list) return (await mytasks.editList(list)) == 0;
    return true;
  };
  const add = async (task: string) => {
    task = task.trim();
    if (task) if ((await mytasks.saveTask({ task } as Task)) != 0) return;
    const selected = document.querySelector(".selected");
    if (selected) selected.remove();
  };
  const edit = async (id: string, task: string) => {
    task = task.trim();
    const index = mytasks.tasks.incomplete.findIndex((task) => task.id === id);
    if (mytasks.tasks.incomplete[index].task != task)
      await mytasks.saveTask({ id, task } as Task);
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
        target.textContent = mytasks.list.list;
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (target.textContent) target.textContent = "";
      else {
        target.textContent = mytasks.list.list;
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
    if (loading.show) return;
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
        target.textContent = mytasks.list.list;
        editable = false;
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
        id="list"
        class:editable
        contenteditable={editable}
        oncompositionstart={() => {
          composition = true;
        }}
        oncompositionend={() => {
          composition = false;
        }}
        onkeydown={listKeydown}
        onpaste={pasteText}
      >
        {mytasks.list.list}
      </span>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <span onclick={listClick}>
        {#if !editable}
          <i class="icon edit">edit</i>
        {:else}<i class="icon edit">delete</i>{/if}
      </span>
    </div>
    <button class="btn btn-primary" onclick={addTask}>Add Task</button>
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
