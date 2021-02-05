<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, post, valid } from "../misc";
  import { component } from "../stores";

  const dispatch = createEventDispatcher();

  let password = "";
  let password1 = "";
  let password2 = "";
  let validated = false;

  const setting = async () => {
    if (valid()) {
      validated = false;
      const resp = await post("@universal@/chgpwd", {
        password,
        password1,
        password2,
      });
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        const json = await resp.json();
        if (json.status == 1) {
          await fire(
            "Success",
            "Your password has changed. Please Re-login!",
            "success"
          );
          dispatch("reload");
          $component = "show";
        } else {
          await fire("Error", json.message, "error");
          if (json.error == 1) password = "";
          else {
            password1 = "";
            password2 = "";
          }
        }
      }
    } else validated = true;
  };

  const cancel = () => {
    $component = "show";
  };

  const handleEscape = (event: KeyboardEvent) => {
    if (event.key === "Escape") cancel();
  };
  const handleEnter = async (event: KeyboardEvent) => {
    if (event.key === "Enter") await setting();
  };
</script>

<svelte:head>
  <title>Setting - My Tasks</title>
</svelte:head>

<svelte:window on:keydown={handleEscape} />

<div on:keyup={handleEnter}>
  <header style="padding-left: 20px">
    <h3>Setting</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="form-group">
      <label for="password">Current Password</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        class="form-control"
        type="password"
        bind:value={password}
        id="password"
        maxlength="20"
        autofocus
        required
      />
      <div class="invalid-feedback">This field is required.</div>
    </div>
    <div class="form-group">
      <label for="password1">New Password</label>
      <input
        class="form-control"
        type="password"
        bind:value={password1}
        id="password1"
        maxlength="20"
        required
      />
      <div class="invalid-feedback">This field is required.</div>
    </div>
    <div class="form-group">
      <label for="password2">Confirm Password</label>
      <input
        class="form-control"
        type="password"
        bind:value={password2}
        id="password2"
        maxlength="20"
        required
      />
      <div class="invalid-feedback">This field is required.</div>
      <small class="form-text text-muted">
        Max password length: 20 characters.
      </small>
    </div>
    <button class="btn btn-primary" on:click={setting}>Change</button>
    <button class="btn btn-primary" on:click={cancel}>Cancel</button>
  </div>
</div>

<style>
  .form {
    padding: 0 20px;
  }

  .form-control {
    width: 250px;
  }
</style>
