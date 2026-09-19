// API-key auth state. The key is stored in localStorage and sent as an
// Authorization bearer header on every admin request. A 401 flips needsAuth
// and opens the dialog; authGeneration invalidates responses that were issued
// under an older key so they cannot clobber fresh data ("stale auth").

import { authenticationResponseMetadata } from "./external-auth.js";
import { aigatewayPath } from "$lib/api/paths.js";

const API_KEY_STORAGE_KEY = "aigateway_api_key";

export function normalizeApiKey(value) {
  const key = String(value || "").trim();
  if (/^Bearer\s*$/i.test(key)) {
    return "";
  }
  const match = key.match(/^Bearer\s+(.+)$/i);
  return match ? match[1].trim() : key;
}

class AuthStore {
  apiKey = $state("");
  username = $state("");
  password = $state("");
  needsAuth = $state(false);
  authError = $state(false);
  // Optional specific error text for the auth dialog; empty = generic copy.
  authErrorMessage = $state("");
  dialogOpen = $state(false);
  // Provider-neutral browser auth metadata advertised by an extension.
  externalLoginURL = $state("");
  externalLogoutURL = $state("");
  externalUser = $state("");
  // True while the browser holds an admin username/password session cookie.
  passwordSession = $state(false);
  generation = $state(0);
  // Incremented whenever the whole dashboard should re-fetch (key change,
  // timezone change, runtime refresh). Pages watch this in an $effect.
  refreshTick = $state(0);

  init() {
    try {
      this.apiKey = normalizeApiKey(
        localStorage.getItem(API_KEY_STORAGE_KEY) || "",
      );
    } catch {
      this.apiKey = "";
    }
  }

  hasApiKey() {
    return normalizeApiKey(this.apiKey) !== "";
  }

  save() {
    this.apiKey = normalizeApiKey(this.apiKey);
    try {
      localStorage.setItem(API_KEY_STORAGE_KEY, this.apiKey);
    } catch {
      // Keep the in-memory key when storage is unavailable.
    }
  }

  selectExternalAuthentication() {
    this.apiKey = "";
    try {
      localStorage.removeItem(API_KEY_STORAGE_KEY);
    } catch {
      // The in-memory key is still cleared when storage is unavailable.
    }
    this.generation++;
    this.authError = false;
    this.authErrorMessage = "";
    this.needsAuth = true;
  }

  openDialog() {
    this.dialogOpen = true;
  }

  closeDialog() {
    this.dialogOpen = false;
  }

  // submit applies a newly entered key and triggers a global refresh.
  submit() {
    const apiKey = normalizeApiKey(this.apiKey);
    if (!apiKey) {
      this.apiKey = "";
      this.authError = true;
      this.authErrorMessage = "";
      this.needsAuth = true;
      this.openDialog();
      return false;
    }
    this.apiKey = apiKey;
    this.save();
    this.generation++;
    this.authError = false;
    this.authErrorMessage = "";
    this.needsAuth = false;
    this.closeDialog();
    this.refresh();
    return true;
  }

  async login() {
    this.authError = false;
    this.authErrorMessage = "";
    try {
      const response = await fetch(aigatewayPath("/admin/auth/login"), {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username: this.username, password: this.password }),
      });
      if (!response.ok) {
        this.authError = true;
        this.authErrorMessage = "Invalid username or password.";
        return false;
      }
      // An old API key would take precedence over the browser session.
      this.apiKey = "";
      try {
        localStorage.removeItem(API_KEY_STORAGE_KEY);
      } catch {
        // Storage may be unavailable in restricted browser contexts.
      }
      this.password = "";
      this.passwordSession = true;
      this.generation++;
      this.needsAuth = false;
      this.refresh();
      return true;
    } catch {
      this.authError = true;
      this.authErrorMessage = "Unable to reach the authentication service.";
      return false;
    }
  }

  // checkSession asks the gateway whether the session cookie is still valid.
  // Gateways without password login answer 404, which leaves it false.
  async checkSession() {
    try {
      const response = await fetch(aigatewayPath("/admin/auth/session"), {
        credentials: "same-origin",
      });
      if (!response.ok) return;
      const body = await response.json();
      this.passwordSession = Boolean(body && body.authenticated === true);
    } catch {
      // Keep the current state when the gateway is unreachable.
    }
  }

  // logout clears the session cookie and returns to the login screen. On
  // failure the dashboard stays signed in, since a reload would restore the
  // cookie anyway.
  async logout() {
    try {
      const response = await fetch(aigatewayPath("/admin/auth/logout"), {
        method: "POST",
        credentials: "same-origin",
      });
      if (!response.ok) return false;
    } catch {
      return false;
    }
    this.passwordSession = false;
    this.username = "";
    this.password = "";
    this.generation++;
    this.authError = false;
    this.authErrorMessage = "";
    this.needsAuth = true;
    return true;
  }

  refresh() {
    this.refreshTick++;
  }

  observeResponse(response) {
    const metadata = authenticationResponseMetadata(response);
    if (!metadata) return;
    if (Object.hasOwn(metadata, "loginURL")) {
      this.externalLoginURL = metadata.loginURL;
    }
    if (Object.hasOwn(metadata, "logoutURL")) {
      this.externalLogoutURL = metadata.logoutURL;
    }
    if (Object.hasOwn(metadata, "user")) {
      this.externalUser = metadata.user;
    }
  }

  // handleUnauthorized reports whether the 401/403 belongs to the current
  // key. Stale responses (older generation) are ignored silently. An optional
  // message replaces the generic dialog error text (e.g. a valid key that
  // lacks dashboard access).
  handleUnauthorized(requestGeneration, message = "") {
    if (
      typeof requestGeneration === "number" &&
      requestGeneration < this.generation
    ) {
      return false;
    }
    this.authError = true;
    this.authErrorMessage = message;
    this.needsAuth = true;
    this.openDialog();
    return true;
  }
}

export const auth = new AuthStore();
