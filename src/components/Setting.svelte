<script  lang="ts">
  import { BootstrapButtons, post, valid } from "../misc";
  import { username, component } from "../stores";

  let password = "";
  let password1 = "";
  let password2 = "";
  let validated = false;

  const setting = async () => {
    if (valid()) {
      validated = false;
      const resp = await post("/setting", {
        password,
        password1,
        password2,
      });
      if (!resp.ok)
        await BootstrapButtons.fire("Error", await resp.text(), "error");
      else {
        const json = await resp.json();
        if (json.status == 1) {
          await BootstrapButtons.fire(
            "Success",
            "Your password has changed. Please Re-login!",
            "success"
          );
          username.set("");
        } else {
          await BootstrapButtons.fire("Error", json.message, "error");
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
    component.set("tasks");
  };
</script>

<svelte:head>
  <title>Setting - My Tasks</title>
</svelte:head>

<svelte:window on:keydown={cancel} />

<div on:keyup={setting}>
  <header style="padding-left: 20px">
    <h3>Setting</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="form-group">
      <label for="password">Current Password</label>
      <input
        class="form-control"
        type="password"
        bind:value={password}
        id="password"
        maxlength="20"
        required />
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
        required />
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
        required />
      <div class="invalid-feedback">This field is required.</div>
      <small class="form-text text-muted">Max password length: 20 characters.</small>
    </div>
    <button class="btn btn-primary" on:click={setting}>Change</button>
    <button class="btn btn-primary" on:click={cancel}>Cancel</button>
  </div>
</div>
