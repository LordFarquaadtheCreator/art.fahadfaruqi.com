import { formatExif, type Exif, type RawExif } from './exif';
import { mediaUrl } from './media';

// Dev reads the listing through the dev server's proxy, so a fresh upload is in the
// next reload rather than in ten minutes. VITE_METADATA_API is read by builds.
const METADATA_API = import.meta.env.DEV
	? '/api/metadata'
	: (import.meta.env.VITE_METADATA_API ?? 'https://assets.fahadfaruqi.com/api/metadata');

/** The shape the metadata API returns: generated fields plus the object's custom metadata. */
interface ApiObject extends RawExif {
	url: string;
	key: string;
	size: number;
	uploaded: string;
	etag?: string;
	master_url?: string | null;
	master_size?: number | null;
	compressed_url?: string | null;
	compressed_size?: number | null;
	width?: number | null;
	height?: number | null;
	title?: string;
	alttext?: string;
	description?: string;
	set?: string;
	number?: string;
	zoom_x?: string;
	zoom_y?: string;
	zoom_level?: string;
	zoom_duration?: string;
}

interface MetadataResponse {
	count: number;
	objects: ApiObject[];
}

/** A programmed zoom: settle the framing on a point, at a scale, over a time. */
export interface ZoomSpec {
	/** The point to settle on, as fractions of the photograph (0–1 per axis). */
	x: number;
	y: number;
	/** The scale to settle at: 1 = no move, 2 = 2× in, 0.5 = 2× out. */
	level: number;
	/** Milliseconds one full traverse takes. */
	duration: number;
}

/** A photo as the gallery uses it: curated fields resolved, EXIF formatted. */
export interface Photo {
	key: string;
	compressed: string;
	master: string;
	width: number;
	height: number;
	size: number;
	uploaded: string;
	set: string;
	number: number;
	title: string;
	alt: string;
	description: string;
	zoom: ZoomSpec | null;
	exif: Exif;
}

/** "ceres-and-kimi-1.png" -> "Ceres and kimi 1" — only used for photos with no title. */
const humanize = (key: string) =>
	key
		.replace(/\.[^./]+$/, '')
		.replace(/[-_]+/g, ' ')
		.replace(/\s+/g, ' ')
		.trim()
		.replace(/^[a-z]/, (character) => character.toUpperCase());

/**
 * The four zoom keys arrive as strings like every other curated field, and they are
 * written as a set: any missing or malformed value means no programmed zoom, and a
 * level of 1 is the CLI's own "no move".
 */
function parseZoom(object: ApiObject): ZoomSpec | null {
	const x = Number.parseFloat(object.zoom_x ?? '');
	const y = Number.parseFloat(object.zoom_y ?? '');
	const level = Number.parseFloat(object.zoom_level ?? '');
	const duration = Number.parseFloat(object.zoom_duration ?? '');

	if (![x, y, level, duration].every(Number.isFinite)) return null;
	if (x < 0 || x > 1 || y < 0 || y > 1) return null;
	if (level <= 0 || level === 1) return null;
	if (duration <= 0) return null;

	return { x, y, level, duration };
}

export function toPhoto(object: ApiObject): Photo {
	const title = object.title?.trim() || humanize(object.key);
	const number = Number.parseInt(object.number ?? '', 10);
	const compressed = mediaUrl(object.compressed_url ?? object.url) ?? object.url;
	const master = mediaUrl(object.master_url ?? object.url) ?? object.url;

	return {
		key: object.key,
		compressed,
		master,
		width: object.width ?? 0,
		height: object.height ?? 0,
		size: object.master_size ?? object.size,
		uploaded: object.uploaded,
		set: object.set?.trim() || 'unfiled',
		number: Number.isFinite(number) ? number : 0,
		title,
		alt: object.alttext?.trim() || title,
		description: object.description?.trim() || '',
		zoom: parseZoom(object),
		exif: formatExif(object)
	};
}

export async function fetchPhotos(signal?: AbortSignal): Promise<Photo[]> {
	const response = await fetch(METADATA_API, { signal });
	if (!response.ok) throw new Error(`metadata API returned HTTP ${response.status}`);

	const { objects } = (await response.json()) as MetadataResponse;

	return objects.map(toPhoto);
}
