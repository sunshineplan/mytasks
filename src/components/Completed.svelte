<script lang="ts">
  import CompletedTask from "./CompletedTask.svelte";
  import { confirm } from "../misc";
  import { list, tasks, lists } from "../task";

  export let show = false;

  $: index = $lists.findIndex((i) => i.list === $list.list);

  const expand = (event: MouseEvent) => {
    const target = event.target as Element;
    if (!target.classList.contains("delete")) show = !show;
  };

  const empty = async () => {
    if (await confirm("All completed tasks"))
      if ((await tasks.empty()) == 0) show = false;
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
      {#each $tasks.completed as task (task.id)}
        <CompletedTask bind:task />
      {/each}
      {#if $tasks.completed.length < $lists[index].completed}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
        <li class="list-group-item" on:click={async () => await tasks.get(15)}>
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
