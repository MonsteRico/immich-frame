import { cp, rm } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

const copies = [
  ["ui/frame/dist", "internal/api/static/frame"],
  ["ui/setup/dist", "internal/api/static/setup"]
];

for (const [from, to] of copies) {
  const source = resolve(root, from);
  const target = resolve(root, to);
  await rm(target, { recursive: true, force: true });
  await cp(source, target, { recursive: true });
  console.log(`synced ${from} -> ${to}`);
}
