<script lang="ts">
  import JSEncrypt from "jsencrypt";
  import { createEventDispatcher } from "svelte";
  import { fire, post } from "../misc";

  const dispatch = createEventDispatcher();

  let username = "";
  let password = "";
  let rememberme = localStorage.getItem("rememberme") == "true" ? true : false;
  let pubkey = "@pubkey@";

  const login = async () => {
    if (
      !(
        document.querySelector("#username") as HTMLSelectElement
      ).checkValidity()
    )
      await fire("Error", "Username cannot be empty.", "error");
    else if (
      !(
        document.querySelector("#password") as HTMLSelectElement
      ).checkValidity()
    )
      await fire("Error", "Password cannot be empty.", "error");
    else {
      var pwd: string;
      if (pubkey) pwd = encrypt(password) as string;
      else pwd = password;
      const resp = await post(
        "@universal@/login",
        {
          username,
          password: pwd,
          rememberme,
        },
        true
      );
      if (resp.ok) {
        const json = await resp.json();
        if (json.status == 1) {
          if (rememberme) localStorage.setItem("rememberme", "true");
          else localStorage.removeItem("rememberme");
          dispatch("info");
        } else await fire("Error", json.message, "error");
      } else await fire("Error", await resp.text(), "error");
    }
  };

  const encrypt = (password: string) => {
    const encrypt = new JSEncrypt();
    encrypt.setPublicKey(pubkey);
    return encrypt.encrypt(password);
  };

  const handleEnter = async (event: KeyboardEvent) => {
    if (event.key === "Enter") await login();
  };
</script>

<svelte:head>
  <title>Log In - My Tasks</title>
</svelte:head>

<div class="content" on:keyup={handleEnter}>
  <header>
    <h3
      class="d-flex justify-content-center align-items-center"
      style="height: 100%"
    >
      Log In
    </h3>
  </header>
  <div class="login">
    <div class="mb-3">
      <label class="form-label" for="username">Username</label>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        class="form-control"
        bind:value={username}
        id="username"
        maxlength="20"
        placeholder="Username"
        autofocus
        required
      />
    </div>
    <div class="mb-3">
      <label class="form-label" for="password">Password</label>
      <input
        class="form-control"
        type="password"
        bind:value={password}
        id="password"
        maxlength="20"
        placeholder="Password"
        required
      />
    </div>
    <div class="form-group form-check">
      <input
        type="checkbox"
        class="form-check-input"
        bind:checked={rememberme}
        id="rememberme"
      />
      <label class="form-check-label" for="rememberme">Remember Me</label>
    </div>
    <hr />
    <button class="btn btn-primary login" on:click={login}>Log In</button>
  </div>
</div>

<style>
  .login {
    width: 250px;
    margin: 0 auto 20px;
  }

  .form-control {
    width: 250px;
  }
</style>
