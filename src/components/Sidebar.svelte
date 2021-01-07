<script lang="ts">
  import { slide } from "svelte/transition";
  import { List, current } from "../stores";

  export let lists: List[];

  $current = lists[0];
  let hover = false;
  let showSidebar = true;

  const toggle = () => {
    showSidebar = !showSidebar;
  };

  const arrow = (event: KeyboardEvent) => {
    const len = lists.length;
    var index = lists.findIndex((list) => list.id === $current.id);
    if (event.key == "ArrowUp") {
      if (index > 0) $current = lists[index - 1];
    } else if (event.key == "ArrowDown")
      if (index < len - 1) $current = lists[index + 1];
  };
  const add = () => {
    if (window.innerWidth <= 900) showSidebar = false;
    console.log("/list/add");
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

<svelte:window on:keydown={arrow} />

{#if window.innerWidth <= 900}
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
  hidden={!showSidebar}
  transition:slide>
  <div class="list-menu">
    <button class="btn btn-primary btn-sm" on:click={add}> Add List </button>
    <ul class="navbar-nav">
      {#each lists as list}
        <li>
          <span
            class="nav-link list"
            class:active={$current.id === list.id}
            on:click={() => ($current = list)}>
            {list.list}
            ({list.count})
          </span>
        </li>
      {/each}
    </ul>
  </div>
</nav>
