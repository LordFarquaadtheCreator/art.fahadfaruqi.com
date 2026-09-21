// Formats the raw camera strings stored in R2 object metadata into display values.
// R2 lowercases metadata keys, so the source fields are `fnumber`, `focallength`,
// `exposuretime`, `isospeedratings`, `datetimeoriginal`, `make`, `model`, `lens` —
// each EXIF value arrives as its raw rational string ("7/2", "18/1", "1/60").

export interface Exif {
	camera: string | null;
	lens: string | null;
	focalLength: string | null;
	aperture: string | null;
	shutter: string | null;
	iso: string | null;
	date: string | null;
}

export interface RawExif {
	make?: string;
	model?: string;
	lens?: string;
	fnumber?: string;
	focallength?: string;
	exposuretime?: string;
	isospeedratings?: string;
	datetimeoriginal?: string;
}

const trim = (value?: string) => {
	const trimmed = value?.trim();
	return trimmed ? trimmed : undefined;
};

/** "7/2" -> 3.5, "18/1" -> 18. Returns null for anything not a rational. */
function rational(value?: string): number | null {
	const trimmed = trim(value);
	if (!trimmed) return null;

	const [numerator, denominator] = trimmed.split('/').map(Number);
	if (!Number.isFinite(numerator)) return null;
	if (denominator === undefined) return numerator;
	if (!Number.isFinite(denominator) || denominator === 0) return null;

	return numerator / denominator;
}

const decimal = (value: number) => Number(value.toFixed(1)).toString();

/** "NIKON CORPORATION" + "NIKON D3300" -> "Nikon D3300" (no duplicated brand). */
function cameraName(make?: string, model?: string): string | null {
	const model_ = trim(model);
	const make_ = trim(make);
	if (!model_) return make_ ? titleCase(make_) : null;

	const brand = make_?.split(/\s+/)[0].toLowerCase();
	const name = brand && model_.toLowerCase().startsWith(brand) ? model_ : [make_, model_].join(' ');

	return titleCase(name.replace(/\s+(corporation|inc\.?|co\.?)$/i, ''));
}

function titleCase(value: string): string {
	return value
		.toLowerCase()
		.replace(/\b[a-z]/g, (character) => character.toUpperCase())
		.replace(/\bDslr\b/g, 'DSLR');
}

function shutterSpeed(value?: string): string | null {
	const seconds = rational(value);
	if (seconds === null) return null;

	if (seconds >= 1) return `${decimal(seconds)}s`;
	return `1/${Math.round(1 / seconds)}s`;
}

/** "2026-09-19T20:56:31.00-05:00" -> "2026-09-19" */
function isoDate(value?: string): string | null {
	const trimmed = trim(value);
	if (!trimmed) return null;

	const match = trimmed.match(/^(\d{4})-(\d{2})-(\d{2})/);
	return match ? `${match[1]}-${match[2]}-${match[3]}` : null;
}

export function formatExif(raw: RawExif): Exif {
	const focalLength = rational(raw.focallength);
	const aperture = rational(raw.fnumber);

	return {
		camera: cameraName(raw.make, raw.model),
		lens: trim(raw.lens) ?? null,
		focalLength: focalLength === null ? null : `${decimal(focalLength)}mm`,
		aperture: aperture === null ? null : `f/${decimal(aperture)}`,
		shutter: shutterSpeed(raw.exposuretime),
		iso: trim(raw.isospeedratings) ?? null,
		date: isoDate(raw.datetimeoriginal)
	};
}

/** Compact one-line summary for grid captions: "18mm · f/3.5 · 1/60s · ISO 400" */
export function exifLine(exif: Exif): string {
	return [exif.focalLength, exif.aperture, exif.shutter, exif.iso && `ISO ${exif.iso}`]
		.filter(Boolean)
		.join(' · ');
}

export const exifRows = (exif: Exif): { label: string; value: string }[] =>
	(
		[
			['Camera', exif.camera],
			['Lens', exif.lens],
			['Focal length', exif.focalLength],
			['Aperture', exif.aperture],
			['Shutter', exif.shutter],
			['ISO', exif.iso],
			['Captured', exif.date]
		] as const
	)
		.filter(([, value]) => Boolean(value))
		.map(([label, value]) => ({ label, value: value as string }));
