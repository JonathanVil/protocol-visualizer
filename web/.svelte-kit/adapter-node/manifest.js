export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.svg","icon_func.png"]),
	mimeTypes: {".svg":"image/svg+xml",".png":"image/png"},
	_: {
		client: {start:"_app/immutable/entry/start.DTc7ZoVu.js",app:"_app/immutable/entry/app.BLd2Dm1z.js",imports:["_app/immutable/entry/start.DTc7ZoVu.js","_app/immutable/chunks/d9qix_EJ.js","_app/immutable/chunks/BDApZGyf.js","_app/immutable/chunks/AIA2V70g.js","_app/immutable/entry/app.BLd2Dm1z.js","_app/immutable/chunks/DNUIruGW.js","_app/immutable/chunks/BDApZGyf.js","_app/immutable/chunks/Cv7XtZgV.js","_app/immutable/chunks/AIA2V70g.js","_app/immutable/chunks/Byb9vkdB.js","_app/immutable/chunks/zYhoFZmd.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
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

export const prerendered = new Set([]);

export const base = "";