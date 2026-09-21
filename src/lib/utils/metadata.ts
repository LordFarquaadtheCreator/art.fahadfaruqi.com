import { formatExif, type Exif, type RawExif } from './exif';
import { derivativeUrl } from './variants';

const METADATA_API =
	import.meta.env.VITE_METADATA_API ?? 'https://assets.fahadfaruqi.com/api/metadata';

/** The shape the metadata API returns: generated fields plus the object's custom metadata. */
interface ApiObject extends RawExif {
	url: string;
	key: string;
	size: number;
	uploaded: string;
	etag?: string;
	title?: string;
	alttext?: string;
	description?: string;
	set?: string;
	number?: string;
}

interface MetadataResponse {
	count: number;
	objects: ApiObject[];
}

/** A photo as the gallery uses it: curated fields resolved, EXIF formatted, variants addressed. */
export interface Photo {
	key: string;
	original: string;
	display: string;
	grid: string;
	lqip: string;
	size: number;
	uploaded: string;
	set: string;
	number: number;
	title: string;
	alt: string;
	description: string;
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

export function toPhoto(object: ApiObject): Photo {
	const title = object.title?.trim() || humanize(object.key);
	const number = Number.parseInt(object.number ?? '', 10);

	return {
		key: object.key,
		original: object.url,
		display: derivativeUrl('display', object.key),
		grid: derivativeUrl('grid', object.key),
		lqip: derivativeUrl('lqip', object.key),
		size: object.size,
		uploaded: object.uploaded,
		set: object.set?.trim() || 'unfiled',
		number: Number.isFinite(number) ? number : 0,
		title,
		alt: object.alttext?.trim() || title,
		description: object.description?.trim() || '',
		exif: formatExif(object)
	};
}

export async function fetchPhotos(signal?: AbortSignal): Promise<Photo[]> {
	const response = await fetch(METADATA_API, { signal });
	if (!response.ok) throw new Error(`metadata API returned HTTP ${response.status}`);

	const { objects } = (await response.json()) as MetadataResponse;

	return objects.map(toPhoto);
}
