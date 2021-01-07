<script lang="ts">
  import Login from "./components/Login.svelte";
  import Setting from "./components/Setting.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Tasks from "./components/Tasks.svelte";
  import { post } from "./misc";
  import { username as user, component, tasks, loading, List } from "./stores";

  let lists: List[];

  const getLists = async () => {
    let resp = await post("/list/get");
    lists = await resp.json();
    lists.sort((a, b) => a.seq - b.seq);
    resp = await post("/task/get", { list: lists[0].id });
    $tasks[lists[0].list] = await resp.json();
  };

  const promise = getLists();

  const components = {
    setting: Setting,
    tasks: Tasks,
  } as any;

  const setting = () => {
    component.set("setting");
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
{#if !$user}
  <Login />
{:else}
  {#await promise then _}
    <Sidebar bind:lists />
    <div
      class="content"
      style="padding-left: 250px; opacity: {$loading ? 0.5 : 1}">
      <svelte:component this={components[$component]} />
    </div>
  {/await}
{/if}
<div class="loading" hidden={!$loading}>
  <div class="sk-wave sk-center">
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
    <div class="sk-wave-rect" />
  </div>
</div>
