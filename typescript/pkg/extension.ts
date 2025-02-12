import type { Action } from "./action.ts";

export type Extension = {
  name: string;
} & Manifest;

export type Manifest = {
  title: string;
  description?: string;
  commands?: readonly Command[];
  actions?: Action[];
};

export type Command = {
  name: string;
  description?: string;
  params?: readonly ParamDef[];
  mode: "filter" | "search" | "detail" | "silent" | "action";
};

export type ParamDef = {
  name: string;
  type: "string" | "number" | "boolean";
  description?: string;
  optional?: boolean;
};
