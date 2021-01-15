<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { fire, post } from "../misc";
  import { current, component, showSidebar, loading, lists } from "../stores";
  import type { List } from "../stores";

  let hover = false;
  let smallSize = window.innerWidth <= 900;

  const toggle = () => {
    $showSidebar = !$showSidebar;
  };

  const goto = (list: List) => {
    if (window.innerWidth <= 900) $showSidebar = false;
    $current = list;
    $component = "show";
  };

  const add = async (list: string) => {
    $loading++;
    const resp = await post("/list/add", { list: list.trim() });
    const json = await resp.json();
    $loading--;
    if (json.status) {
      if (json.id) {
        (document.querySelector(".new") as Element).remove();
        const newList: List = {
          id: json.id,
          list,
          incomplete: 0,
          completed: 0,
        };
        $lists = [...$lists, newList];
        goto(newList);
        return true;
      }
    }
    await fire("Error", json.message ? json.message : "Error", "error");
    return false;
  };

  const addList = async () => {
    if (window.innerWidth <= 900) $showSidebar = false;
    const newList = document.querySelector(".new");
    let ok = true;
    if (newList) ok = await add((newList as HTMLElement).innerText);
    if (ok) {
      const ul = document.querySelector("ul.navbar-nav") as Element;
      const li = document.createElement("li");
      li.classList.add("nav-link", "new");
      ul.appendChild(li);
      li.addEventListener("keydown", async (event) => {
        const target = event.target as Element;
        const list = (target.textContent as string).trim();
        if (event.key == "Enter") {
          event.preventDefault();
          if (list) await add(list);
          else target.remove();
        } else if (event.key == "Escape") {
          if (list) target.textContent = "";
          else target.remove();
        }
      });
      li.setAttribute("contenteditable", "true");
      li.focus();
      const range = document.createRange();
      range.selectNodeContents(li);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };

  const checkSize = () => {
    if (smallSize != window.innerWidth <= 900)
      smallSize = window.innerWidth <= 900;
  };
  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      const newList = document.querySelector(".new");
      if (newList) newList.remove();
      const len = $lists.length;
      const index = $lists.findIndex((list) => list.id === $current.id);
      if ($current.id && $component === "show")
        if (event.key == "ArrowUp") {
          if (index > 0) goto($lists[index - 1]);
        } else if (event.key == "ArrowDown")
          if (index < len - 1) goto($lists[index + 1]);
    }
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      !target.classList.contains("new") &&
      !target.classList.contains("swal2-confirm") &&
      target.textContent !== "Add List"
    ) {
      const newList = document.querySelector(".new");
      if (newList) {
        const list = (newList.textContent as string).trim();
        if (list) await add(list);
        else newList.remove();
      }
    }
  };

  onMount(() => {
    const sortable = new Sortable(
      document.querySelector("#lists") as HTMLElement,
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
    const resp = await post("/list/reorder", {
      old: $lists[event.oldIndex as number].id,
      new: $lists[event.newIndex as number].id,
    });
    if ((await resp.text()) == "1") {
      const list = $lists[event.oldIndex as number];
      $lists.splice(event.oldIndex as number, 1);
      $lists.splice(event.newIndex as number, 0, list);
    } else await fire("Error", "Failed to reorder list", "error");
  };
</script>

<svelte:window
  on:keydown={handleKeydown}
  on:resize={checkSize}
  on:click={handleClick}
/>

{#if smallSize}
  <span
    class="toggle"
    on:click={toggle}
    on:mouseenter={() => (hover = true)}
    on:mouseleave={() => (hover = false)}>
    <svg viewBox="0 0 70 70" width="40" height="30">
      {#each [10, 30, 50] as y}
        <rect {y} width="100%" height="10" fill={hover ? "#1a73e8" : "white"} />
      {/each}
    </svg>
  </span>
{/if}
<nav
  class="nav flex-column navbar-light sidebar"
  hidden={!$showSidebar && smallSize}
>
  <div class="list-menu">
    <button class="btn btn-primary btn-sm" on:click={addList}>Add List</button>
    <ul class="navbar-nav" id="lists">
      {#each $lists as list (list.id)}
        <li
          class="nav-link list"
          class:active={$current.id === list.id && $component === "show"}
          on:click={() => goto(list)}
        >
          {list.list} ({list.incomplete})
        </li>
      {/each}
    </ul>
  </div>
</nav>

<style>
  .toggle {
    position: fixed;
    z-index: 100;
    top: 0;
    padding: 20px;
    color: white !important;
  }

  .toggle:hover {
    background-color: rgb(232, 232, 232);
  }

  .sidebar {
    position: fixed;
    top: 0;
    z-index: 1;
    height: 100%;
    width: 250px;
    padding-top: 70px;
    user-select: none;
  }

  .list-menu {
    height: 100%;
    width: 100%;
    padding-top: 10px;
    overflow-x: hidden;
    border-right: 1px solid #e9ecef;
    background-color: white;
  }

  .list-menu .btn {
    margin-left: 20px;
    margin-bottom: 5px;
  }

  .list-menu .navbar-nav {
    text-indent: 20px;
  }

  .list-menu .nav-link:hover {
    background-color: rgb(232, 232, 232);
  }

  .list {
    display: block;
    cursor: pointer;
    margin: 0;
    border-left: 5px solid transparent;
    color: rgba(0, 0, 0, 0.7) !important;
  }

  .active {
    border-left: 5px solid #1a73e8;
    color: #1a73e8 !important;
  }

  .nav-link.active {
    background-color: #eaf5fd;
  }

  @media (min-width: 901px) {
    .sidebar {
      display: block !important;
    }
  }

  @media (max-width: 900px) {
    .sidebar {
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }
  }
</style>
