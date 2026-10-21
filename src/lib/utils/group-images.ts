import type { Photo } from './metadata';

export interface PhotoSet {
	name: string;
	slug: string;
	description: string;
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

			// Every photograph carries the set's description; the first one that has it
			// speaks for the set, so the text does not depend on the upload the API
			// happens to return first.
			const description = ordered.find((photo) => photo.setDescription)?.setDescription ?? '';

			return {
				name,
				slug: slugify(name),
				description,
				photos: ordered,
				latestUpload: ordered.reduce(
					(latest, photo) => (photo.uploaded > latest ? photo.uploaded : latest),
					ordered[0].uploaded
				)
			};
		})
		.sort((a, b) => b.latestUpload.localeCompare(a.latestUpload));
}
