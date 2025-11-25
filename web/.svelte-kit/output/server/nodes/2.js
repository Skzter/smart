

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_page.svelte.js')).default;
export const imports = ["_app/immutable/nodes/2.CD75qVGM.js","_app/immutable/chunks/r308AAZP.js","_app/immutable/chunks/lj6ma2N6.js","_app/immutable/chunks/yIRrVp8Q.js"];
export const stylesheets = [];
export const fonts = [];
