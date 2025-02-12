#!/usr/bin/env -S deno run -A
import type * as sunbeam from "jsr:@pomdtr/sunbeam@0.0.14";

const manifest = {
  title: "Tailscale",
  description: "Manage your tailscale devices",
  commands: [
    {
      name: "list-devices",
      description: "Search My Devices",
      mode: "filter",
    },
  ],
} as const satisfies sunbeam.Manifest;

if (Deno.args.length == 0) {
  console.log(JSON.stringify(manifest));

  Deno.exit(0);
}

type Device = {
  TailscaleIPs: string[];
  DNSName: string;
  OS: string;
  Online: boolean;
};

const command = Deno.args[0];
if (command == "list-devices") {
  const command = new Deno.Command("tailscale", { args: ["status", "--json"] });
  const { stdout } = await command.output();
  const status = JSON.parse(new TextDecoder().decode(stdout));
  const devices: Device[] = Object.values(status.Peer);
  const items: sunbeam.ListItem[] = devices.map((device) => ({
    title: device.DNSName.split(".")[0],
    subtitle: device.TailscaleIPs[0],
    accessories: [device.OS, device.Online ? "online" : "offline"],
    actions: [
      {
        title: "Copy SSH Command",
        type: "copy",
        text: `ssh ${device.TailscaleIPs[0]}`,
      },
      {
        title: "Copy IP",
        key: "i",
        type: "copy",
        text: device.TailscaleIPs[0],
      },
    ],
  }));

  const list: sunbeam.List = { items };

  console.log(JSON.stringify(list));
}
