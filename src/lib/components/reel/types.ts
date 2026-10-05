/** How the turn occupies scroll: the arrival, the rest once complete, the hand-off after. */
export interface Progression {
	/** viewport heights spent arriving, the crank parked at 0 */
	entry: number;
	/** viewport heights of rest once the turn is complete */
	hold: number;
	/** viewport heights spent leaving */
	exit: number;
}

interface SectionBase {
	id: string;
	/** which section component renders this beat — a key in components/reel/sections */
	component: string;
	/** every number this beat's own animation needs; only its component interprets them */
	params: Record<string, number>;
	/** its position in the turn, 1-based — the counter reads it */
	index: number;
}

export interface ImageSection extends SectionBase {
	kind: 'image';
	src: string;
	alt: string;
	title: string;
	width: number;
	height: number;
}

export interface NoteSection extends SectionBase {
	kind: 'note';
	label: string;
	text: string;
}

export interface KeptSection extends SectionBase {
	kind: 'kept';
	label: string;
	text: string;
}

/** One turn of The Turn — a single beat: an image, a note, or an archive slot. */
export type Section = ImageSection | NoteSection | KeptSection;

export interface Reel {
	progression: Progression;
	/** The Turn's own knobs */
	params: Record<string, number>;
	sections: Section[];
}
