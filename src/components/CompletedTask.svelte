<script lang="ts">
  import { confirm } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let {
    task = $bindable(),
  }: {
    task: Task;
  } = $props();

  let hover = $state(false);

  const del = async () => {
    if (await confirm("This completed task"))
      await mytasks.deleteTask(task, true);
  };
</script>

<li
  class="list-group-item"
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <i class="icon revert" onclick={async () => await mytasks.revertTask(task)}
    >done</i
  >
  <span class="task">{task.task}</span>
  <span class="created">{new Date(task.created).toLocaleDateString()}</span>
  {#if hover}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <i class="icon delete" onclick={del}>delete</i>
  {/if}
</li>

<style>
  li {
    display: inline-flex;
  }

  .revert {
    content: "done";
    color: #468dff;
  }

  .revert:hover {
    background-color: #e6ecf0;
    border-radius: 50%;
  }

  .task {
    text-decoration: line-through;
  }
</style>
