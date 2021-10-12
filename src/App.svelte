<script lang="ts">
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Show from "./components/Show.svelte";
  import { fire, post } from "./misc";
  import {
    username as user,
    current,
    showSidebar,
    component,
    loading,
    lists,
    reset,
  } from "./stores";

  const getInfo = async () => {
    $loading++;
    const resp = await fetch("/info");
    const info = await resp.json();
    if (Object.keys(info).length) {
      $user = info.username;
      $lists = info.lists;
    } else reset();
    $loading--;
    if ($lists.length) if (!$current.list) $current = $lists[0];
  };
  const promise = getInfo();

  const components = {
    setting: Setting,
    show: Show,
  } as { [component: string]: any };

  const setting = () => {
    if (window.innerWidth <= 900) $showSidebar = false;
    $component = "setting";
  };

  const logout = async () => {
    const resp = await post("@universal@/logout", undefined, true);
    if (resp.ok) {
      await getInfo();
      window.history.pushState({}, "", "/");
      $component = "show";
    } else await fire("Error", "Unknow error", "error");
  };
</script>

<nav class="navbar navbar-light topbar">
  <div class="d-flex" style="height: 100%">
    <a class="brand" class:user={$user} href="/">My Tasks</a>
  </div>
  {#if $user}
    <div class="navbar-nav flex-row">
      <span class="nav-link">{$user}</span>
      <span class="nav-link link" on:click={setting}>Setting</span>
      <span class="nav-link link" on:click={logout}>Logout</span>
    </div>
  {:else}
    <div class="navbar-nav flex-row"><span class="nav-link">Log In</span></div>
  {/if}
</nav>
{#await promise then _}
  {#if !$user}
    {#if !$loading}
      <Login on:info={getInfo} />
    {/if}
  {:else}
    <Sidebar on:reload={getInfo} />
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}"
      on:mousedown={() => ($showSidebar = false)}
    >
      <svelte:component this={components[$component]} on:reload={getInfo} />
    </div>
  {/if}
{/await}
<div class={$user ? "loading" : "initializing"} hidden={!$loading}>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
  </div>
</div>

<style>
  .topbar {
    position: fixed;
    top: 0px;
    z-index: 2;
    width: 100%;
    height: 70px;
    padding: 0 10px 0 0;
    background-color: #1a73e8;
    user-select: none;
  }

  .topbar .nav-link {
    padding-left: 8px;
    padding-right: 8px;
    color: white !important;
  }

  .topbar .link:hover {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 5px;
    cursor: pointer;
  }

  .brand {
    padding-left: 20px;
    margin: auto;
    font-size: 25px;
    letter-spacing: 0.3px;
    color: white;
    text-decoration: none;
  }

  .initializing {
    position: fixed;
    top: 70px;
    height: calc(100% - 70px);
    width: 100%;
    display: flex;
  }

  .loading {
    position: fixed;
    z-index: 2;
    top: 70px;
    left: 250px;
    height: calc(100% - 70px);
    width: calc(100% - 250px);
    display: flex;
  }

  :global(:root) {
    --sk-color: #1a73e8;
  }

  :global(header) {
    height: 100px;
  }

  :global(.content) {
    position: fixed;
    top: 0;
    padding-top: 90px;
    height: 100%;
    width: 100%;
  }

  :global(li > span) {
    outline: 0;
    display: inline-block;
    min-width: 10px;
    word-break: break-word;
  }

  :global(.list-group-item) {
    padding: 0;
  }

  :global(.list-group-flush > .list-group-item:first-child) {
    border-top-width: 1px;
  }

  :global(.list-group-flush > .list-group-item:first-child.selected) {
    border-top: 1px solid transparent;
  }

  :global(.list-group-flush > .list-group-item:last-child) {
    border-bottom-width: 1px;
  }

  :global(.list-group-item:hover) {
    box-shadow: 0 1px 2px 0 rgba(60, 64, 67, 0.302),
      0 1px 3px 1px rgba(60, 64, 67, 0.149);
    outline: 0;
    z-index: 2000;
  }

  :global(.task) {
    padding: 0.75rem 0;
    width: calc(100% - 178px);
  }

  :global(.created) {
    padding: 0.75rem 0;
    color: #5f6368;
    width: 82px;
    text-align: right;
  }

  :global(.icon) {
    font-family: "Material Icons";
    font-style: normal;
    font-size: 1.5rem;
    padding: 12px;
    line-height: normal;
    color: #5f6368;
    cursor: pointer;
    height: fit-content;
  }

  :global(.delete:hover) {
    color: #d93025;
  }

  :global(.swal) {
    margin: 8px 6px;
  }

  :global(.sortable-ghost) {
    opacity: 0;
  }

  @media (max-width: 900px) {
    .brand.user {
      padding-left: 90px;
    }

    .loading {
      left: 0;
      width: 100%;
    }

    :global(.content) {
      padding-left: 0 !important;
    }
  }
</style>
