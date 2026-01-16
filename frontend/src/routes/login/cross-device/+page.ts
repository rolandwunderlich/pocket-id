import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const code = url.searchParams.get('code');
	const redirect = url.searchParams.get('redirect') || '/settings';

	return {
		code,
		redirect
	};
};
