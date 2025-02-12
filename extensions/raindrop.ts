#!/usr/bin/env -S deno run -A
import type * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.15";

const manifest = {
  title: "Raindrop",
  description: "Manage your raindrop bookmarks",
  commands: [
    {
      description: "Search Bookmarks",
      name: "search-bookmarks",
      mode: "filter",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length === 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const raindropToken = Deno.env.get("RAINDROP_TOKEN");
if (!raindropToken) {
  console.error("No raindrop token found, please set it in your config");
  Deno.exit(1);
}

const command = Deno.args[0];

if (command == "search-bookmarks") {
  const resp = await fetch("https://api.raindrop.io/rest/v1/raindrops/0", {
    headers: {
      Authorization: `Bearer ${raindropToken}`,
    },
  });

  const { items: bookmarks } = (await resp.json()) as {
    items: {
      title: string;
      link: string;
      domain: string;
    }[];
  };

  const list: sunbeam.List = {
    items: bookmarks.map((bookmark) => ({
      title: bookmark.title,
      subtitle: bookmark.domain,
      actions: [
        {
          title: "Open URL",
          type: "open",
          target: bookmark.link,
        },
        {
          title: "Copy URL",
          type: "copy",
          text: bookmark.link,
        },
      ],
    })),
  };

  console.log(JSON.stringify(list));
} else {
  console.error(`Unknown command: ${command}`);
  Deno.exit(1);
}
