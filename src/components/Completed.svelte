<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CompletedTask from "./CompletedTask.svelte";
  import { fire, confirm, post } from "../misc";
  import { current, loading, tasks } from "../stores";
  import type { Task } from "../stores";

  const dispatch = createEventDispatcher();

  export let show = false;
  export let completedTasks: Task[] = [];

  const expand = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (!target.classList.contains("delete")) show = !show;
  };

  const empty = async () => {
    if (await confirm("These completed tasks")) {
      $loading++;
      const resp = await post("/completed/empty/" + $current.id);
      const json = await resp.json();
      $loading--;
      if (json.status) {
        $tasks[$current.list].completed = [];
        dispatch("refresh");
        show = false;
      } else await fire("Error", "Error", "error");
    }
  };
</script>

<div style="height: 100%">
  <div class="completed" on:click={expand}>
    <span>Completed ({completedTasks.length})</span>
    <i class="expand">{show ? "expand_more" : "expand_less"}</i>
    {#if show && completedTasks.length}
      <i class="expand delete" on:click={empty}>delete</i>
    {/if}
  </div>
  {#if show}
    <ul class="list-group list-group-flush" style="height:calc(50% - 85px)">
      {#each completedTasks as task (task.id)}
        <CompletedTask bind:task on:refresh />
      {/each}
    </ul>
  {/if}
</div>

<style>
  ul {
    cursor: default;
    overflow-y: auto;
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
</style>
