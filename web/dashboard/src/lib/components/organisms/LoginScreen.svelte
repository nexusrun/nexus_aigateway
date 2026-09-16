<script>
  import GoModelLogo from "$lib/components/atoms/GoModelLogo.svelte";
  import Icon from "$lib/components/atoms/Icon.svelte";
  import { auth } from "$lib/stores/auth.svelte.js";
  import { authenticationLoginURL } from "$lib/stores/external-auth.js";
  import { gomodelPath } from "$lib/api/paths.js";
  import * as m from "$lib/paraglide/messages.js";
  import { ArrowRight, Check, KeyRound, LockKeyhole } from "lucide";
</script>

<main class="login-screen">
  <div class="login-glow login-glow-one"></div>
  <div class="login-glow login-glow-two"></div>
  <section class="login-card" aria-labelledby="login-title">
    <div class="login-brand">
      <div class="login-logo"><GoModelLogo variant="full" /></div>
    </div>
    <div class="login-heading">
      <p class="login-kicker">SECURE CONTROL PLANE</p>
      <h1 id="login-title">{m.auth_dialog_locked_title()}</h1>
      <p>{m.auth_banner_required()}</p>
    </div>

    <form
      class="login-form"
      onsubmit={(event) => {
        event.preventDefault();
        void auth.login();
      }}
    >
      {#if auth.externalLoginURL}
        <a
          class="btn btn-primary btn-with-icon login-sso"
          href={authenticationLoginURL(gomodelPath(auth.externalLoginURL))}
          onclick={() => auth.selectExternalAuthentication()}
        >
          <Icon icon={KeyRound} />
          <span>{m.auth_dialog_sign_in_with_sso()}</span>
          <Icon icon={ArrowRight} class="login-arrow" />
        </a>
        <div class="login-separator"><span>{m.auth_dialog_or_use_api_key()}</span></div>
      {/if}
      <label class="login-label" for="loginUsername">Username</label>
      <input
        id="loginUsername"
        class="login-plain-input"
        type="text"
        autocomplete="username"
        bind:value={auth.username}
      />
      <label class="login-label" for="loginPassword">Password</label>
      <div class="login-input-shell">
        <Icon icon={LockKeyhole} class="login-input-icon" />
        <input
          id="loginPassword"
          type="password"
          autocomplete="current-password"
          data-modal-autofocus
          bind:value={auth.password}
        />
      </div>
      {#if auth.authError}
        <p class="login-error" role="alert">
          {auth.authErrorMessage || m.auth_api_key_invalid()}
        </p>
      {/if}
      <button type="submit" class="btn btn-primary btn-with-icon login-submit">
        <Icon icon={Check} />
        <span>{m.auth_action_unlock_dashboard()}</span>
      </button>
    </form>
    <p class="login-hint">{m.auth_api_key_storage_hint()}</p>
  </section>
</main>

<style>
  .login-screen {
    position: relative;
    display: grid;
    place-items: center;
    width: 100%;
    min-height: 100vh;
    overflow: hidden;
    background:
      radial-gradient(circle at 50% 20%, color-mix(in srgb, var(--accent) 14%, transparent), transparent 28rem),
      var(--bg);
  }

  .login-glow {
    position: absolute;
    width: 28rem;
    height: 28rem;
    border-radius: 50%;
    filter: blur(90px);
    opacity: 0.16;
    pointer-events: none;
  }

  .login-glow-one {
    top: -14rem;
    left: -10rem;
    background: var(--accent);
  }

  .login-glow-two {
    right: -12rem;
    bottom: -16rem;
    background: var(--info);
  }

  .login-card {
    position: relative;
    z-index: 1;
    width: min(430px, calc(100% - 32px));
    padding: 36px;
    border: 1px solid color-mix(in srgb, var(--border) 65%, var(--accent) 35%);
    border-radius: 18px;
    background:
      linear-gradient(145deg, color-mix(in srgb, var(--bg-surface) 94%, #fff 6%), color-mix(in srgb, var(--bg-surface) 84%, var(--accent) 16%));
    box-shadow:
      0 28px 80px color-mix(in srgb, #000 34%, transparent),
      0 0 70px color-mix(in srgb, var(--accent) 12%, transparent),
      inset 0 1px 0 color-mix(in srgb, #fff 18%, transparent);
    backdrop-filter: blur(22px);
  }

  .login-brand {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text);
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 0.8px;
    text-transform: uppercase;
  }

  .login-logo {
    width: 190px;
    height: 190px;
    filter: drop-shadow(0 0 18px color-mix(in srgb, var(--accent) 44%, transparent));
  }

  .login-logo :global(img) {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .login-heading {
    margin: 22px 0 26px;
  }

  .login-kicker {
    margin-bottom: 10px;
    color: var(--accent);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 2px;
  }

  .login-heading h1 {
    font-size: 30px;
    letter-spacing: -0.8px;
  }

  .login-heading p:last-child,
  .login-hint {
    margin-top: 8px;
    color: var(--text-muted);
    font-size: 13px;
  }

  .login-form {
    display: grid;
    gap: 12px;
  }

  .login-sso,
  .login-submit {
    width: 100%;
    justify-content: center;
    text-decoration: none;
  }

  .login-sso :global(.login-arrow) {
    width: 16px;
    margin-left: auto;
  }

  .login-separator {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 4px 0;
    color: var(--text-muted);
    font-size: 11px;
  }

  .login-separator::before,
  .login-separator::after {
    content: "";
    flex: 1;
    border-top: 1px solid var(--border);
  }

  .login-label {
    color: var(--text-muted);
    font-size: 12px;
    font-weight: 600;
  }

  .login-input-shell {
    position: relative;
  }

  .login-input-shell input {
    width: 100%;
    padding: 12px 14px 12px 40px;
    background: color-mix(in srgb, var(--bg) 76%, transparent);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text);
    font: inherit;
    outline: none;
  }

  .login-plain-input {
    width: 100%;
    padding: 12px 14px;
    background: color-mix(in srgb, var(--bg) 76%, transparent);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    color: var(--text);
    font: inherit;
    outline: none;
  }

  .login-plain-input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 18%, transparent);
  }

  .login-input-shell input:focus {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 18%, transparent);
  }

  .login-input-shell :global(.login-input-icon) {
    position: absolute;
    top: 50%;
    left: 13px;
    width: 16px;
    height: 16px;
    color: var(--text-muted);
    pointer-events: none;
    transform: translateY(-50%);
  }

  .login-error {
    color: var(--danger);
    font-size: 12px;
  }

  .login-hint {
    text-align: center;
  }

  @media (max-width: 520px) {
    .login-card {
      padding: 28px 22px;
    }
  }
</style>
