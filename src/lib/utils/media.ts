// Dev reads CDN images through the dev server, so nothing is served from cache there.

const CDN_BASE = 'https://assets.fahadfaruqi.com';

export function mediaUrl(url: string | null | undefined): string | null {
	if (!url) return null;
	if (!import.meta.env.DEV) return url;
	return url.startsWith(CDN_BASE) ? `/cdn${url.slice(CDN_BASE.length)}` : url;
}
