import adapter from '@sveltejs/adapter-node';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		paths: { base: '/studio' },
		adapter: adapter()
	}
}

export default config;
