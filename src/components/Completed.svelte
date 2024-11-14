<script lang="ts">
  import { confirm } from "../misc.svelte";
  import { mytasks } from "../task.svelte";
  import CompletedTask from "./CompletedTask.svelte";

  let {
    show = $bindable(),
  }: {
    show: boolean;
  } = $props();

  let index = $derived(
    mytasks.lists.findIndex((i) => i.list === mytasks.list.list),
  );

  const expand = (event: MouseEvent) => {
    const target = event.target as Element;
    if (!target.classList.contains("delete")) show = !show;
  };

  const empty = async () => {
    if (await confirm("All completed tasks"))
      if ((await mytasks.empty()) == 0) show = false;
  };
</script>

<div style="height: 100%">
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="completed" onclick={expand}>
    <span>
      Completed ({mytasks.lists[index].completed})
    </span>
    <i class="expand">{show ? "expand_more" : "expand_less"}</i>
    {#if show && mytasks.lists[index].completed}
      <i class="expand delete" onclick={empty}>delete</i>
    {/if}
  </div>
  {#if show}
    <ul class="list-group list-group-flush" style="height:calc(50% - 85px)">
      {#each mytasks.completed as task, i (task.id)}
        <CompletedTask bind:task={mytasks.completed[i]} />
      {/each}
      {#if mytasks.completed.length < mytasks.lists[index].completed}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li
          class="list-group-item"
          onclick={async () => await mytasks.getTasks(15)}
        >
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
