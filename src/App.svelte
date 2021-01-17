<script lang="ts">
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Show from "./components/Show.svelte";
  import {
    username as user,
    current,
    showSidebar,
    component,
    loading,
    lists,
  } from "./stores";

  const getInfo = async () => {
    const resp = await fetch("/info");
    const info = await resp.json();
    if (Object.keys(info).length) {
      $user = info.username;
      $lists = info.lists;
    } else return;
    if ($lists.length) if (!$current.id) $current = $lists[0];
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
</script>

<nav class="navbar navbar-light topbar">
  <div class="d-flex" style="height: 100%">
    <a class="brand" href="/">My Tasks</a>
  </div>
  {#if $user}
    <div class="navbar-nav flex-row">
      <span class="nav-link">{$user}</span>
      <span class="nav-link link" on:click={setting}>Setting</span>
      <a class="nav-link link" href="/logout">Log Out</a>
    </div>
  {:else}
    <div class="navbar-nav flex-row"><span class="nav-link">Log In</span></div>
  {/if}
</nav>
{#await promise then _}
  {#if !$user}
    <Login on:info={getInfo} />
  {:else}
    <Sidebar />
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}"
      on:mousedown={() => ($showSidebar = false)}
    >
      <svelte:component this={components[$component]} />
    </div>
  {/if}
{/await}
<div class="loading" hidden={!$loading}>
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
  }

  .brand:hover {
    color: white;
    text-decoration: none;
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

  @media (max-width: 900px) {
    .brand {
      padding-left: 90px;
    }

    .loading {
      left: 0;
      width: 100%;
    }
  }
</style>
