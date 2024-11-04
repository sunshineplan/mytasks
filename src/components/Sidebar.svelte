<script lang="ts">
  import { pasteText, showSidebar } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let hover = $state(false);

  const goto = (list: List) => {
    if (window.innerWidth <= 900) showSidebar.close();
    mytasks.list = list;
    mytasks.component = "show";
    const ul = document.querySelector("#tasks");
    if (ul) ul.scrollTop = 0;
  };

  const add = async (list: string) => {
    list = list.trim();
    document.querySelector(".new")?.remove();
    const newList: List = {
      list,
      incomplete: 0,
      completed: 0,
    };
    await mytasks.addList(newList);
    goto(newList);
  };

  const addList = async () => {
    if (window.innerWidth <= 900) showSidebar.close();
    const newList = document.querySelector<HTMLElement>(".new");
    if (newList) await add(newList.innerText);
    const ul = document.querySelector("ul.navbar-nav")!;
    const li = document.createElement("li");
    li.classList.add("nav-link", "new");
    ul.appendChild(li);
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
      const target = event.target as Element;
      const list = target.textContent!.trim();
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
    const sel = window.getSelection()!;
    sel.removeAllRanges();
    sel.addRange(range);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      const newList = document.querySelector(".new");
      if (newList) newList.remove();
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
      !target.classList.contains("new") &&
      !target.classList.contains("swal2-confirm") &&
      target.textContent !== "Add List"
    ) {
      const newList = document.querySelector(".new");
      if (newList) {
        const list = newList.textContent!.trim();
        if (list) await add(list);
        else newList.remove();
      }
    }
  };
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleClick} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<span
  class="toggle"
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
>
  <div class="list-menu">
    <button class="btn btn-primary btn-sm" onclick={addList}>Add List</button>
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
