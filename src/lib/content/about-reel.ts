// The About page's content: The Turn, one entry per beat. Each entry names the section
// component that renders it — a key in components/reel/sections — and may carry its own
// params, numbers only that component interprets. Images carry their path and shape;
// notes and kept slots carry their words.
//
// Validated at module load, so a malformed entry fails the build (the routes are
// prerendered) instead of rendering a half-empty turn at runtime.

import type { Progression, Reel, Section } from '$lib/components/reel/types';
import { mediaUrl } from '$lib/utils/media';
import raw from './about-reel.json';

/** Missing phase values fall back to these; all three are viewport heights. */
const DEFAULT_PROGRESSION: Progression = {
	entry: 0.45,
	hold: 0.45,
	exit: 0.4
};

function fail(path: string, message: string): never {
	throw new Error(`about-reel.json — ${path} ${message}`);
}

function number(value: unknown, path: string, fallback: number): number {
	if (typeof value === 'number' && Number.isFinite(value)) return value;
	if (value === undefined) return fallback;
	fail(path, `is "${String(value)}" — expected a number`);
}

function params(input: unknown, path: string): Record<string, number> {
	if (input === undefined) return {};
	if (typeof input !== 'object' || input === null || Array.isArray(input)) {
		fail(path, 'must be an object of numbers');
	}

	const out: Record<string, number> = {};

	for (const [key, value] of Object.entries(input as Record<string, unknown>)) {
		if (typeof value !== 'number' || !Number.isFinite(value)) {
			fail(`${path}.${key}`, 'must be a number');
		}
		out[key] = value;
	}

	return out;
}

function section(input: unknown, path: string, index: number): Section {
	if (typeof input !== 'object' || input === null) fail(path, 'must be an object');

	const raw = input as Record<string, unknown>;
	const id = typeof raw.id === 'string' ? raw.id : fail(`${path}.id`, 'is required');
	const component =
		typeof raw.component === 'string'
			? raw.component
			: fail(`${path}.component`, 'is required — a key in components/reel/sections');
	const beatParams = params(raw.params, `${path}.params`);
	const kind = raw.kind;

	if (kind === 'image') {
		const src = typeof raw.src === 'string' ? mediaUrl(raw.src) : null;
		if (!src) fail(`${path}.src`, 'is required for an image section');

		const title = typeof raw.title === 'string' ? raw.title : '';

		return {
			id,
			component,
			params: beatParams,
			index,
			kind: 'image',
			src,
			alt: typeof raw.alt === 'string' ? raw.alt : title,
			title,
			width: number(raw.width, `${path}.width`, 0),
			height: number(raw.height, `${path}.height`, 0)
		};
	}

	if (kind === 'note' || kind === 'kept') {
		const text = typeof raw.text === 'string' ? raw.text : '';
		if (!text) fail(`${path}.text`, 'is required for a text section');

		return {
			id,
			component,
			params: beatParams,
			index,
			kind,
			label: typeof raw.label === 'string' ? raw.label : 'Note',
			text
		};
	}

	fail(`${path}.kind`, `is "${String(kind)}" — expected image, note or kept`);
}

function load(): Reel {
	const entries: unknown[] = raw.sections;
	if (!Array.isArray(entries) || entries.length === 0) {
		throw new Error('about-reel.json — "sections" must be a non-empty array');
	}

	const progressionRaw =
		typeof raw.progression === 'object' && raw.progression !== null
			? (raw.progression as Record<string, unknown>)
			: {};

	return {
		progression: {
			entry: number(progressionRaw.entry, 'progression.entry', DEFAULT_PROGRESSION.entry),
			hold: number(progressionRaw.hold, 'progression.hold', DEFAULT_PROGRESSION.hold),
			exit: number(progressionRaw.exit, 'progression.exit', DEFAULT_PROGRESSION.exit)
		},
		params: params(raw.params, 'params'),
		sections: entries.map((entry, i) => section(entry, `sections[${i}]`, i + 1))
	};
}

export const turn = load();
