<script lang="ts">
  import { BootstrapButtons, post } from "../misc";
  import { username as user } from "../stores";

  let username = "";
  let password = "";
  let rememberme = false;

  const login = async () => {
    if (
      !(document.querySelector(
        "#username"
      ) as HTMLSelectElement).checkValidity()
    )
      await BootstrapButtons.fire(
        "Error",
        "Username cannot be empty.",
        "error"
      );
    else if (
      !(document.querySelector(
        "#password"
      ) as HTMLSelectElement).checkValidity()
    )
      await BootstrapButtons.fire(
        "Error",
        "Password cannot be empty.",
        "error"
      );
    else {
      const resp = await post("/login", {
        username,
        password,
        rememberme,
      });
      if (!resp.ok)
        await BootstrapButtons.fire("Error", await resp.text(), "error");
      else user.set(username);
    }
  };
</script>

<style>
  .login {
    width: 250px;
    margin: 0 auto 20px;
  }
</style>

<svelte:head>
  <title>Log In - My Tasks</title>
</svelte:head>

<div class="content">
  <header>
    <h3
      class="d-flex justify-content-center align-items-center"
      style="height: 100%">
      Log In
    </h3>
  </header>
  <div class="login">
    <div class="form-group">
      <label for="username">Username</label>
      <input
        class="form-control"
        bind:value={username}
        id="username"
        maxlength="20"
        placeholder="Username"
        required />
    </div>
    <div class="form-group">
      <label for="password">Password</label>
      <input
        class="form-control"
        type="password"
        bind:value={password}
        id="password"
        maxlength="20"
        placeholder="Password"
        required />
    </div>
    <div class="form-group form-check">
      <input
        type="checkbox"
        class="form-check-input"
        bind:checked={rememberme}
        id="rememberme" />
      <label class="form-check-label" for="rememberme">Remember Me</label>
    </div>
    <hr />
    <button class="btn btn-primary login" on:click={login}>Log In</button>
  </div>
</div>
