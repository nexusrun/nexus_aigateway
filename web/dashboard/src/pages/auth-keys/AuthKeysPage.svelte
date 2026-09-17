<script>
  // API Keys page: managed gateway API keys (create with one-time secret
  // reveal, label editing, permanent deactivation).
  import LoadingState from "$lib/components/molecules/LoadingState.svelte";
  import Icon from "$lib/components/atoms/Icon.svelte";
  import CopyButton from "$lib/components/atoms/CopyButton.svelte";
  import FilterInput from "$lib/components/molecules/FilterInput.svelte";
  import InactiveToggle from "$lib/components/molecules/InactiveToggle.svelte";
  import { router } from "$lib/stores/router.svelte.js";
  import { auth } from "$lib/stores/auth.svelte.js";
  import { authKeysStore as store } from "./authKeys.svelte.js";
  import AuthKeyEditor from "./AuthKeyEditor.svelte";
  import AuthKeyLabelsEditor from "./AuthKeyLabelsEditor.svelte";
  import AuthKeyAllowedModelsEditor from "./AuthKeyAllowedModelsEditor.svelte";
  import AuthKeyList from "./AuthKeyList.svelte";
  import { Plus, X } from "lucide";
  import * as m from "$lib/paraglide/messages.js";

  const PAGE = "auth-keys";

  // Re-fetch when the page becomes active or the API key / timezone changes.
  $effect(() => {
    void auth.refreshTick;
    if (router.page === PAGE) store.fetchKeys();
  });
</script>

<div>
  <div class="page-header">
    <h2>{m.api_keys_title()}</h2>
    <div class="page-header-controls">
      {#if store.available && !auth.authError}
        <button
          type="button"
          class="btn btn-primary btn-with-icon"
          disabled={store.formSubmitting}
          onclick={() => {
            if (!store.formSubmitting) store.openForm();
          }}
        >
          <Icon icon={Plus} class="table-icon-svg" />
          <span>{m.api_keys_create()}</span>
        </button>
      {/if}
    </div>
  </div>

  {#if !store.available && !auth.authError}
    <div class="alert alert-warning">{m.api_keys_unavailable()}</div>
  {/if}
  {#if store.error && !auth.authError && !store.formOpen}
    <p class="form-error" role="alert" aria-live="assertive">{store.error}</p>
  {/if}

  {#if store.available && !auth.authError}
    <p class="form-hint auth-keys-help-notice">
      {m.api_keys_help()}
    </p>
  {/if}

  {#if store.issuedValue}
    <section class="auth-key-issued-banner" role="status" aria-live="polite">
      <div class="auth-key-issued-header">
        <div>
          <strong>{m.api_keys_created_title()}</strong>
          <p>{m.api_keys_store_warning()}</p>
        </div>
        <button
          type="button"
          class="auth-key-issued-dismiss"
          aria-label={m.api_keys_done()}
          title={m.api_keys_done()}
          onclick={() => store.dismissIssuedKey()}
        >
          <Icon icon={X} width="18" height="18" />
        </button>
      </div>
      <div class="auth-key-issued-value-row">
        <code class="auth-key-issued-token">{store.issuedValue}</code>
        <CopyButton
          state={store.copyState}
          onclick={() => store.copyIssuedValue()}
        />
      </div>
      {#if store.copyState.error}
        <p class="form-error" role="alert" aria-live="assertive">
          {m.api_keys_copy_failed()}
        </p>
      {/if}
    </section>
  {/if}

  <AuthKeyEditor />
  <AuthKeyLabelsEditor />
  <AuthKeyAllowedModelsEditor />

  {#if store.loading && store.keys.length === 0}
    <LoadingState label={m.api_keys_loading()} />
  {/if}

  {#if store.keys.length > 0 && store.available}
    <div class="table-toolbar">
      <div class="table-toolbar-main">
        <FilterInput
          placeholder={m.api_keys_filter_placeholder()}
          label={m.api_keys_filter_label()}
          bind:value={store.filter}
        />
      </div>
      <div class="table-toolbar-actions">
        {#if store.userPathFilter}
          <span class="auth-keys-path-chip">
            <code>{store.userPathFilter}</code>
            <button
              type="button"
              class="auth-keys-path-chip-clear"
              aria-label={m.api_keys_path_filter_clear()}
              title={m.api_keys_path_filter_clear()}
              onclick={() => store.clearUserPathFilter()}
            >
              <Icon icon={X} width="12" height="12" />
            </button>
          </span>
        {/if}
        <InactiveToggle
          bind:checked={store.showInactive}
          label={m.api_keys_show_inactive()}
          count={store.inactiveCount}
        />
      </div>
    </div>
  {/if}

  {#if store.visibleKeys.length > 0 && store.available}
    <AuthKeyList />
  {/if}

  {#if store.keys.length > 0 && store.visibleKeys.length === 0 && store.available}
    <p class="empty-state">
      {m.api_keys_no_match()}{store.inactiveCount > 0 && !store.showInactive
        ? " " + m.api_keys_hidden({ count: store.inactiveCount })
        : ""}
    </p>
  {/if}

  {#if store.keys.length === 0 && !store.loading && !auth.authError && !store.error && store.available}
    <p class="empty-state">{m.api_keys_empty()}</p>
  {/if}
</div>

<style>
/* --- API Keys page --- */
.auth-keys-help-notice {
  margin-bottom: 20px;
}

.auth-key-issued-banner {
  margin-bottom: 20px;
  padding: 16px;
  border: 1px solid color-mix(in srgb, var(--success) 38%, var(--border));
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--success) 9%, var(--bg-surface));
  box-shadow: inset 0 1px 0 color-mix(in srgb, #fff 10%, transparent);
}

.auth-key-issued-header,
.auth-key-issued-value-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.auth-key-issued-header {
  justify-content: space-between;
  margin-bottom: 12px;
}

.auth-key-issued-header strong {
  color: color-mix(in srgb, var(--success) 78%, var(--text));
  font-size: 14px;
}

.auth-key-issued-header p {
  margin-top: 3px;
  color: var(--text-muted);
  font-size: 12px;
}

.auth-key-issued-dismiss {
  display: inline-flex;
  flex: 0 0 auto;
  padding: 5px;
  border: 0;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
}

.auth-key-issued-dismiss:hover {
  color: var(--text);
}

.auth-key-issued-token {
  flex: 1;
  min-width: 0;
  overflow-x: auto;
  padding: 9px 12px;
  border: 1px solid var(--border);
  border-radius: calc(var(--radius) - 2px);
  background: color-mix(in srgb, var(--bg) 82%, transparent);
  font-size: 13px;
  word-break: break-all;
}

@media (max-width: 520px) {
  .auth-key-issued-value-row {
    align-items: stretch;
    flex-direction: column;
  }
}

.auth-keys-path-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 6px 3px 8px;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: color-mix(in srgb, var(--accent) 10%, var(--bg));
  font-size: 12px;
}

.auth-keys-path-chip-clear {
  display: inline-flex;
  padding: 2px;
  border: 0;
  background: none;
  color: var(--text-muted);
  cursor: pointer;
}

.auth-keys-path-chip-clear:hover {
  color: var(--text);
}
</style>
