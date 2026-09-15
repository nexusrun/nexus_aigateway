import { test } from "node:test";
import assert from "node:assert/strict";
import { sortModelsByPrice } from "../src/pages/models/priceSorting.js";
import { groupDisplayModels } from "../src/pages/models/displayRows.js";

const rows = [
  { key: "expensive", provider_name: "a", input_per_mtok: 10, output_per_mtok: 1 },
  { key: "unknown", provider_name: "a" },
  { key: "cheap", provider_name: "a", input_per_mtok: 2, output_per_mtok: 20 },
  { key: "free", provider_name: "b", input_per_mtok: 0, output_per_mtok: 0 },
];

for (const [sort, expected] of [
  ["input_asc", ["free", "cheap", "expensive", "unknown"]],
  ["input_desc", ["expensive", "cheap", "free", "unknown"]],
  ["output_asc", ["free", "expensive", "cheap", "unknown"]],
  ["output_desc", ["cheap", "expensive", "free", "unknown"]],
]) {
  test(sort + " sorts numerically with free prices first/last and unknown prices last", () => {
    const original = [...rows];
    assert.deepEqual(sortModelsByPrice(rows, sort, (row) => row).map((row) => row.key), expected);
    assert.deepEqual(rows, original);
  });
}

test("default sorting preserves the original order", () => {
  assert.equal(sortModelsByPrice(rows, "", () => undefined), rows);
});

test("uses effective pricing and preserves equal-price order", () => {
  const sorted = sortModelsByPrice(rows, "input_asc", () => ({ input_per_mtok: 3 }));
  assert.deepEqual(sorted, rows);
  const overridden = sortModelsByPrice(rows, "input_asc", (row) =>
    row.key === "expensive" ? { input_per_mtok: 0.01 } : row,
  );
  assert.deepEqual(overridden.map((row) => row.key), ["free", "expensive", "cheap", "unknown"]);
});

test("provider grouping retains price order", () => {
  const sorted = sortModelsByPrice(rows, "input_asc", (row) => row);
  const groups = groupDisplayModels(sorted, [], []);
  assert.deepEqual(groups.map((group) => group.rows.map((row) => row.key)), [
    ["cheap", "expensive", "unknown"], ["free"],
  ]);
});

test("non-finite and missing prices sort last", () => {
  const values = [NaN, null, Infinity, undefined, 1];
  assert.deepEqual(sortModelsByPrice(values, "input_desc", (value) => ({ input_per_mtok: value })),
    [1, NaN, null, Infinity, undefined]);
});
