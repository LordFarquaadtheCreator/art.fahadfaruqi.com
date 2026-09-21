import type { Photo } from './metadata';

export interface PhotoSet {
	name: string;
	slug: string;
	photos: Photo[];
	latestUpload: string;
}

export const slugify = (name: string) =>
	name
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

/** Groups photos into sets ordered by most recent upload, photos ordered by their `number`. */
export function groupBySet(photos: Photo[]): PhotoSet[] {
	const grouped = new Map<string, Photo[]>();

	for (const photo of photos) {
		const bucket = grouped.get(photo.set);
		if (bucket) bucket.push(photo);
		else grouped.set(photo.set, [photo]);
	}

	return [...grouped.entries()]
		.map(([name, setPhotos]) => {
			const ordered = [...setPhotos].sort((a, b) => a.number - b.number);

			return {
				name,
				slug: slugify(name),
				photos: ordered,
				latestUpload: ordered.reduce(
					(latest, photo) => (photo.uploaded > latest ? photo.uploaded : latest),
					ordered[0].uploaded
				)
			};
		})
		.sort((a, b) => b.latestUpload.localeCompare(a.latestUpload));
}
