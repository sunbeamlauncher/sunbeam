type ActionProps = {
  title: string;
};

export type CopyAction = {
  type: "copy";
  text: string;
} & ActionProps;

export type OpenAction = {
  type: "open";
  target: string;
} & ActionProps;

export type RunAction = {
  type: "run";
  command: string;
  extension?: string;
  params?: Params;
} & ActionProps;

export type Params = Record<string, string | number | boolean>;

export type Action =
  | CopyAction
  | OpenAction
  | RunAction;
