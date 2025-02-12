#!/usr/bin/env -S deno run -A

import { toJson } from "jsr:@std/streams";
import type * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.14";

const manifest = {
  title: "Gists",
  description: "Manage your gists",
  commands: [
    {
      name: "manage",
      description: "Search Gists",
      mode: "filter",
    },
    {
      name: "browse",
      description: "Browse Gist Files",
      mode: "filter",
      params: [{ name: "id", description: "Gist ID", type: "string" }],
    },
    {
      name: "view",
      description: "View Gist File",
      mode: "detail",
      params: [
        { name: "id", description: "Gist ID", type: "string" },
        { name: "filename", description: "Filename", type: "string" },
      ],
    },
    {
      name: "delete",
      description: "Delete Gist",
      mode: "silent",
      params: [{ name: "id", description: "Gist ID", type: "string" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}
const githubToken = Deno.env.get("GITHUB_TOKEN");
if (!githubToken) {
  console.error("No github token set");
  Deno.exit(1);
}

try {
  const res = await run(
    Deno.args[0],
    await toJson(Deno.stdin.readable) as sunbeam.Params,
  );
  if (res) {
    console.log(JSON.stringify(res));
  }
} catch (e) {
  console.error(e);
  Deno.exit(1);
}

async function run(command: string, params: sunbeam.Params) {
  switch (command) {
    case "manage": {
      const resp = await fetchGithub("/gists");
      if (resp.status != 200) {
        throw new Error("Failed to fetch gists");
      }

      const gists = (await resp.json()) as any[];
      return {
        items: gists.map((gist) => ({
          title: Object.keys(gist.files)[0],
          subtitle: gist.description || "",
          accessories: [gist.public ? "Public" : "Private"],
          actions: [
            Object.keys(gist.files).length > 1
              ? {
                type: "run",
                title: "Browse Files",
                command: "browse",
                params: {
                  id: gist.id,
                },
              }
              : {
                type: "run",
                title: "View File",
                command: "view",
                params: {
                  id: gist.id,
                  filename: Object.keys(gist.files)[0],
                },
              },
            {
              type: "open",
              title: "Open in Browser",
              target: gist.html_url,
            },
            {
              type: "copy",
              title: "Copy URL",
              text: gist.html_url,
            },
            {
              title: "Create Gist",
              type: "run",
              command: "create",
            },
            {
              title: "Delete Gist",
              type: "run",
              command: "delete",
              params: {
                id: gist.id,
              },
              reload: true,
            },
          ],
        })),
      } as sunbeam.List;
    }
    case "browse": {
      const resp = await fetchGithub(`/gists/${params.id}`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch gist");
      }

      const gist = (await resp.json()) as any;
      return {
        items: Object.entries(gist.files).map(([filename]) => ({
          title: filename,
          actions: [
            {
              title: "View",
              type: "run",
              command: "view",
              params: {
                id: params.id,
                filename,
              },
            },
            {
              title: "Edit",
              type: "run",
              command: "edit",
              params: {
                id: params.id,
                filename,
              },
            },
          ],
        })),
      } as sunbeam.List;
    }
    case "view": {
      const { id, filename } = params as { id: string; filename: string };
      const resp = await fetchGithub(`/gists/${id}`);
      if (resp.status != 200) {
        throw new Error("Failed to fetch gist");
      }

      const gist = (await resp.json()) as any;
      const file = gist.files[filename];
      if (!file) {
        throw new Error("File not found");
      }
      const lang = file.language?.toLowerCase();

      return {
        markdown: lang == "md"
          ? file.content
          : `\`\`\`${lang || ""}\n${file.content}\n\`\`\``,
        actions: [
          {
            title: "Edit File",
            type: "run",
            command: "edit",
            params: {
              id,
              filename,
            },
            reload: true,
          },
          {
            title: "Copy Content",
            type: "copy",
            text: file.content,
          },
          {
            title: "Copy Raw URL",
            type: "copy",
            text: file.raw_url,
          },
          {
            title: "Open in Browser",
            type: "open",
            target: gist.html_url,
          },
        ],
      } as sunbeam.Detail;
    }
    case "delete": {
      const { id } = params as { id: string };
      const resp = await fetchGithub(`/gists/${id}`, {
        method: "DELETE",
      });
      if (resp.status != 204) {
        throw new Error("Failed to delete gist");
      }
    }
  }
}

function fetchGithub(endpoint: string, init?: RequestInit) {
  return fetch(`https://api.github.com${endpoint}`, {
    ...init,
    headers: {
      ...init?.headers,
      Authorization: `Bearer ${githubToken}`,
    },
  });
}
