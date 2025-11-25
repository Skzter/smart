// fixes import error because moo is written in js
declare module "moo" {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const moo: any;
    export = moo;
}
