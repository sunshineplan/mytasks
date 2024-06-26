<script lang="ts">
  import type { ComponentType } from "svelte";
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Show from "./components/Show.svelte";
  import { fire, post } from "./misc";
  import { username as user, current, lists, init } from "./task";
  import { showSidebar, component, loading } from "./stores";

  const getInfo = async () => {
    loading.start();
    await init();
    loading.end();
  };
  const promise = getInfo();

  const components: { [component: string]: ComponentType } = {
    setting: Setting,
    show: Show,
  };

  const setting = () => {
    if (window.innerWidth <= 900) $showSidebar = false;
    $component = "setting";
  };

  const logout = async () => {
    const resp = await post(window.universal + "/logout", undefined, true);
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
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <span class="nav-link link" on:click={setting}>Setting</span>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
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
    <!-- svelte-ignore a11y-no-static-element-interactions -->
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

  @media (max-width: 900px) {
    .brand.user {
      padding-left: 90px;
    }

    .loading {
      left: 0;
      width: 100%;
    }
  }
</style>
