// Derivative URLs. The derivatives are generated separately and uploaded to the
// bucket under a `d/` prefix that the metadata API filters out of its listing, so
// the naming convention below is the contract between the generator and this app.

const CDN_BASE = import.meta.env.VITE_CDN_BASE ?? 'https://assets.fahadfaruqi.com';

export type Variant = 'display' | 'grid' | 'lqip';

const DIR = { display: 'w2200', grid: 'w800', lqip: 'lqip' } as const;

export function derivativeKey(variant: Variant, key: string): string {
	return `d/${DIR[variant]}/${key.replace(/\.[^./]+$/, '')}.webp`;
}

export function derivativeUrl(variant: Variant, key: string): string {
	return `${CDN_BASE}/${derivativeKey(variant, key)}`;
}

export function originalUrl(key: string): string {
	return `${CDN_BASE}/${key}`;
}
