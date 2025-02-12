#!/usr/bin/env -S deno run -A
import { toJson } from "jsr:@std/streams";
import * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.14";

const manifest: sunbeam.Manifest = {
  title: "NPM Search",
  description: "Search NPM packages",
  root: [
    { title: "Search NPM Packages", type: "run", command: "search" },
  ],
  commands: [
    {
      name: "search",
      description: "Search NPM Packages",
      mode: "search",
      params: [
        {
          name: "query",
          description: "Search Query",
          type: "string",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const params = await toJson(Deno.stdin.readable) as { query: string };
const resp = await fetch(
  `https://registry.npmjs.com/-/v1/search?text=${
    encodeURIComponent(params.query)
  }`,
);
const { objects: packages } = await resp.json();
const items: sunbeam.ListItem[] = [];
for (const pkg of packages) {
  const item: sunbeam.ListItem = {
    title: pkg.package.name,
    subtitle: pkg.package.description || "",
    actions: [
      {
        type: "open",
        title: "Open Package",
        target: pkg.package.links.npm,
      },
      {
        type: "copy",
        title: "Open Package Name",
        text: pkg.package.name,
      },
    ],
  };

  items.push(item);
}
console.log(JSON.stringify({ items }));
