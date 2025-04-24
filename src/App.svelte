<script lang="ts">
  import type { Component } from "svelte";
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Show from "./components/Show.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import { fire, post } from "./misc.svelte";
  import { loading } from "./misc.svelte";
  import { mytasks } from "./task.svelte";

  const promise = mytasks.init();

  const components: { [component: string]: Component } = {
    setting: Setting,
    show: Show,
  };

  const Content = $derived(components[mytasks.component]);

  const setting = () => {
    mytasks.component = "setting";
  };

  const logout = async () => {
    mytasks.abort();
    const resp = await post(window.universal + "/logout", undefined, true);
    if (resp.ok) {
      await mytasks.init();
      window.history.pushState({}, "", "/");
      mytasks.component = "show";
    } else await fire("Error", "Unknow error", "error");
  };
</script>

<nav class="navbar navbar-light topbar">
  <div class="d-flex" style="height: 100%">
    <a class="brand" class:user={mytasks.username} href="/">My Tasks</a>
  </div>
  {#if mytasks.username}
    <div class="navbar-nav flex-row">
      <span class="nav-link">{mytasks.username}</span>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <span class="nav-link link" onclick={setting}>Setting</span>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <span class="nav-link link" onclick={logout}>Logout</span>
    </div>
  {:else}
    <div class="navbar-nav flex-row"><span class="nav-link">Log In</span></div>
  {/if}
</nav>
{#await promise then _}
  {#if !mytasks.username}
    {#if !loading.show}
      <Login />
    {/if}
  {:else}
    <Sidebar />
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="content"
      style="padding-left: 250px; opacity: {loading.show ? 0.5 : 1}"
    >
      <Content />
    </div>
  {/if}
{/await}
<div
  class={mytasks.username ? "loading" : "initializing"}
  hidden={!loading.show}
>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
    <div class="sk-wave-rect"></div>
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
