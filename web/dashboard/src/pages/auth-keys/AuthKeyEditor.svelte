<script>
  // Create-API-key modal (EditorDialog shell). A successful creation closes
  // this editor; the one-time value is shown by the page-level success banner.
  import EditorDialog from "$lib/components/organisms/EditorDialog.svelte";
  import FormField from "$lib/components/molecules/FormField.svelte";
  import InlineHelpSection from "$lib/components/molecules/InlineHelpSection.svelte";
  import SearchSelect from "$lib/components/molecules/SearchSelect.svelte";
  import { modelsStore } from "$lib/stores/models.svelte.js";
  import { authKeysStore as store } from "./authKeys.svelte.js";
  import { authKeySelectorOptions } from "./authKeysLogic.js";
  import { Plus } from "lucide";
  import * as m from "$lib/paraglide/messages.js";

  const selectorOptions = $derived(authKeySelectorOptions(modelsStore.models));
</script>

<EditorDialog
  open={store.formOpen}
  title={m.api_keys_create()}
  ariaLabel={m.api_keys_editor()}
  error={store.error}
  submitting={store.formSubmitting}
  submitLabel={m.api_keys_create()}
  submittingLabel={m.api_keys_creating()}
  submitIcon={Plus}
  dialogClass="auth-key-editor"
  onclose={() => store.closeForm()}
  onsubmit={() => store.submitForm()}
>
  <div class="auth-key-form-fields">
      <div class="form-grid">
        <div class="form-field">
          <label class="form-field-label" for="auth-key-name">
            {m.api_keys_name()} <span class="form-hint">({m.api_keys_required()})</span>
          </label>
          <input
            id="auth-key-name"
            type="text"
            placeholder="e.g. ci-deploy"
            autocomplete="off"
            data-modal-autofocus
            bind:value={store.form.name}
          />
        </div>
        <div class="form-field">
          <label class="form-field-label" for="auth-key-expires">
            {m.api_keys_expires()} <span class="form-hint">({m.api_keys_expiry_help()})</span>
          </label>
          <input id="auth-key-expires" type="date" bind:value={store.form.expires_at} />
        </div>
      </div>
      <div class="form-field">
        <InlineHelpSection copyId="auth-key-user-path-help-copy" label={m.api_keys_user_path_help_label()}>
          {#snippet title()}
            <label class="form-field-label" for="auth-key-user-path">{m.api_keys_user_path()}</label>
          {/snippet}
          {#snippet help()}
            {m.api_keys_user_path_help()}
          {/snippet}
        </InlineHelpSection>
        <input
          id="auth-key-user-path"
          type="text"
          placeholder="ex. /department1/team-a"
          aria-describedby="auth-key-user-path-help-copy"
          bind:value={store.form.user_path}
        />
      </div>
      <div class="form-field">
        <InlineHelpSection copyId="auth-key-labels-help-copy" label={m.api_keys_labels_help_label()}>
          {#snippet title()}
            <label class="form-field-label" for="auth-key-labels">
              {m.api_keys_labels()}
            </label>
          {/snippet}
          {#snippet help()}
            {m.api_keys_labels_help()}
          {/snippet}
        </InlineHelpSection>
        <input
          id="auth-key-labels"
          type="text"
          placeholder="ex. team-a, batch-jobs"
          aria-describedby="auth-key-labels-help-copy"
          bind:value={store.form.labels}
        />
      </div>
      <div class="form-field">
        <InlineHelpSection copyId="auth-key-allowed-models-help-copy" label={m.api_keys_allowed_models_help_label()}>
          {#snippet title()}
            <label class="form-field-label" for="auth-key-allowed-models">
              {m.api_keys_allowed_models()}
            </label>
          {/snippet}
          {#snippet help()}
            {m.api_keys_allowed_models_help()}
          {/snippet}
        </InlineHelpSection>
        <div class="auth-key-allowed-models-select">
          <SearchSelect
            id="auth-key-allowed-models"
            options={selectorOptions}
            multiple
            bind:values={store.form.allowed_models}
            placeholder={m.model_selectors_placeholder()}
            searchPlaceholder={m.model_selectors_search()}
            ariaLabel={m.api_keys_allowed_models()}
            allowCustom
            mono
          />
        </div>
      </div>
      <div class="form-field">
        <InlineHelpSection copyId="auth-key-dashboard-access-help-copy" label={m.api_keys_dashboard_help_label()}>
          {#snippet title()}
            <label class="form-field-label" for="auth-key-dashboard-access">{m.api_keys_dashboard_access()}</label>
          {/snippet}
          {#snippet help()}
            {m.api_keys_dashboard_help()}
          {/snippet}
        </InlineHelpSection>
        <label class="auth-key-dashboard-toggle">
          <input
            id="auth-key-dashboard-access"
            type="checkbox"
            aria-describedby="auth-key-dashboard-access-help-copy"
            bind:checked={store.form.dashboard_access}
          />
          <span>{m.api_keys_dashboard_allow()}</span>
        </label>
      </div>
      <FormField id="auth-key-description" label={m.api_keys_description_optional()}>
        <textarea
          id="auth-key-description"
          rows="2"
          placeholder={m.api_keys_description_placeholder()}
          bind:value={store.form.description}
        ></textarea>
      </FormField>
  </div>
</EditorDialog>

<style>
  .auth-key-form-fields > :global(.form-field) {
    margin-bottom: 4px;
  }

  .auth-key-allowed-models-select :global(.search-select) {
    display: flex;
    width: 100%;
  }

  .auth-key-dashboard-toggle {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    cursor: pointer;
  }

</style>
