// The section registry. One entry per section component — a section being one turn of The
// Turn. The JSON's `component` field names one of these keys. Adding a turn is two lines
// here, one file in this folder, and one entry in about-reel.json.

import type { Component } from 'svelte';
import type { Section } from '../types';
import Rise from './Rise.svelte';
import Develop from './Develop.svelte';
import Drift from './Drift.svelte';

export interface SectionProps {
	section: Section;
}

const SECTIONS: Record<string, Component<SectionProps>> = {
	rise: Rise,
	develop: Develop,
	drift: Drift
};

export function sectionComponent(name: string): Component<SectionProps> {
	const found = SECTIONS[name];

	if (!found) {
		throw new Error(
			`No section component named "${name}" — add it to src/lib/components/reel/sections/index.ts`
		);
	}

	return found;
}
