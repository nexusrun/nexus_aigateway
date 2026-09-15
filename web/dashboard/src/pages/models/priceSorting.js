// Sort numeric effective prices before rendering batches. Grouping preserves
// this order within each provider; unknown prices stay last in either direction.
export function sortModelsByPrice(rows, sort, pricingForRow) {
  const fields = { input: "input_per_mtok", output: "output_per_mtok" };
  const [kind, direction] = sort.split("_");
  const field = fields[kind];
  if (!field || !["asc", "desc"].includes(direction)) return rows;

  return rows.map((row) => {
    const value = pricingForRow(row)?.[field];
    return { row, price: typeof value === "number" && Number.isFinite(value) ? value : null };
  }).sort((a, b) => {
    if (a.price === null) return b.price === null ? 0 : 1;
    if (b.price === null) return -1;
    return (a.price - b.price) * (direction === "asc" ? 1 : -1);
  }).map(({ row }) => row);
}
