#!/usr/bin/env -S deno run -A

import * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.14";
import * as base64 from "https://deno.land/std@0.202.0/encoding/base64.ts";
import { toJson } from "jsr:@std/streams";

const manifest = {
  title: "GitHub",
  description: "Search GitHub repositories",
  root: [
    {
      title: "Search Repositories",
      type: "run",
      command: "search-repos",
    },
    {
      title: "Open Sunbeam Repo",
      type: "open",
      target: "https://github.com/pomdtr/sunbeam",
    },
  ],
  commands: [
    {
      description: "Search Repositories",
      name: "search-repos",
      mode: "search",
    },
    {
      description: "List Issues",
      name: "list-issues",
      mode: "filter",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
    {
      description: "List Pull Requests",
      name: "list-prs",
      mode: "filter",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
    {
      description: "View Readme",
      name: "view-readme",
      mode: "detail",
      params: [
        {
          name: "repo",
          description: "Repository Name",
          type: "string",
        },
      ],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest, null, 2));
  Deno.exit(0);
}

const token = Deno.env.get("GITHUB_TOKEN");
if (!token) {
  console.error("No GitHub token set");
  Deno.exit(1);
}

const command = Deno.args[0];
const params = await toJson(Deno.stdin.readable) as sunbeam.Params;

if (command == "search-repos") {
  if (!params.query) {
    const list: sunbeam.List = {
      emptyText: "Enter a query to search for repositories",
      items: [],
    };
    console.log(JSON.stringify(list, null, 2));
    Deno.exit(0);
  }

  const resp = await fetch(
    `https://api.github.com/search/repositories?q=${
      encodeURIComponent(
        params.query,
      )
    }`,
    {
      headers: {
        Authorization: `token ${token}`,
      },
    },
  );

  if (!resp.ok) {
    throw new Error(
      `Failed to search repositories: ${resp.status} ${resp.statusText}`,
    );
  }

  const data = await resp.json();
  const list: sunbeam.List = {
    items: data.items.map(
      (item: any) => ({
        title: item.full_name,
        accessories: [`${item.stargazers_count} *`],
        actions: [
          {
            title: "View README",
            type: "run",
            command: "view-readme",
            params: {
              repo: item.full_name,
            },
          },
          {
            title: "Open In Browser",
            type: "open",
            target: item.html_url,
          },
          {
            title: "List Issues",
            type: "run",
            command: "list-issues",
            params: {
              repo: item.full_name,
            },
          },
          {
            title: "List Pull Requests",
            type: "run",
            command: "list-prs",
            params: {
              repo: item.full_name,
            },
          },
          {
            title: "Copy URL",
            type: "copy",
            text: item.html_url,
          },
        ],
      } as sunbeam.ListItem),
    ),
  };

  console.log(JSON.stringify(list, null, 2));
} else if (command == "list-issues") {
  const resp = await fetch(
    `https://api.github.com/repos/${params.repo}/issues`,
    {
      headers: {
        Authorization: `token ${token}`,
      },
    },
  );

  if (!resp.ok) {
    throw new Error(
      `Failed to list issues: ${resp.status} ${resp.statusText}`,
    );
  }

  const data = await resp.json();
  const list: sunbeam.List = {
    items: data.map(
      (item: any) => ({
        title: item.title,
        accessories: [`#${item.number}`],
        actions: [
          {
            title: "Open In Browser",
            type: "open",
            target: item.html_url,
          },
          {
            title: "Copy URL",
            type: "copy",
            text: item.html_url,
          },
        ],
      } as sunbeam.ListItem),
    ),
  };

  console.log(JSON.stringify(list, null, 2));
} else if (command == "list-prs") {
  const resp = await fetch(
    `https://api.github.com/repos/${params.repo}/pulls`,
    {
      headers: {
        Authorization: `token ${token}`,
      },
    },
  );

  if (!resp.ok) {
    throw new Error(
      `Failed to list pull requests: ${resp.status} ${resp.statusText}`,
    );
  }

  const data = await resp.json();
  const list: sunbeam.List = {
    items: data.map(
      (item: any) => ({
        title: item.title,
        accessories: [`#${item.number}`],
        actions: [
          {
            title: "Open In Browser",
            type: "open",
            target: item.html_url,
          },
          {
            title: "Copy URL",
            key: "c",
            type: "copy",
            text: item.html_url,
          },
        ],
      } as sunbeam.ListItem),
    ),
  };

  console.log(JSON.stringify(list, null, 2));
} else if (command == "view-readme") {
  const resp = await fetch(
    `https://api.github.com/repos/${params.repo}/readme`,
    {
      headers: {
        Authorization: `token ${token}`,
      },
    },
  );

  if (!resp.ok) {
    throw new Error(
      `Failed to view readme: ${resp.status} ${resp.statusText}`,
    );
  }

  const data = await resp.json();
  const markdown = new TextDecoder().decode(base64.decode(data.content));

  const detail: sunbeam.Detail = {
    markdown,
    actions: [
      {
        title: "Open in Browser",
        type: "open",
        target: data.html_url,
      },
      {
        title: "List Issues",
        type: "run",
        command: "list-issues",
        params: {
          repo: params.repo,
        },
      },
      {
        title: "List Pull Requests",
        type: "run",
        command: "list-prs",
        params: {
          repo: params.repo,
        },
      },
    ],
  };

  console.log(JSON.stringify(detail, null, 2));
}
