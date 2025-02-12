#!/usr/bin/env -S deno run -A
import { toJson } from "jsr:@std/streams";
import * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.15";

const manifest = {
  title: "Val Town",
  description: "Search and view Val Town vals",
  commands: [
    {
      description: "List Vals",
      name: "list",
      mode: "filter",
      params: [{
        name: "user",
        optional: true,
        type: "string",
      }],
    },
    {
      description: "Search Vals",
      name: "search",
      mode: "search",
    },
    {
      description: "View Readme",
      name: "readme",
      mode: "detail",
      params: [{ name: "id", type: "string" }],
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));
  Deno.exit(0);
}

const token = Deno.env.get("VALTOWN_TOKEN");
if (!token) {
  console.error("No Val Town token found, please set it in your config");
  Deno.exit(1);
}

async function run(command: string, params: sunbeam.Params) {
  const client = new ValTownClient(token!);
  if (command == "list") {
    const username = params.user;
    const { id: userID } = await client.fetchJSON(
      username ? `/v1/alias/${username}` : "/v1/me",
    );

    const vals = await client.paginate(`/v1/users/${userID}/vals`);
    const items = vals.map(valToListItem);

    const list: sunbeam.List = {
      showDetail: true,
      items,
    };

    console.log(JSON.stringify(list));
  } else if (command == "search") {
    const query = params.query;
    if (query) {
      const { data: vals } = await client.fetchJSON(
        `/v1/search/vals?query=${encodeURIComponent(query)}&limit=50`,
      );
      console.log(
        JSON.stringify({
          showDetail: true,
          items: vals.map(valToListItem),
          emptyText: "No results",
        }),
      );
    } else {
      console.log(JSON.stringify({ emptyText: "No query" }));
    }
  } else if (command == "readme") {
    const { readme } = await client.fetchJSON(`/v1/vals/${params.id}`);
    const detail: sunbeam.Detail = {
      markdown: readme || "No readme",
      actions: readme
        ? [{ type: "copy", title: "Copy Readme", text: readme }]
        : [],
    };
    console.log(JSON.stringify(detail));
  }
}

class ValTownClient {
  constructor(private token: string) {}

  _fetch(target: string, init?: RequestInit) {
    return fetch(url, {
      ...init,
      headers: {
        ...init?.headers,
        Authorization: `Bearer ${this.token}`,
      },
    });
  }

  fetch(endpoint: string, init?: RequestInit) {
    return this._fetch(`https://api.val.town${endpoint}`, init);
  }

  async fetchJSON(endpoint: string, init?: RequestInit) {
    const resp = await this._fetch(`https://api.val.town${endpoint}`, init);
    if (!resp.ok) {
      throw new Error("Failed to fetch");
    }

    return resp.json();
  }

  async paginate(endpoint: string, init?: RequestInit) {
    const url = new URL(`https://api.val.town${endpoint}`);
    url.searchParams.set("limit", "100");

    let link: string = url.toString();
    const items: any = [];
    while (true) {
      const resp = await this._fetch(link, init);
      if (!resp.ok) {
        throw new Error(`Failed to fetch: ${resp.statusText}`);
      }

      const { data, links } = await resp.json();
      items.push(...data);
      if (!links.next) {
        break;
      }
      link = links.next;
    }

    return items;
  }
}

function valToListItem(val: any): sunbeam.ListItem {
  return {
    title: val.name,
    subtitle: val.author.username,
    detail: {
      markdown: "```tsx\n" + val.code + "\n```",
    },
    actions: [
      {
        title: "Open in Browser",
        type: "open",
        target: `https://val.town/v/${
          val.author.username.slice(1)
        }/${val.name}`,
      },
      {
        title: "Edit Val",
        type: "run",
        command: "edit",
        params: {
          id: val.id,
        },
        reload: true,
      },
      {
        title: "Open Web Endpoint",
        type: "open",
        target: `https://${
          val.author.username.slice(1)
        }-${val.name}.web.val.run`,
      },
      {
        title: "Copy URL",
        type: "copy",
        text: `https://val.town/v/${val.author.username.slice(1)}/${val.name}`,
      },
      {
        title: "Copy Web Endpoint",
        type: "copy",
        text: `https://${val.author.username.slice(1)}-${val.name}.web.val.run`,
      },
      {
        title: "Copy ID",
        type: "copy",
        text: val.id,
      },
      {
        title: "View Readme",
        type: "run",
        command: "readme",
        params: {
          id: val.id,
        },
      },
    ],
  };
}

try {
  await run(Deno.args[1], await toJson(Deno.stdin.readable) as sunbeam.Params);
} catch (e) {
  console.error(e);
  Deno.exit(1);
}
