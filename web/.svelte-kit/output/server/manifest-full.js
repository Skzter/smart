export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["robots.txt"]),
	mimeTypes: {".txt":"text/plain"},
	_: {
		client: {start:"_app/immutable/entry/start.DtrQNxTD.js",app:"_app/immutable/entry/app.epUTbfoa.js",imports:["_app/immutable/entry/start.DtrQNxTD.js","_app/immutable/chunks/CsGReZ_D.js","_app/immutable/chunks/lj6ma2N6.js","_app/immutable/chunks/B32v5VJ4.js","_app/immutable/entry/app.epUTbfoa.js","_app/immutable/chunks/lj6ma2N6.js","_app/immutable/chunks/DmMW_KmW.js","_app/immutable/chunks/r308AAZP.js","_app/immutable/chunks/B32v5VJ4.js","_app/immutable/chunks/Bs4o_7-u.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
