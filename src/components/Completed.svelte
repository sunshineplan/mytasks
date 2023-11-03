<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CompletedTask from "./CompletedTask.svelte";
  import { fire, confirm, post } from "../misc";
  import { current, tasks, lists } from "../stores";

  const dispatch = createEventDispatcher();

  export let show = false;
  export let completedTasks: Task[] = [];

  $: index = $lists.findIndex((list) => list.list === $current.list);

  const expand = (event: MouseEvent) => {
    const target = event.target as Element;
    if (!target.classList.contains("delete")) show = !show;
  };

  const empty = async () => {
    if (await confirm("All completed tasks")) {
      const resp = await post("/completed/empty", { list: $current.list });
      if (resp.ok) {
        const json = await resp.json();
        if (json.status) {
          $lists[index].completed = 0;
          $tasks[$current.list].completed = [];
          dispatch("refresh");
          show = false;
        } else await fire("Error", "Error", "error");
      } else await fire("Error", await resp.text(), "error");
    }
  };

  const more = async () => {
    const resp = await post("/completed/more", {
      list: $current.list,
      start: completedTasks.length,
    });
    if (resp.ok) {
      $tasks[$current.list].completed = completedTasks.concat(
        await resp.json()
      );
      dispatch("refresh");
    } else await fire("Error", await resp.text(), "error");
  };
</script>

<div style="height: 100%">
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="completed" on:click={expand}>
    <span>
      Completed ({$lists[index].completed})
    </span>
    <i class="expand">{show ? "expand_more" : "expand_less"}</i>
    {#if show && $lists[index].completed}
      <i class="expand delete" on:click={empty}>delete</i>
    {/if}
  </div>
  {#if show}
    <ul class="list-group list-group-flush" style="height:calc(50% - 85px)">
      {#each completedTasks as task (task.id)}
        <CompletedTask bind:task on:refresh on:reload />
      {/each}
      {#if completedTasks.length < $lists[index].completed}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
        <li class="list-group-item" on:click={more}>
          <i class="icon">sync</i>
          <span class="load">Load more</span>
        </li>
      {/if}
    </ul>
  {/if}
</div>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
  }

  li {
    display: inline-flex;
  }

  .completed {
    padding: 15px 20px;
    margin-top: 16px;
    font-weight: 500;
    color: #5f6368;
    cursor: pointer;
    background-color: rgba(0, 0, 0, 0.125);
  }

  .expand {
    font-family: "Material Icons";
    font-style: normal;
    font-size: 1.5rem;
    float: right;
    line-height: normal;
  }

  .load {
    padding: 0.75rem 0;
    color: #5f6368;
    font-weight: bold;
  }
</style>
