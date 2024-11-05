<script lang="ts">
  import { pasteText, showSidebar } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let hover = $state(false);
  let composition = $state(false);
  let toggle: HTMLElement;
  let sidebar: HTMLElement;
  let addListButton: HTMLElement;
  let newListElement: HTMLElement;
  let showNewList = $state(false);
  let newList = $state("");

  $effect(() => {
    if (showNewList) newListElement.focus();
  });

  const goto = (list: List) => {
    showSidebar.close();
    mytasks.list = list;
    mytasks.component = "show";
  };

  const add = async () => {
    newList = newList.trim();
    if (newList) {
      const list: List = { list: newList, incomplete: 0, completed: 0 };
      await mytasks.addList(list);
      goto(list);
    }
  };

  const addList = async () => {
    if (showNewList) await add();
    else showNewList = true;
    newList = "";
    const range = document.createRange();
    range.selectNodeContents(newListElement);
    range.collapse(false);
    const sel = window.getSelection()!;
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      if (showNewList) {
        showNewList = false;
        newList = "";
      }
      const len = mytasks.lists.length;
      const index = mytasks.lists.findIndex(
        (list) => list.list === mytasks.list.list,
      );
      if (mytasks.component === "show")
        if (event.key == "ArrowUp") {
          if (index > 0) goto(mytasks.lists[index - 1]);
        } else if (event.key == "ArrowDown")
          if (index < len - 1) goto(mytasks.lists[index + 1]);
    }
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      showNewList &&
      !addListButton.contains(target) &&
      !newListElement.contains(target) &&
      !target.classList.contains("swal2-confirm")
    ) {
      await add();
      showNewList = false;
      newList = "";
    }
    if (
      showSidebar.status &&
      !toggle.contains(target) &&
      !sidebar.contains(target)
    )
      showSidebar.close();
  };
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleClick} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<span
  class="toggle"
  bind:this={toggle}
  onclick={() => showSidebar.toggle()}
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
>
  <svg viewBox="0 0 70 70" width="40" height="30">
    {#each [10, 30, 50] as y}
      <rect {y} width="100%" height="10" fill={hover ? "#1a73e8" : "white"} />
    {/each}
  </svg>
</span>
<nav
  class="nav flex-column navbar-light sidebar"
  class:show={showSidebar.status}
  bind:this={sidebar}
>
  <div class="list-menu">
    <button
      class="btn btn-primary btn-sm"
      bind:this={addListButton}
      onclick={addList}>Add List</button
    >
    <ul class="navbar-nav" id="lists">
      {#each mytasks.lists as l (l.list)}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li
          class="nav-link list"
          class:active={mytasks.list.list === l.list &&
            mytasks.component === "show"}
          onclick={() => goto(l)}
        >
          {l.list} ({l.incomplete})
        </li>
      {/each}
      <li
        class="nav-link new"
        style:display={showNewList ? "" : "none"}
        bind:this={newListElement}
        bind:textContent={newList}
        contenteditable
        onpaste={pasteText}
        oncompositionstart={() => (composition = true)}
        oncompositionend={() => (composition = false)}
        onkeydown={async (event) => {
          if (composition) return;
          if (event.key == "Enter") {
            event.preventDefault();
            newList = newList.trim();
            if (newList) await add();
            else newList = "";
            showNewList = false;
          } else if (event.key == "Escape") {
            newList = "";
            showNewList = false;
          }
        }}
      ></li>
    </ul>
  </div>
</nav>

<style>
  .toggle {
    display: none;
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

  @media (max-width: 900px) {
    .toggle {
      display: block;
    }

    .sidebar {
      left: -100%;
      transition: left 0.3s ease-in-out;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }

    .show {
      left: 0;
    }
  }
</style>
