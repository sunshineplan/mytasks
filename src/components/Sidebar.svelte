<script lang="ts">
  import { slide } from "svelte/transition";
  import { fire, post } from "../misc";
  import { current, component, showSidebar, lists } from "../stores";
  import type { List } from "../stores";

  let hover = false;
  let smallSize = window.innerWidth <= 900;

  const toggle = () => {
    $showSidebar = !$showSidebar;
  };

  const goto = (list: List) => {
    if (window.innerWidth <= 900) $showSidebar = false;
    $current = list;
    $component = "tasks";
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      const len = $lists.length;
      const index = $lists.findIndex((list) => list.id === $current.id);
      if ($current.id && $component === "tasks")
        if (event.key == "ArrowUp") {
          if (index > 0) goto($lists[index - 1]);
        } else if (event.key == "ArrowDown")
          if (index < len - 1) goto($lists[index + 1]);
    } else if (event.key == "Enter" || event.key == "Escape") {
      const newList = document.querySelector(".new");
      if (newList) {
        event.preventDefault();
        const list = (newList.textContent as string).trim();
        if (list) {
          const id = await add(list);
          if (id) {
            newList.remove();
            $lists = [...$lists, { id, list, count: 0 }];
            goto({ id, list, count: 0 });
          }
        } else newList.remove();
      }
    }
  };

  const add = async (list: string) => {
    const resp = await post("/list/add", { list });
    const json = await resp.json();
    if (json.status) return json.id as number;
    else {
      await fire("Error", json.message ? json.message : "Error", "error");
      return 0;
    }
  };

  const checkSize = () => {
    if (smallSize != window.innerWidth <= 900)
      smallSize = window.innerWidth <= 900;
  };

  const addList = () => {
    if (window.innerWidth <= 900) $showSidebar = false;
    const ul = document.querySelector("ul.navbar-nav") as HTMLLIElement;
    const li = document.createElement("li");
    li.classList.add("nav-link", "new");
    ul.appendChild(li);
    li.setAttribute("contenteditable", "true");
    li.focus();
    const range = document.createRange();
    range.selectNodeContents(li);
    range.collapse(false);
    const sel = window.getSelection() as Selection;
    sel.removeAllRanges();
    sel.addRange(range);
  };
</script>

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

<svelte:window on:keydown={handleKeydown} on:resize={checkSize} />

{#if smallSize}
  <span
    class="toggle"
    on:click={toggle}
    on:mouseenter={() => (hover = true)}
    on:mouseleave={() => (hover = false)}>
    <svg viewBox="0 0 70 70" width="40" height="30">
      {#each [10, 30, 50] as y}
        <rect {y} width="100%" height="10" fill={hover ? '#1a73e8' : 'white'} />
      {/each}
    </svg>
  </span>
{/if}
<nav
  class="nav flex-column navbar-light sidebar"
  hidden={!$showSidebar && smallSize}
  transition:slide>
  <div class="list-menu">
    <button class="btn btn-primary btn-sm" on:click={addList}>Add List</button>
    <ul class="navbar-nav">
      {#each $lists as list (list.id)}
        <li
          class="nav-link list"
          class:active={$current.id === list.id && $component === 'tasks'}
          on:click={() => goto(list)}>
          {list.list}
          ({list.count})
        </li>
      {/each}
    </ul>
  </div>
</nav>
