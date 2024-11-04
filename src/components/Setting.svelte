<script lang="ts">
  import { encrypt, fire, post, valid } from "../misc.svelte";
  import { mytasks } from "../task.svelte";

  let password = $state("");
  let password1 = $state("");
  let password2 = $state("");
  let validated = $state(false);

  const setting = async () => {
    if (valid()) {
      validated = false;
      var pwd: string, p1: string, p2: string;
      if (window.pubkey && window.pubkey.length) {
        pwd = encrypt(window.pubkey, password);
        p1 = encrypt(window.pubkey, password1);
        p2 = encrypt(window.pubkey, password2);
      } else {
        pwd = password;
        p1 = password1;
        p2 = password2;
      }
      const resp = await post(
        window.universal + "/chgpwd",
        {
          password: pwd,
          password1: p1,
          password2: p2,
        },
        true,
      );
      if (resp.ok) {
        const json = await resp.json();
        if (json.status == 1) {
          await fire(
            "Success",
            "Your password has changed. Please Re-login!",
            "success",
          );
          await mytasks.init();
          mytasks.component = "show";
        } else {
          await fire("Error", json.message, "error");
          if (json.error == 1) password = "";
          else {
            password1 = "";
            password2 = "";
          }
        }
      } else await fire("Error", await resp.text(), "error");
    } else validated = true;
  };

  const cancel = () => {
    mytasks.component = "show";
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

<svelte:window onkeydown={handleEscape} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div onkeyup={handleEnter}>
  <header style="padding-left: 20px">
    <h3>Setting</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="mb-3">
      <label class="form-label" for="password">Current Password</label>
      <!-- svelte-ignore a11y_autofocus -->
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
    <div class="mb-3">
      <label class="form-label" for="password1">New Password</label>
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
    <div class="mb-3">
      <label class="form-label" for="password2">Confirm Password</label>
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
    <button class="btn btn-primary" onclick={setting}>Change</button>
    <button class="btn btn-primary" onclick={cancel}>Cancel</button>
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
