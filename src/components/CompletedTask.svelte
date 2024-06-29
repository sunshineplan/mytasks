<script lang="ts">
  import { confirm } from "../misc";
  import { tasks } from "../task";

  export let task: Task;
  let hover = false;

  const del = async () => {
    if (await confirm("This completed task")) await tasks.delete(task, true);
  };
</script>

<li
  class="list-group-item"
  on:mouseenter={() => (hover = true)}
  on:mouseleave={() => (hover = false)}
>
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <i class="icon revert" on:click={async () => await tasks.revert(task)}>done</i
  >
  <span class="task">{task.task}</span>
  <span class="created">{new Date(task.created).toLocaleDateString()}</span>
  {#if hover}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <i class="icon delete" on:click={del}>delete</i>
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
