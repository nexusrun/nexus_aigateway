// Wiring tests for password-session logout. The auth store uses Svelte runes
// and there is no DOM harness in this suite, so these assert the contract on
// the source (like editor-dialog.test.js).
import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";

const SRC = fileURLToPath(new URL("../src", import.meta.url));
const read = (path) => readFileSync(join(SRC, path), "utf8");

function method(source, name) {
  const match = source.match(new RegExp(`\\n  async ${name}\\(\\) \\{[\\s\\S]*?\\n  \\}`));
  assert.ok(match, `${name} method missing`);
  return match[0];
}

test("logout posts to the gateway and only signs out on success", () => {
  const logout = method(read("lib/stores/auth.svelte.js"), "logout");
  assert.match(logout, /aigatewayPath\("\/admin\/auth\/logout"\)/);
  assert.match(logout, /method: "POST"/);
  assert.match(logout, /if \(!response\.ok\) return false;/);
  assert.ok(
    logout.indexOf("return false") < logout.indexOf("this.needsAuth = true"),
    "needsAuth must flip only after a successful logout",
  );
  assert.match(logout, /this\.passwordSession = false;/);
  assert.match(logout, /this\.generation\+\+;/);
});

test("login and the boot refresh track the password session", () => {
  const store = read("lib/stores/auth.svelte.js");
  assert.match(method(store, "login"), /this\.passwordSession = true;/);
  assert.match(method(store, "checkSession"), /aigatewayPath\("\/admin\/auth\/session"\)/);
  assert.match(read("App.svelte"), /void auth\.refreshTick;\n\s+auth\.checkSession\(\);/);
});

test("sidebar shows sign out only during a password session", () => {
  const sidebar = read("lib/components/organisms/Sidebar.svelte");
  assert.match(sidebar, /\{#if auth\.passwordSession && !auth\.needsAuth\}/);
  assert.match(sidebar, /onclick=\{signOut\}/);
  assert.match(sidebar, /flash\.error\(m\.sidebar_sign_out_failed\(\)\)/);
});
